#!/bin/bash
# report.sh - E2E テストレポート生成ライブラリ
# JSON/HTML/Markdown 形式での統合レポート出力

set -euo pipefail

# common.sh を読み込み
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

# =============================================================================
# Markdown レポート生成
# =============================================================================

# テスト結果を Markdown で出力
generate_markdown_report() {
    local report_file="$1"
    local test_start="$2"
    local test_end="$3"
    local steps_passed="$4"
    local steps_total="$5"
    local golden_file="$6"

    local success_rate=0
    if [[ $steps_total -gt 0 ]]; then
        success_rate=$((steps_passed * 100 / steps_total))
    fi

    local status_badge="❌ FAIL"
    if [[ $success_rate -eq 100 ]]; then
        status_badge="✅ PASS"
    fi

    cat > "$report_file" <<EOF
# Zeus E2E テスト レポート

## 実行概要

| 項目 | 値 |
|------|-----|
| 実行日時 | \`$test_start\` 〜 \`$test_end\` |
| 結果 | $status_badge |
| 成功ステップ | $steps_passed/$steps_total |
| 成功率 | **${success_rate}%** |

## テストステップ詳細

### ✓ 実行されたステップ

1. **前提条件チェック** - Zeus バイナリ、ダッシュボード、必須コマンド確認
2. **テストプロジェクトセットアップ** - テンポラリディレクトリ作成、プロジェクト初期化
3. **ダッシュボードサーバー起動** - Go サーバー起動、API Ready 待機
4. **ブラウザセッション開始** - agent-browser でページを開く
5. **アプリケーション Ready 待機** - \`window.__ZEUS__.isReady()\` で描画完了を待機
6. **グラフ状態取得** - \`window.__ZEUS__.getGraphState()\` で状態を JSON でキャプチャ
7. **状態アーティファクト保存** - キャプチャした状態を JSON ファイルに保存
8. **メトリクス収集** - \`window.__VIEWER_METRICS__\` から計測ログを収集（情報のみ）
9. **構造比較実行** - 実際の状態をゴールデンファイルと構造比較

## 検証方式

### 座標除外の構造比較

このテストは以下のフィールドを**完全に除外**して比較します:

- \`nodes[*].x\`, \`nodes[*].y\` - 描画座標（レイアウトアルゴリズムに依存）
- \`nodes[*].id\` - UUID（動的生成される）
- \`viewport\` - ビューポート情報（環境依存）

**比較対象:**
- \`nodes[*].name\` - タスク名
- \`nodes[*].status\` - ステータス（pending/in_progress/completed）
- \`nodes[*].progress\` - 進捗度（0-100）
- \`edges[].from\`, \`edges[].to\` - 依存関係（名前ベース）

これにより、**構造的な正確性**を検証しながら、**レイアウト変更の影響を排除**します。

## ゴールデンファイル

**参照ファイル:** \`$golden_file\`

ゴールデンファイルは以下を含みます:

- \`metadata\` - テスト ID、スキーマバージョン、作成日
- \`comparison\` - 比較モード、除外フィールド、エッジ比較方式
- \`expected\` - 期待値（nodes, edges, counts）

## 環境情報

- **Zeus バージョン:** $(cd "${PROJECT_ROOT}" && ./zeus --version 2>/dev/null || echo 'unknown')
- **Git ブランチ:** $(git -C "${PROJECT_ROOT}" rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')
- **Git コミット:** $(git -C "${PROJECT_ROOT}" rev-parse --short HEAD 2>/dev/null || echo 'unknown')
- **ダッシュボードポート:** \`$DASHBOARD_PORT\`
- **アーティファクト:** \`$ARTIFACTS_DIR\`

## トラブルシューティング

### テスト失敗時

1. **アーティファクト確認**
   \`\`\`bash
   ls -lh $ARTIFACTS_DIR/
   cat $ARTIFACTS_DIR/actual-state.json | jq .
   \`\`\`

2. **サーバーログ確認**
   \`\`\`bash
   cat $ARTIFACTS_DIR/server.log
   \`\`\`

3. **スクリーンショット確認**
   \`\`\`bash
   open $ARTIFACTS_DIR/screenshot.png
   \`\`\`

4. **差分確認**
   \`\`\`bash
   diff <(jq . $ARTIFACTS_DIR/actual-state.json) <(jq . $golden_file)
   \`\`\`

## ゴールデンファイル更新

ゴールデンファイルを更新する場合:

\`\`\`bash
./scripts/e2e/update-golden.sh
\`\`\`

更新内容を確認してからコミットしてください:

\`\`\`bash
git diff scripts/e2e/golden/
git add scripts/e2e/golden/
git commit -m 'chore: update E2E golden files'
\`\`\`

---

*このレポートは Zeus E2E テストスイートから自動生成されました。*
*$(date '+%Y-%m-%d %H:%M:%S')*
EOF

    log_info "Markdown レポート生成: $report_file"
}

# HTML レポート生成
generate_html_report() {
    local report_file="$1"
    local test_start="$2"
    local test_end="$3"
    local steps_passed="$4"
    local steps_total="$5"

    local success_rate=0
    if [[ $steps_total -gt 0 ]]; then
        success_rate=$((steps_passed * 100 / steps_total))
    fi

    local status_color="dc2626"  # 赤
    local status_text="FAIL"
    if [[ $success_rate -eq 100 ]]; then
        status_color="16a34a"  # 緑
        status_text="PASS"
    fi

    cat > "$report_file" <<EOF
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Zeus E2E テストレポート</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; padding: 20px; }
        .container { max-width: 900px; margin: 0 auto; background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        header { background: #1f2937; color: white; padding: 30px; border-radius: 8px 8px 0 0; }
        h1 { margin-bottom: 10px; font-size: 28px; }
        .subtitle { font-size: 14px; opacity: 0.9; }
        .content { padding: 30px; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .stat-card { background: #f9fafb; border-left: 4px solid #3b82f6; padding: 20px; border-radius: 4px; }
        .stat-label { font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 8px; }
        .stat-value { font-size: 24px; font-weight: bold; color: #1f2937; }
        .stat-card.status { border-left-color: #$status_color; }
        .stat-card.status .stat-value { color: #$status_color; }
        .badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; }
        .badge.pass { background: #dcfce7; color: #166534; }
        .badge.fail { background: #fee2e2; color: #991b1b; }
        .section { margin-bottom: 30px; }
        .section h2 { font-size: 18px; margin-bottom: 15px; color: #1f2937; border-bottom: 2px solid #e5e7eb; padding-bottom: 10px; }
        table { width: 100%; border-collapse: collapse; }
        th { background: #f3f4f6; text-align: left; padding: 12px; font-weight: 600; font-size: 13px; color: #374151; }
        td { padding: 12px; border-bottom: 1px solid #e5e7eb; }
        tr:hover { background: #f9fafb; }
        .progress-bar { width: 100%; height: 8px; background: #e5e7eb; border-radius: 4px; overflow: hidden; }
        .progress-fill { height: 100%; background: #3b82f6; }
        .progress-fill.high { background: #16a34a; }
        .progress-fill.low { background: #dc2626; }
        footer { border-top: 1px solid #e5e7eb; padding: 20px; font-size: 12px; color: #6b7280; text-align: center; }
        .timestamp { color: #9ca3af; font-size: 12px; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🧪 Zeus E2E テストレポート</h1>
            <p class="subtitle">自動テストスイート実行結果</p>
        </header>
        <div class="content">
            <!-- 統計情報 -->
            <div class="stats">
                <div class="stat-card status">
                    <div class="stat-label">テスト結果</div>
                    <div class="stat-value"><span class="badge $([ $success_rate -eq 100 ] && echo 'pass' || echo 'fail')">$status_text</span></div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">成功率</div>
                    <div class="stat-value">${success_rate}%</div>
                    <div class="progress-bar">
                        <div class="progress-fill $([ $success_rate -eq 100 ] && echo 'high' || echo 'low')" style="width: ${success_rate}%;"></div>
                    </div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">実行ステップ</div>
                    <div class="stat-value">$steps_passed/<span style="opacity: 0.7;">$steps_total</span></div>
                </div>
            </div>

            <!-- タイミング情報 -->
            <div class="section">
                <h2>実行タイミング</h2>
                <table>
                    <tr>
                        <th style="width: 150px;">項目</th>
                        <th>値</th>
                    </tr>
                    <tr>
                        <td>開始時刻</td>
                        <td><code>$test_start</code></td>
                    </tr>
                    <tr>
                        <td>終了時刻</td>
                        <td><code>$test_end</code></td>
                    </tr>
                </table>
            </div>

            <!-- テストステップ -->
            <div class="section">
                <h2>テストステップ</h2>
                <table>
                    <tr>
                        <th style="width: 40px;">#</th>
                        <th>ステップ</th>
                        <th style="width: 100px;">説明</th>
                    </tr>
                    <tr><td>1</td><td>前提条件チェック</td><td>環境検証</td></tr>
                    <tr><td>2</td><td>テストプロジェクトセットアップ</td><td>初期化</td></tr>
                    <tr><td>3</td><td>ダッシュボードサーバー起動</td><td>インフラ</td></tr>
                    <tr><td>4</td><td>ブラウザセッション開始</td><td>インフラ</td></tr>
                    <tr><td>5</td><td>アプリケーション Ready 待機</td><td>同期</td></tr>
                    <tr><td>6</td><td>グラフ状態取得</td><td>データキャプチャ</td></tr>
                    <tr><td>7</td><td>状態アーティファクト保存</td><td>データ保存</td></tr>
                    <tr><td>8</td><td>メトリクス収集</td><td>計測</td></tr>
                    <tr><td>9</td><td>構造比較実行</td><td>検証</td></tr>
                </table>
            </div>

            <!-- 検証方式 -->
            <div class="section">
                <h2>検証方式</h2>
                <p style="line-height: 1.6; color: #374151; margin-bottom: 15px;">
                    このテストは <strong>座標除外の構造比較</strong> を採用しています。
                    ノードの位置情報（x, y座標）やビューポート情報など、環境依存的な要素を除外し、
                    タスク名、ステータス、進捗度、依存関係など構造的な要素のみを検証します。
                </p>
                <p style="line-height: 1.6; color: #374151; font-size: 13px; background: #f3f4f6; padding: 12px; border-radius: 4px;">
                    <strong>除外フィールド:</strong> nodes[*].x, nodes[*].y, nodes[*].id, viewport<br>
                    <strong>検証対象:</strong> nodes[*].name, status, progress; edges[].from, edges[].to
                </p>
            </div>
        </div>
        <footer>
            <div class="timestamp">生成日時: $(date '+%Y-%m-%d %H:%M:%S')</div>
        </footer>
    </div>
</body>
</html>
EOF

    log_info "HTML レポート生成: $report_file"
}

# テキストレポート生成
generate_text_report() {
    local report_file="$1"
    local test_start="$2"
    local test_end="$3"
    local steps_passed="$4"
    local steps_total="$5"
    local golden_file="$6"

    local success_rate=0
    if [[ $steps_total -gt 0 ]]; then
        success_rate=$((steps_passed * 100 / steps_total))
    fi

    cat > "$report_file" <<EOF
================================================================================
                    Zeus E2E テスト レポート
================================================================================

実行日時: $test_start 〜 $test_end
テスト結果: $([ $success_rate -eq 100 ] && echo "PASS ✓" || echo "FAIL ✗")
成功ステップ: $steps_passed/$steps_total ($success_rate%)

================================================================================
テストステップ詳細
================================================================================

1. 前提条件チェック
   - Zeus バイナリ存在確認
   - ダッシュボードビルド確認
   - 必須コマンド確認（jq, curl, agent-browser）

2. テストプロジェクトセットアップ
   - テンポラリディレクトリ作成
   - zeus init 実行
   - サンプルタスク追加

3. ダッシュボードサーバー起動
   - Go サーバーをバックグラウンド起動
   - ポート $DASHBOARD_PORT でリッスン
   - API Ready 待機

4. ブラウザセッション開始
   - agent-browser でページを開く
   - ?e2e パラメータで __ZEUS__ API を有効化

5. アプリケーション Ready 待機
   - window.__ZEUS__.isReady() で描画完了を待機
   - タイムアウト: $TIMEOUT_APP_READY 秒

6. グラフ状態取得
   - window.__ZEUS__.getGraphState() で状態をキャプチャ
   - JSON 形式で状態を取得

7. 状態アーティファクト保存
   - キャプチャした状態をアーティファクト保存

8. メトリクス収集
   - window.__VIEWER_METRICS__ から計測ログを収集

9. 構造比較実行
   - 実際の状態をゴールデンファイルと比較
   - ノード、エッジ、カウント値の検証

================================================================================
検証方式
================================================================================

座標除外の構造比較:
  - 除外フィールド: nodes[*].x, nodes[*].y, nodes[*].id, viewport
  - 検証対象: nodes[*].name, status, progress, edges

ゴールデンファイル:
  $golden_file

================================================================================
環境情報
================================================================================

Zeus バージョン: $(cd "${PROJECT_ROOT}" && ./zeus --version 2>/dev/null || echo 'unknown')
Git ブランチ: $(git -C "${PROJECT_ROOT}" rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')
Git コミット: $(git -C "${PROJECT_ROOT}" rev-parse --short HEAD 2>/dev/null || echo 'unknown')
ダッシュボードポート: $DASHBOARD_PORT
アーティファクト: $ARTIFACTS_DIR

================================================================================
EOF

    log_info "テキストレポート生成: $report_file"
}

# 主要レポート生成関数（複数形式対応）
generate_test_reports() {
    local output_dir="${1:-.}"
    local test_start="$2"
    local test_end="$3"
    local steps_passed="$4"
    local steps_total="$5"
    local golden_file="${6:-${GOLDEN_DIR}/state/basic-tasks.json}"

    mkdir -p "$output_dir"

    # 形式別に生成
    generate_markdown_report "$output_dir/report.md" "$test_start" "$test_end" "$steps_passed" "$steps_total" "$golden_file"
    generate_html_report "$output_dir/report.html" "$test_start" "$test_end" "$steps_passed" "$steps_total"
    generate_text_report "$output_dir/report.txt" "$test_start" "$test_end" "$steps_passed" "$steps_total" "$golden_file"

    log_info "全形式レポート生成完了: $output_dir/"
}

export -f generate_markdown_report generate_html_report generate_text_report generate_test_reports
