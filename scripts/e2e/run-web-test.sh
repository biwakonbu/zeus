#!/bin/bash
# run-web-test.sh - Zeus E2E テストメインスクリプト
# agent-browser を使用して Web UI の状態を検証

set -euo pipefail

# =============================================================================
# 初期化
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# ライブラリ読み込み
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"
# shellcheck source=lib/verify.sh
source "${SCRIPT_DIR}/lib/verify.sh"

# =============================================================================
# グローバル変数
# =============================================================================

TEST_DIR=""
SERVER_PID=""
SESSION=""
EXIT_CODE=0

# =============================================================================
# クリーンアップ
# =============================================================================

cleanup() {
    local exit_code=$?
    log_step "クリーンアップ実行"

    # ブラウザセッション終了
    if [[ -n "$SESSION" ]]; then
        log_info "ブラウザセッション終了: $SESSION"
        agent-browser --session "$SESSION" close 2>/dev/null || true
    fi

    # サーバー終了
    if [[ -n "$SERVER_PID" ]]; then
        log_info "サーバー終了: PID $SERVER_PID"
        safe_kill "$SERVER_PID"
    fi

    # テストディレクトリ削除（成功時かつ KEEP_ARTIFACTS=false の場合）
    if [[ -n "$TEST_DIR" && -d "$TEST_DIR" ]]; then
        if [[ $exit_code -eq 0 && "$KEEP_ARTIFACTS" != "true" ]]; then
            log_info "テストディレクトリ削除: $TEST_DIR"
            rm -rf "$TEST_DIR"
        else
            log_warn "テストディレクトリ保持: $TEST_DIR"
        fi
    fi

    exit $exit_code
}

collect_artifacts() {
    log_step "アーティファクト収集"
    init_artifacts_dir

    # 現在の状態をJSON として保存（取得済みの場合）
    if [[ -n "${ACTUAL_STATE:-}" ]]; then
        save_artifact "actual-state.json" "$ACTUAL_STATE"
    fi

    # サーバーログがあれば保存
    if [[ -f "${TEST_DIR:-/tmp}/server.log" ]]; then
        copy_artifact "${TEST_DIR}/server.log" "server.log"
    fi

    # テストプロジェクトの状態
    if [[ -d "${TEST_DIR:-/tmp}/.zeus" ]]; then
        tar -czf "${ARTIFACTS_DIR}/zeus-data.tar.gz" -C "$TEST_DIR" .zeus 2>/dev/null || true
    fi

    # スクリーンショット取得（セッションが有効な場合）
    if [[ -n "$SESSION" ]]; then
        log_info "スクリーンショット取得試行"
        agent-browser --session "$SESSION" --json screenshot \
            --output "${ARTIFACTS_DIR}/screenshot.png" 2>/dev/null || true
    fi

    log_info "アーティファクト保存先: $ARTIFACTS_DIR"
}

# エラー時のハンドラ
on_error() {
    log_error "テスト失敗"
    collect_artifacts
    EXIT_CODE=1
}

trap cleanup EXIT
trap on_error ERR

# =============================================================================
# 前提条件チェック
# =============================================================================

check_prerequisites() {
    log_step "前提条件チェック"

    # 必須コマンド
    require_command jq
    require_command curl
    require_command agent-browser

    # Zeus バイナリ
    if [[ ! -x "${PROJECT_ROOT}/zeus" ]]; then
        log_error "Zeus バイナリが見つかりません: ${PROJECT_ROOT}/zeus"
        log_info "make build を実行してください"
        return 1
    fi

    # ダッシュボードビルド
    if [[ ! -d "${PROJECT_ROOT}/internal/dashboard/build" ]]; then
        log_error "ダッシュボードビルドが見つかりません"
        log_info "cd zeus-dashboard && npm ci && npm run build を実行してください"
        return 1
    fi

    # ゴールデンファイル
    require_file "${GOLDEN_DIR}/state/basic-tasks.json"

    log_success "前提条件チェック完了"
}

# =============================================================================
# テストプロジェクトセットアップ
# =============================================================================

setup_test_project() {
    log_step "テストプロジェクトセットアップ"

    TEST_DIR=$(create_temp_dir "zeus-e2e")
    log_info "テストディレクトリ: $TEST_DIR"

    create_test_project "$TEST_DIR" "${PROJECT_ROOT}/zeus"
}

# =============================================================================
# ダッシュボードサーバー起動
# =============================================================================

start_dashboard() {
    log_step "ダッシュボードサーバー起動"

    cd "$TEST_DIR"

    # サーバー起動（バックグラウンド）
    "${PROJECT_ROOT}/zeus" dashboard --port "$DASHBOARD_PORT" --no-open \
        > "${TEST_DIR}/server.log" 2>&1 &
    SERVER_PID=$!

    log_info "サーバー PID: $SERVER_PID"

    # サーバー起動待機
    if ! wait_for_port "$DASHBOARD_PORT" "$TIMEOUT_SERVER_START"; then
        log_error "サーバー起動タイムアウト"
        cat "${TEST_DIR}/server.log"
        return 1
    fi

    # API Ready 待機
    if ! wait_for_http "${API_URL}/api/status" "$TIMEOUT_API_READY"; then
        log_error "API Ready タイムアウト"
        cat "${TEST_DIR}/server.log"
        return 1
    fi

    log_success "ダッシュボードサーバー起動完了"
}

