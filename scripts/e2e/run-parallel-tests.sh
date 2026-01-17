#!/bin/bash
# run-parallel-tests.sh - 複数テストシナリオの並列実行
# 異なるプロジェクト構成でのテストを並列実行し、統合レポートを生成

set -euo pipefail

# =============================================================================
# 初期化
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# ライブラリ読み込み
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

# =============================================================================
# グローバル変数
# =============================================================================

# テストシナリオ定義
declare -A TEST_SCENARIOS=(
    [basic]="basic-tasks"
    [complex]="complex-dependencies"
    [large]="large-project"
)

# 実行結果追跡
declare -A SCENARIO_PIDS
declare -A SCENARIO_RESULTS
declare -A SCENARIO_ARTIFACTS
PARALLEL_JOBS=3
EXIT_CODE=0

# =============================================================================
# ユーティリティ関数
# =============================================================================

# テストシナリオ実行（バックグラウンド）
run_scenario() {
    local scenario_name="$1"
    local scenario_config="$2"

    log_step "[$scenario_name] テスト開始"

    # シナリオ別の環境設定
    local scenario_port=$((DASHBOARD_PORT + 1000 + RANDOM % 1000))
    local scenario_artifacts="/tmp/zeus-e2e-${scenario_name}-$(date +%s)"

    export DASHBOARD_PORT="$scenario_port"
    export ARTIFACTS_DIR="$scenario_artifacts"
    export KEEP_ARTIFACTS=true

    # テスト実行（バックグラウンド）
    if "${SCRIPT_DIR}/run-web-test.sh" > "${scenario_artifacts}/test.log" 2>&1; then
        SCENARIO_RESULTS[$scenario_name]="PASS"
        log_success "[$scenario_name] テスト完了: PASS"
    else
        SCENARIO_RESULTS[$scenario_name]="FAIL"
        log_error "[$scenario_name] テスト完了: FAIL"
        EXIT_CODE=1
    fi

    # アーティファクト記録
    SCENARIO_ARTIFACTS[$scenario_name]="$scenario_artifacts"
}

# テストシナリオ実行（並列）
run_all_scenarios() {
    log_step "並列テスト実行開始"
    echo ""

    local active_jobs=0

    for scenario_name in "${!TEST_SCENARIOS[@]}"; do
        # 並列ジョブ数制御
        if [[ $active_jobs -ge $PARALLEL_JOBS ]]; then
            # 1つのジョブ完了を待機
            wait -n 2>/dev/null || true
            active_jobs=$((active_jobs - 1))
        fi

        # シナリオ実行（バックグラウンド）
        run_scenario "$scenario_name" "${TEST_SCENARIOS[$scenario_name]}" &
        SCENARIO_PIDS[$scenario_name]=$!
        active_jobs=$((active_jobs + 1))

        log_info "[$scenario_name] バックグラウンド実行開始: PID ${SCENARIO_PIDS[$scenario_name]}"
    done

    # 全バックグラウンドジョブ完了待機
    log_info "全シナリオの完了を待機中..."
    for scenario_name in "${!SCENARIO_PIDS[@]}"; do
        local pid="${SCENARIO_PIDS[$scenario_name]}"
        wait "$pid" 2>/dev/null || {
            log_warn "[$scenario_name] バックグラウンドジョブがエラーで終了（PID: $pid）"
        }
    done

    echo ""
    log_success "並列テスト実行完了"
}

