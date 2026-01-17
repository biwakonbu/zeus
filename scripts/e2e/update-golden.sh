#!/bin/bash
# update-golden.sh - ゴールデンファイル更新スクリプト
# 実際の状態をキャプチャしてゴールデンファイルを更新

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
# 設定
# =============================================================================

GOLDEN_FILE="${GOLDEN_DIR}/state/basic-tasks.json"
BACKUP_FILE="${GOLDEN_FILE}.backup"

# =============================================================================
# メイン
# =============================================================================

main() {
    log_step "ゴールデンファイル更新開始"
    echo ""

    # アーティファクト保持モードで実行
    log_info "テスト実行（アーティファクト保持モード）"
    export KEEP_ARTIFACTS=true

    # run-web-test.sh を実行（エラーを無視）
    if "${SCRIPT_DIR}/run-web-test.sh"; then
        log_success "テスト成功（既存ゴールデンと一致）"
        log_info "ゴールデンファイルの更新は不要です"
        return 0
    fi

    log_warn "テスト失敗（ゴールデン更新が必要）"
    echo ""

    # アーティファクトから実際の状態を取得
    if [[ ! -f "${ARTIFACTS_DIR}/actual-state.json" ]]; then
        log_error "actual-state.json が見つかりません"
        log_info "テスト実行中にエラーが発生した可能性があります"
        return 1
    fi

    local actual_state
    actual_state=$(cat "${ARTIFACTS_DIR}/actual-state.json")

    # 現在のゴールデンをバックアップ
    if [[ -f "$GOLDEN_FILE" ]]; then
        cp "$GOLDEN_FILE" "$BACKUP_FILE"
        log_info "バックアップ作成: $BACKUP_FILE"
    fi

    # ゴールデン形式に変換して保存
    log_step "ゴールデンファイル生成"

    # スキーマバージョン確認（既存ゴールデンがある場合）
    if [[ -f "$BACKUP_FILE" ]]; then
        local current_schema
        current_schema=$(jq -r '.metadata.schema_version' "$BACKUP_FILE" 2>/dev/null) || {
            log_warn "既存ゴールデンのスキーマバージョンが読めません"
            current_schema="unknown"
        }
        log_info "現在のスキーマバージョン: $current_schema"
    fi

    local new_golden
    new_golden=$(convert_to_golden_format "$actual_state" \
        "basic-tasks-001" \
        "Basic project with 3 tasks forming a chain")

    # スキーマバージョン確認
    local new_schema
    new_schema=$(echo "$new_golden" | jq -r '.metadata.schema_version' 2>/dev/null) || {
        log_error "新規ゴールデンのスキーマバージョン抽出失敗"
        return 1
    }
    log_info "新規スキーマバージョン: $new_schema"

    echo "$new_golden" > "$GOLDEN_FILE"
    log_success "ゴールデンファイル更新: $GOLDEN_FILE"
    echo ""

    # 差分表示
    if [[ -f "$BACKUP_FILE" ]]; then
        log_step "変更差分"
        echo ""
        diff -u "$BACKUP_FILE" "$GOLDEN_FILE" || true
        echo ""
        rm "$BACKUP_FILE"
    fi

    # レビュー促進メッセージ
    echo ""
    log_warn "================================================"
    log_warn "重要: ゴールデンファイルが更新されました"
    log_warn "================================================"
    echo ""
    log_info "以下のコマンドで変更を確認してください:"
    echo ""
    echo "  git diff ${GOLDEN_FILE}"
    echo ""
    log_info "変更が意図的な場合のみコミットしてください:"
    echo ""
    echo "  git add ${GOLDEN_FILE}"
    echo "  git commit -m 'chore: update E2E golden files'"
    echo ""
}

main "$@"
