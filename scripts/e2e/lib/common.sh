#!/bin/bash
# common.sh - E2E テスト用共通ライブラリ
# ログ関数、設定、ユーティリティを提供

set -euo pipefail

# =============================================================================
# 設定
# =============================================================================

# タイムアウト設定（秒）
export TIMEOUT_SERVER_START=${TIMEOUT_SERVER_START:-30}
export TIMEOUT_API_READY=${TIMEOUT_API_READY:-10}
export TIMEOUT_APP_READY=${TIMEOUT_APP_READY:-20}
export TIMEOUT_CAPTURE=${TIMEOUT_CAPTURE:-5}

# ポート設定
export DASHBOARD_PORT=${DASHBOARD_PORT:-18080}
export API_URL="http://localhost:${DASHBOARD_PORT}"

# ディレクトリ設定
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export SCRIPT_DIR
export LIB_DIR="${SCRIPT_DIR}/lib"
export GOLDEN_DIR="${SCRIPT_DIR}/golden"

# アーティファクト設定
export ARTIFACTS_DIR="${ARTIFACTS_DIR:-/tmp/zeus-e2e-artifacts}"
export KEEP_ARTIFACTS=${KEEP_ARTIFACTS:-false}

# =============================================================================
# カラー出力
# =============================================================================

# ANSI カラーコード
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    BOLD='\033[1m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    BOLD=''
    NC=''
fi

export RED GREEN YELLOW BLUE BOLD NC

# =============================================================================
# ログ関数
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $*" >&2
}

log_step() {
    echo -e "${BOLD}==> $*${NC}"
}

export -f log_info log_success log_warn log_error log_step

# =============================================================================
# ユーティリティ関数
# =============================================================================

# コマンド存在チェック
require_command() {
    local cmd="$1"
    if ! command -v "$cmd" &>/dev/null; then
        log_error "必須コマンドが見つかりません: $cmd"
        return 1
    fi
}

# ポート使用中チェック
is_port_in_use() {
    local port="$1"
    lsof -i ":${port}" &>/dev/null
}

# ポート空きチェック（待機）
wait_for_port() {
    local port="$1"
    local timeout="${2:-30}"
    local elapsed=0

    while ! is_port_in_use "$port"; do
        if [[ $elapsed -ge $timeout ]]; then
            log_error "ポート $port の待機がタイムアウト（${timeout}秒）"
            return 1
        fi
        sleep 0.5
        elapsed=$((elapsed + 1))
    done

    log_info "ポート $port が使用可能になりました"
}

# HTTP エンドポイント待機
wait_for_http() {
    local url="$1"
    local timeout="${2:-10}"
    local elapsed=0

    while ! curl -sf "$url" &>/dev/null; do
        if [[ $elapsed -ge $timeout ]]; then
            log_error "HTTP エンドポイント待機がタイムアウト: $url （${timeout}秒）"
            return 1
        fi
        sleep 0.5
        elapsed=$((elapsed + 1))
    done

    log_info "HTTP エンドポイントが応答: $url"
}

# テンポラリディレクトリ作成
create_temp_dir() {
    local prefix="${1:-zeus-e2e}"
    mktemp -d "/tmp/${prefix}.XXXXXX"
}

# JSON からフィールド抽出
json_get() {
    local json="$1"
    local path="$2"
    echo "$json" | jq -r "$path"
}

# ファイルが存在するか確認
require_file() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        log_error "必須ファイルが見つかりません: $file"
        return 1
    fi
}

# プロセス終了（安全）
safe_kill() {
    local pid="$1"
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null || true
        # プロセス終了を待機
        local waited=0
        while kill -0 "$pid" 2>/dev/null && [[ $waited -lt 5 ]]; do
            sleep 0.5
            waited=$((waited + 1))
        done
        # まだ生きていれば強制終了
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi
    fi
}

export -f require_command is_port_in_use wait_for_port wait_for_http
export -f create_temp_dir json_get require_file safe_kill

# =============================================================================
# agent-browser レスポンス検証
# =============================================================================