# 統合レポート生成
generate_summary_report() {
    log_step "統合レポート生成"

    local total_scenarios=${#TEST_SCENARIOS[@]}
    local passed_scenarios=0
    local failed_scenarios=0

    echo ""
    echo "=== テストシナリオ結果 ==="
    echo ""

    for scenario_name in "${!TEST_SCENARIOS[@]}"; do
        local status="${SCENARIO_RESULTS[$scenario_name]:-UNKNOWN}"
        local artifacts="${SCENARIO_ARTIFACTS[$scenario_name]:-N/A}"

        if [[ "$status" == "PASS" ]]; then
            ((passed_scenarios++))
            log_success "✓ $scenario_name: $status"
        else
            ((failed_scenarios++))
            log_error "✗ $scenario_name: $status"
        fi

        log_info "  アーティファクト: $artifacts"
    done

    echo ""
    echo "=== 集計 ==="
    echo ""
    log_info "総シナリオ数: $total_scenarios"
    log_success "成功: $passed_scenarios"
    log_error "失敗: $failed_scenarios"

    if [[ $failed_scenarios -eq 0 ]]; then
        log_success "========================================="
        log_success "全並列テスト: 成功"
        log_success "========================================="
    else
        log_error "========================================="
        log_error "全並列テスト: 失敗（$failed_scenarios/$total_scenarios）"
        log_error "========================================="
        return 1
    fi
}

# 統合 JSON レポート生成
generate_json_summary() {
    log_step "統合 JSON レポート生成"

    local summary_file="${ARTIFACTS_DIR}/parallel-test-summary.json"
    mkdir -p "$(dirname "$summary_file")"

    # 個別テスト結果を収集
    local scenario_results="["
    local first=true

    for scenario_name in "${!TEST_SCENARIOS[@]}"; do
        local status="${SCENARIO_RESULTS[$scenario_name]:-UNKNOWN}"
        local artifacts="${SCENARIO_ARTIFACTS[$scenario_name]:-}"

        if [[ -f "${artifacts}/test-report.json" ]]; then
            local report
            report=$(cat "${artifacts}/test-report.json")
        else
            report='{"error": "レポート不見つかり"}'
        fi

        if [[ "$first" == false ]]; then
            scenario_results="$scenario_results,"
        fi
        first=false

        scenario_results="$scenario_results{
            \"scenario\": \"$scenario_name\",
            \"status\": \"$status\",
            \"report\": $report
        }"
    done

    scenario_results="$scenario_results]"

    # JSON 要約
    local summary
    summary=$(jq -n \
        --arg timestamp "$(date +%Y-%m-%dT%H:%M:%SZ)" \
        --argjson total ${#TEST_SCENARIOS[@]} \
        --argjson passed $(( ${#SCENARIO_RESULTS[@]} - $(printf '%s\n' "${SCENARIO_RESULTS[@]}" | grep -c "FAIL" || echo 0) )) \
        --argjson failed $(printf '%s\n' "${SCENARIO_RESULTS[@]}" | grep -c "FAIL" || echo 0) \
        --argjson scenarios "$scenario_results" \
        '{
            metadata: {
                timestamp: $timestamp,
                execution_mode: "parallel",
                total_scenarios: $total,
                passed_scenarios: $passed,
                failed_scenarios: $failed
            },
            scenarios: $scenarios
        }')

    echo "$summary" > "$summary_file"
    log_success "統合レポート保存: $summary_file"

    # スクリーンに出力
    echo ""
    echo "$summary" | jq .
}

# =============================================================================
# メイン
# =============================================================================

main() {
    log_step "Zeus E2E 並列テスト実行"
    echo ""

    # 前提条件チェック
    if [[ ! -x "${PROJECT_ROOT}/zeus" ]]; then
        log_error "Zeus バイナリが見つかりません: ${PROJECT_ROOT}/zeus"
        log_info "make build を実行してください"
        return 1
    fi

    if [[ ! -d "${PROJECT_ROOT}/internal/dashboard/build" ]]; then
        log_error "ダッシュボードビルドが見つかりません"
        log_info "cd zeus-dashboard && npm ci && npm run build を実行してください"
        return 1
    fi

    # 並列テスト実行
    run_all_scenarios

    # レポート生成
    echo ""
    generate_summary_report || EXIT_CODE=$?
    echo ""
    generate_json_summary

    return $EXIT_CODE
}

main "$@"