# =============================================================================
# ブラウザセッション
# =============================================================================

start_browser_session() {
    log_step "ブラウザセッション開始"

    # 新規セッション作成
    SESSION="zeus-e2e-$(date +%s)"
    log_info "セッション ID: $SESSION"

    # ブラウザ起動とページ遷移（open コマンドでブラウザ起動 + ページ遷移）
    # ?e2e パラメータで __ZEUS__ API を有効化
    log_info "ダッシュボードにアクセス中..."
    agent-browser --session "$SESSION" open "${API_URL}/?e2e" 2>/dev/null || {
        log_error "ブラウザ起動/ページ遷移失敗"
        return 1
    }

    log_success "ブラウザセッション開始完了"
}

# =============================================================================
# アプリケーション Ready 待機
# =============================================================================

wait_for_app_ready() {
    log_step "アプリケーション Ready 待機"

    local elapsed=0
    local ready="false"
    local api_checked=false

    while [[ "$ready" != "true" ]]; do
        if [[ $elapsed -ge $TIMEOUT_APP_READY ]]; then
            log_error "アプリケーション Ready タイムアウト（${TIMEOUT_APP_READY}秒）"
            return 1
        fi

        # 初回時は API 存在確認
        if [[ "$api_checked" == "false" ]]; then
            if ! check_zeus_api "$SESSION"; then
                # API 確認失敗は致命的
                return 1
            fi
            api_checked=true
        fi

        # __ZEUS__.isReady() をチェック（検証関数を使用）
        ready=$(eval_with_validation "$SESSION" \
            "window.__ZEUS__.isReady() ? 'true' : 'false'") || ready="false"

        if [[ "$ready" == "true" ]]; then
            break
        fi

        sleep 0.5
        elapsed=$((elapsed + 1))
    done

    log_success "アプリケーション Ready"
}

# =============================================================================
# グラフ状態取得
# =============================================================================

capture_graph_state() {
    log_step "グラフ状態取得"

    local result json_string
    # agent-browser eval で状態取得（検証付き）
    json_string=$(eval_with_validation "$SESSION" \
        "JSON.stringify(window.__ZEUS__.getGraphState())") || {
        log_error "グラフ状態取得失敗"
        return 1
    }

    # JSON 文字列をパース
    ACTUAL_STATE=$(echo "$json_string" | jq '.' 2>/dev/null) || {
        log_error "グラフ状態の JSON パース失敗"
        log_info "取得結果: $json_string"
        return 1
    }

    # 必須フィールドの確認
    local nodes_count edges_count
    nodes_count=$(echo "$ACTUAL_STATE" | jq '.nodes | length' 2>/dev/null) || {
        log_error "ノード情報がありません"
        return 1
    }
    edges_count=$(echo "$ACTUAL_STATE" | jq '.edges | length' 2>/dev/null) || {
        log_error "エッジ情報がありません"
        return 1
    }

    if [[ -z "$nodes_count" || -z "$edges_count" ]]; then
        log_error "グラフ状態が無効です"
        return 1
    fi

    log_info "取得したノード数: $nodes_count"
    log_info "取得したエッジ数: $edges_count"

    log_success "グラフ状態取得完了"
}

# =============================================================================
# メトリクス収集（情報のみ）
# =============================================================================

collect_metrics() {
    log_step "メトリクス収集（情報のみ）"

    local metrics json_string
    # メトリクス取得（検証付き）
    json_string=$(eval_with_validation "$SESSION" \
        "JSON.stringify(window.__VIEWER_METRICS__ || [])") || {
        log_warn "メトリクス取得失敗（非致命的）"
        metrics="[]"
        json_string="[]"
    }

    metrics=$(echo "$json_string" | jq '.' 2>/dev/null) || metrics="[]"

    local count
    count=$(echo "$metrics" | jq 'length' 2>/dev/null) || count=0

    log_info "収集メトリクス数: $count"

    # アーティファクトとして保存（KEEP_ARTIFACTS=true の場合）
    if [[ "$KEEP_ARTIFACTS" == "true" && -n "$metrics" ]]; then
        mkdir -p "${ARTIFACTS_DIR}"
        save_artifact "metrics.json" "$metrics"
    fi
}

# =============================================================================
# アーティファクト保存（KEEP_ARTIFACTS=true の場合）
# =============================================================================

save_state_artifact() {
    if [[ "$KEEP_ARTIFACTS" == "true" && -n "${ACTUAL_STATE:-}" ]]; then
        mkdir -p "${ARTIFACTS_DIR}"
        save_artifact "actual-state.json" "$ACTUAL_STATE"
    fi
}

# =============================================================================
# 構造比較
# =============================================================================

run_verification() {
    log_step "構造比較実行"

    if ! verify_state "$ACTUAL_STATE" "${GOLDEN_DIR}/state/basic-tasks.json"; then
        log_error "構造比較失敗"
        return 1
    fi

    log_success "構造比較成功"
}

# =============================================================================
# メイン
# =============================================================================

main() {
    log_step "Zeus E2E テスト開始"
    echo ""

    check_prerequisites
    setup_test_project
    start_dashboard
    start_browser_session
    wait_for_app_ready
    capture_graph_state
    save_state_artifact
    collect_metrics
    run_verification

    echo ""
    log_success "============================================"
    log_success "Zeus E2E テスト: 全て成功"
    log_success "============================================"
}

main "$@"
