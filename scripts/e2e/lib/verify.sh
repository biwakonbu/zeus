#!/bin/bash
# verify.sh - E2E テスト用構造比較ライブラリ
# jq を使った座標除外版の構造比較を提供

set -euo pipefail

# common.sh を読み込み
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

# =============================================================================
# ノード抽出（座標除外）
# =============================================================================

# 実際の状態からノードを抽出（name, status, progress のみ、名前順ソート）
extract_nodes() {
    local json="$1"
    local result
    local jq_error

    # jq エラーをキャプチャ
    jq_error=$(echo "$json" | jq -S '[.nodes[] | {name, status, progress}] | sort_by(.name)' 2>&1) || {
        log_error "ノード抽出に失敗しました"
        log_info "jq エラー: $jq_error"
        return 1
    }

    result="$jq_error"

    if [[ -z "$result" || "$result" == "null" ]]; then
        log_error "ノード抽出結果が空またはnullです"
        log_info "入力JSON の最初の 100 文字: ${json:0:100}"
        return 1
    fi

    echo "$result"
}

# ゴールデンファイルからノードを抽出
extract_golden_nodes() {
    local golden_file="$1"
    jq -S '.expected.nodes | sort_by(.name)' "$golden_file"
}

export -f extract_nodes extract_golden_nodes

# =============================================================================
# エッジ抽出（ID→名前変換）
# =============================================================================