# agent-browser JSON レスポンスを検証・解析
# @param $1: JSON レスポンス
# @return: 成功時は .data.result の内容を出力、失敗時は空を出力
validate_agent_browser_response() {
    local response="$1"
    local operation="${2:-eval}"  # eval, snapshot, screenshot など

    # JSON 形式の妥当性チェック
    if ! echo "$response" | jq empty 2>/dev/null; then
        log_error "agent-browser: 無効な JSON レスポンス（$operation）"
        log_info "レスポンス: $response"
        return 1
    fi

    # success フィールドの確認
    local success
    success=$(echo "$response" | jq -r '.success // false')
    if [[ "$success" != "true" ]]; then
        local error
        error=$(echo "$response" | jq -r '.error // "Unknown error"')
        log_error "agent-browser: 操作失敗（$operation）: $error"
        return 1
    fi

    # .data.result フィールドの抽出
    local result
    result=$(echo "$response" | jq -r '.data.result // empty' 2>/dev/null)

    if [[ -z "$result" ]]; then
        log_warn "agent-browser: .data.result が空です（$operation）"
        return 1
    fi

    echo "$result"
}

# agent-browser eval コマンドの実行と検証
# @param $1: セッション ID
# @param $2: JavaScript コード
# @return: eval 結果（.data.result の内容）
eval_with_validation() {
    local session="$1"
    local js_code="$2"

    local response
    response=$(agent-browser --session "$session" --json eval "$js_code" 2>/dev/null) || {
        log_error "agent-browser eval: コマンド実行失敗"
        return 1
    }

    validate_agent_browser_response "$response" "eval"
}

# window.__ZEUS__ API の存在確認
# @param $1: セッション ID
# @return: 0 (API 存在), 1 (API 不存在)
check_zeus_api() {
    local session="$1"
    local result

    result=$(eval_with_validation "$session" \
        "typeof window.__ZEUS__ !== 'undefined' && typeof window.__ZEUS__.isReady === 'function' ? 'ok' : 'missing'") || {
        log_error "window.__ZEUS__ API チェック失敗"
        return 1
    }

    if [[ "$result" != "ok" ]]; then
        log_error "window.__ZEUS__ API が見つかりません（API 未公開またはスクリプト読み込み失敗）"
        log_info "確認項目:"
        log_info "  1. ダッシュボードが ?e2e パラメータ付きで起動しているか確認"
        log_info "  2. FactorioViewer.svelte が window.__ZEUS__ を公開しているか確認"
        return 1
    fi

    log_info "window.__ZEUS__ API が正常に利用可能です"
    return 0
}

export -f validate_agent_browser_response eval_with_validation check_zeus_api

# =============================================================================
# アーティファクト収集
# =============================================================================

# アーティファクトディレクトリを初期化
init_artifacts_dir() {
    rm -rf "${ARTIFACTS_DIR}"
    mkdir -p "${ARTIFACTS_DIR}"
}

# アーティファクトを保存
save_artifact() {
    local name="$1"
    local content="$2"
    echo "$content" > "${ARTIFACTS_DIR}/${name}"
    log_info "アーティファクト保存: ${ARTIFACTS_DIR}/${name}"
}

# ファイルをアーティファクトにコピー
copy_artifact() {
    local src="$1"
    local name="$2"
    if [[ -f "$src" ]]; then
        cp "$src" "${ARTIFACTS_DIR}/${name}"
        log_info "アーティファクトコピー: ${ARTIFACTS_DIR}/${name}"
    fi
}

export -f init_artifacts_dir save_artifact copy_artifact

# =============================================================================
# テストプロジェクト作成
# =============================================================================

# 基本的なテストプロジェクトを作成（3タスクのチェーン）
create_test_project() {
    local dir="$1"
    local zeus_bin="${2:-./zeus}"

    log_step "テストプロジェクト作成: $dir"

    # ディレクトリ作成
    mkdir -p "$dir"
    cd "$dir"

    # Zeus 初期化
    "$zeus_bin" init

    # タスク追加（JSON出力でIDを取得）
    local task_a_result task_a_id
    task_a_result=$("$zeus_bin" add task "Task A" -f json)
    task_a_id=$(echo "$task_a_result" | jq -r '.ID')
    log_info "Task A 作成: $task_a_id"

    "$zeus_bin" add task "Task B"
    log_info "Task B 作成"

    "$zeus_bin" add task "Task C" --parent "$task_a_id"
    log_info "Task C 作成 (親: $task_a_id)"

    log_success "テストプロジェクト作成完了"
}

export -f create_test_project
