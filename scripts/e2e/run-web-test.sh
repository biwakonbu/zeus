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
# shellcheck source=lib/report.sh
source "${SCRIPT_DIR}/lib/report.sh"

# =============================================================================
# グローバル変数
# =============================================================================

TEST_DIR=""
SERVER_PID=""
SESSION=""
EXIT_CODE=0
TEST_START_TIME=""
TEST_END_TIME=""
STEP_COUNT=0
STEP_PASSED=0

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

    local artifacts_collected=0
    local artifacts_failed=0

    # 現在の状態をJSON として保存（取得済みの場合）
    if [[ -n "${ACTUAL_STATE:-}" ]]; then
        if save_artifact "actual-state.json" "$ACTUAL_STATE" 2>/dev/null; then
            ((artifacts_collected++))
            log_info "✓ 状態ファイル保存"
        else
            ((artifacts_failed++))
            log_warn "✗ 状態ファイル保存失敗"
        fi
    fi

    # サーバーログがあれば保存
    if [[ -f "${TEST_DIR:-/tmp}/server.log" ]]; then
        if copy_artifact "${TEST_DIR}/server.log" "server.log" 2>/dev/null; then
            ((artifacts_collected++))
            log_info "✓ サーバーログ保存"
        else
            ((artifacts_failed++))
            log_warn "✗ サーバーログ保存失敗"
        fi
    fi

    # テストプロジェクトの状態
    if [[ -d "${TEST_DIR:-/tmp}/.zeus" ]]; then
        if tar -czf "${ARTIFACTS_DIR}/zeus-data.tar.gz" -C "$TEST_DIR" .zeus 2>/dev/null; then
            ((artifacts_collected++))
            log_info "✓ Zeus データ保存"
        else
            ((artifacts_failed++))
            log_warn "✗ Zeus データ保存失敗"
        fi
    fi

    # スクリーンショット取得（セッションが有効な場合）
    if [[ -n "$SESSION" ]]; then
        if agent-browser --session "$SESSION" --json screenshot \
            --output "${ARTIFACTS_DIR}/screenshot.png" 2>/dev/null; then
            ((artifacts_collected++))
            log_info "✓ スクリーンショット保存"
        else
            ((artifacts_failed++))
            log_warn "✗ スクリーンショット保存失敗"
        fi
    fi

    log_info "アーティファクト収集完了: 成功=$artifacts_collected, 失敗=$artifacts_failed"
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

    # 環境設定の妥当性チェック
    validate_environment || return 1

    # ポートの利用可能性チェック
    check_port_available "$DASHBOARD_PORT" || return 1

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

    # テストディレクトリに移動（失敗ハンドリング）
    if ! cd "$TEST_DIR"; then
        log_error "テストディレクトリへの移動に失敗: $TEST_DIR"
        return 1
    fi

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
    if ! agent-browser --session "$SESSION" open "${API_URL}/?e2e" 2>/dev/null; then
        log_error "ブラウザ起動/ページ遷移失敗"
        log_info "確認項目:"
        log_info "  1. agent-browser がインストール済みか確認: agent-browser install"
        log_info "  2. ダッシュボードサーバーが起動しているか確認: curl $API_URL"
        log_info "  3. ネットワーク接続を確認"
        return 1
    fi

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
        json_string="[]"
    }

    # JSON パースと null チェック
    metrics=$(echo "$json_string" | jq '.' 2>/dev/null) || metrics="[]"

    if [[ -z "$metrics" || "$metrics" == "null" ]]; then
        log_warn "メトリクスが無効です（null またはパース失敗）"
        metrics="[]"
    fi

    local count
    count=$(echo "$metrics" | jq 'length' 2>/dev/null) || count=0

    # count の妥当性チェック
    if [[ -z "$count" || ! "$count" =~ ^[0-9]+$ ]]; then
        log_warn "メトリクス数が不正です: $count"
        count=0
    fi

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

    # テスト実行時刻をログに記録
    TEST_START_TIME=$(date '+%Y-%m-%d %H:%M:%S')
    log_info "テスト実行時刻: $TEST_START_TIME"
    log_info "ZEUS: $PROJECT_ROOT"
    log_info "アーティファクト保存先: $ARTIFACTS_DIR"
    echo ""

    # テストステップの実行と追跡
    ((STEP_COUNT++))
    if check_prerequisites; then ((STEP_PASSED++)); fi

    ((STEP_COUNT++))
    if setup_test_project; then ((STEP_PASSED++)); fi

    ((STEP_COUNT++))
    if start_dashboard; then ((STEP_PASSED++)); fi

    ((STEP_COUNT++))
    if start_browser_session; then ((STEP_PASSED++)); fi

    ((STEP_COUNT++))
    if wait_for_app_ready; then ((STEP_PASSED++)); fi

    ((STEP_COUNT++))
    if capture_graph_state; then ((STEP_PASSED++)); fi

    ((STEP_COUNT++))
    if save_state_artifact; then ((STEP_PASSED++)); fi

    ((STEP_COUNT++))
    if collect_metrics; then ((STEP_PASSED++)); fi

    ((STEP_COUNT++))
    if run_verification; then ((STEP_PASSED++)); fi

    echo ""

    # テスト完了時刻をログに記録
    TEST_END_TIME=$(date '+%Y-%m-%d %H:%M:%S')
    log_info "テスト完了時刻: $TEST_END_TIME"

    # テスト統計を記録
    echo ""
    record_test_stats "$TEST_START_TIME" "$TEST_END_TIME" "$STEP_PASSED" "$STEP_COUNT"

    # レポート生成（KEEP_ARTIFACTS=true の場合）
    if [[ "$KEEP_ARTIFACTS" == "true" ]]; then
        # JSON レポート保存
        save_test_report_json "$TEST_START_TIME" "$TEST_END_TIME" "$STEP_PASSED" "$STEP_COUNT"

        # 複数形式レポート生成
        echo ""
        log_step "テストレポート生成"
        generate_test_reports "$ARTIFACTS_DIR" "$TEST_START_TIME" "$TEST_END_TIME" "$STEP_PASSED" "$STEP_COUNT" "${GOLDEN_DIR}/state/basic-tasks.json"
    fi

    echo ""
    log_success "============================================"
    log_success "Zeus E2E テスト: 全て成功"
    log_success "============================================"
}

main "$@"