# 実際の状態からエッジを抽出（ID を名前に変換）
extract_edges() {
    local json="$1"
    local result
    local jq_error

    # ID→名前変換でエラーをチェック（null 値が含まれないことを確認）
    jq_error=$(echo "$json" | jq -S '
        .nodes as $nodes |
        [.edges[] as $e | {
            from: ([$nodes[] | select(.id == $e.from)][0].name // empty),
            to: ([$nodes[] | select(.id == $e.to)][0].name // empty)
        } | select(.from and .to)] | sort_by(.from, .to)
    ' 2>&1) || {
        log_error "エッジ抽出でエラーが発生しました"
        log_info "jq エラー: $jq_error"
        echo "[]"
        return 1
    }

    result="$jq_error"

    # jq の出力がエラーまたは空になっていないかチェック
    if [[ -z "$result" || "$result" == "null" ]]; then
        log_error "エッジ抽出でエラーが発生しました（null 値またはパース失敗）"
        log_info "取得したノード数: $(echo "$json" | jq '.nodes | length' 2>/dev/null)"
        log_info "取得したエッジ数: $(echo "$json" | jq '.edges | length' 2>/dev/null)"
        echo "[]"
        return 1
    fi

    echo "$result"
}

# ゴールデンファイルからエッジを抽出
extract_golden_edges() {
    local golden_file="$1"
    jq -S '.expected.edges | sort_by(.from, .to)' "$golden_file"
}

export -f extract_edges extract_golden_edges

# =============================================================================
# カウント抽出
# =============================================================================

# 実際の状態からカウントを抽出
extract_counts() {
    local json="$1"
    local node_count edge_count

    # jq エラーハンドリング
    node_count=$(echo "$json" | jq '.nodes | length' 2>/dev/null) || {
        log_error "ノードのカウント抽出に失敗しました"
        return 1
    }
    edge_count=$(echo "$json" | jq '.edges | length' 2>/dev/null) || {
        log_error "エッジのカウント抽出に失敗しました"
        return 1
    }

    # null チェック
    if [[ -z "$node_count" || -z "$edge_count" ]]; then
        log_error "カウント値が空です: nodeCount=$node_count, edgeCount=$edge_count"
        return 1
    fi

    echo "{\"nodeCount\": ${node_count}, \"edgeCount\": ${edge_count}}"
}

# ゴールデンファイルからカウントを抽出
extract_golden_counts() {
    local golden_file="$1"
    jq '.expected.counts' "$golden_file"
}

export -f extract_counts extract_golden_counts

# =============================================================================
# 比較関数
# =============================================================================

# ノード比較
compare_nodes() {
    local actual_json="$1"
    local golden_file="$2"

    local actual_nodes expected_nodes

    actual_nodes=$(extract_nodes "$actual_json")
    expected_nodes=$(extract_golden_nodes "$golden_file")

    if [[ "$actual_nodes" == "$expected_nodes" ]]; then
        log_success "ノード比較: 一致"
        return 0
    else
        log_error "ノード比較: 不一致"
        echo "期待値:"
        echo "$expected_nodes" | jq .
        echo "実際値:"
        echo "$actual_nodes" | jq .
        echo "差分:"
        diff -u <(echo "$expected_nodes" | jq -S .) <(echo "$actual_nodes" | jq -S .) || true
        return 1
    fi
}

# エッジ比較
compare_edges() {
    local actual_json="$1"
    local golden_file="$2"

    local actual_edges expected_edges

    actual_edges=$(extract_edges "$actual_json")
    expected_edges=$(extract_golden_edges "$golden_file")

    if [[ "$actual_edges" == "$expected_edges" ]]; then
        log_success "エッジ比較: 一致"
        return 0
    else
        log_error "エッジ比較: 不一致"
        echo "期待値:"
        echo "$expected_edges" | jq .
        echo "実際値:"
        echo "$actual_edges" | jq .
        echo "差分:"
        diff -u <(echo "$expected_edges" | jq -S .) <(echo "$actual_edges" | jq -S .) || true
        return 1
    fi
}

# カウント比較
compare_counts() {
    local actual_json="$1"
    local golden_file="$2"

    local actual_counts expected_counts

    actual_counts=$(extract_counts "$actual_json")
    expected_counts=$(extract_golden_counts "$golden_file")

    local actual_node_count expected_node_count actual_edge_count expected_edge_count

    actual_node_count=$(echo "$actual_counts" | jq '.nodeCount')
    expected_node_count=$(echo "$expected_counts" | jq '.nodeCount')
    actual_edge_count=$(echo "$actual_counts" | jq '.edgeCount')
    expected_edge_count=$(echo "$expected_counts" | jq '.edgeCount')

    local passed=true

    if [[ "$actual_node_count" == "$expected_node_count" ]]; then
        log_success "ノード数: ${actual_node_count} == ${expected_node_count}"
    else
        log_error "ノード数: ${actual_node_count} != ${expected_node_count}"
        passed=false
    fi

    if [[ "$actual_edge_count" == "$expected_edge_count" ]]; then
        log_success "エッジ数: ${actual_edge_count} == ${expected_edge_count}"
    else
        log_error "エッジ数: ${actual_edge_count} != ${expected_edge_count}"
        passed=false
    fi

    if $passed; then
        return 0
    else
        return 1
    fi
}

export -f compare_nodes compare_edges compare_counts

# =============================================================================
# ゴールデンファイル検証
# =============================================================================

# ゴールデンファイルのフォーマット検証
validate_golden_file() {
    local golden_file="$1"

    # ファイル存在確認
    if [[ ! -f "$golden_file" ]]; then
        log_error "ゴールデンファイルが見つかりません: $golden_file"
        return 1
    fi

    # JSON フォーマット確認
    if ! jq empty "$golden_file" 2>/dev/null; then
        log_error "ゴールデンファイルの JSON フォーマットが無効です: $golden_file"
        return 1
    fi

    # 必須フィールド確認
    local required_fields=("metadata" "expected")
    for field in "${required_fields[@]}"; do
        if ! jq -e ".$field" "$golden_file" >/dev/null 2>&1; then
            log_error "ゴールデンファイルに必須フィールドがありません: .$field"
            return 1
        fi
    done

    # expected の必須サブフィールド確認
    local expected_fields=("nodes" "edges" "counts")
    for field in "${expected_fields[@]}"; do
        if ! jq -e ".expected.$field" "$golden_file" >/dev/null 2>&1; then
            log_error "ゴールデンファイルの expected に必須フィールドがありません: .expected.$field"
            return 1
        fi
    done

    log_info "ゴールデンファイル: 妥当"
}

export -f validate_golden_file

# =============================================================================
# メイン検証関数
# =============================================================================

# 構造比較を実行
verify_state() {
    local actual_json="$1"
    local golden_file="$2"

    log_step "構造比較を実行"

    require_file "$golden_file"

    # ゴールデンファイルの妥当性チェック
    if ! validate_golden_file "$golden_file"; then
        return 1
    fi

    local failed=false

    # カウント比較
    if ! compare_counts "$actual_json" "$golden_file"; then
        failed=true
    fi

    # ノード比較
    if ! compare_nodes "$actual_json" "$golden_file"; then
        failed=true
    fi

    # エッジ比較
    if ! compare_edges "$actual_json" "$golden_file"; then
        failed=true
    fi

    if $failed; then
        log_error "構造比較: 失敗"
        return 1
    else
        log_success "構造比較: 成功"
        return 0
    fi
}

# 実際の状態をゴールデン形式に変換
convert_to_golden_format() {
    local actual_json="$1"
    local test_id="${2:-basic-tasks-001}"
    local description="${3:-Basic project with 3 tasks forming a chain}"

    local nodes edges counts

    nodes=$(extract_nodes "$actual_json")
    edges=$(extract_edges "$actual_json")
    counts=$(extract_counts "$actual_json")

    jq -n \
        --arg test_id "$test_id" \
        --arg description "$description" \
        --arg date "$(date +%Y-%m-%d)" \
        --argjson nodes "$nodes" \
        --argjson edges "$edges" \
        --argjson counts "$counts" \
        '{
            metadata: {
                test_id: $test_id,
                schema_version: "1.0",
                zeus_version: ">=0.1.0",
                created: $date,
                description: $description
            },
            comparison: {
                mode: "structural",
                ignore_fields: ["nodes[*].x", "nodes[*].y", "nodes[*].id", "viewport"],
                edge_mode: "by_name"
            },
            expected: {
                nodes: $nodes,
                edges: $edges,
                counts: $counts
            }
        }'
}

export -f verify_state convert_to_golden_format
