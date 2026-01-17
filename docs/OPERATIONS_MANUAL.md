# Zeus 運用マニュアル

## 1. クイックスタート

### 1.1 インストール
```bash
# Claude Code Pluginとしてインストール
claude plugin install zeus

# または、ローカルからインストール
claude plugin add ./zeus-plugin
```

### 1.2 プロジェクト初期化
```bash
# プロジェクト初期化（全機能を使用可能な状態で初期化）
zeus init
```

### 1.3 最初のステータス確認
```bash
zeus status
```

## 2. ユーザー別ガイド

### 2.1 プロジェクトマネージャー向け

#### 日常業務
```bash
# 朝のチェック
zeus status                    # プロジェクト状況確認
zeus pending                   # 承認待ちアイテム確認

# タスク管理
zeus list tasks --status=in_progress  # 進行中タスク
zeus suggest                   # AI提案を確認
zeus apply <id>                # 提案を適用

# 週次レビュー
zeus report --format=html      # レポート生成
zeus predict                   # 予測分析
```

#### 承認ワークフロー
```bash
# 承認待ちの確認
zeus pending

# 詳細確認
zeus suggest --detail <id>

# 承認/却下
zeus approve <id>
zeus reject <id> --reason "理由をここに記述"
```

#### ビュー切替
```bash
# 詳細ビュー
zeus status

# ダッシュボードで可視化
zeus dashboard
```

### 2.2 技術リーダー/アーキテクト向け

#### タスク分析
```bash
# タスク構造の確認
zeus list tasks                # タスク一覧
zeus explain <task-id>         # AI解説
zeus graph                     # 依存関係グラフ

# リスク分析
zeus predict risk              # リスク予測
zeus doctor                    # プロジェクト診断
```

#### YAML直接編集
```bash
# .zeus/tasks/active.yaml を直接編集
# 編集後は zeus doctor で構文チェック
```

### 2.3 プロダクトマネージャー向け

#### 目標管理
```bash
# タスク一覧の確認
zeus list tasks

# 新規タスク追加
zeus add task "新機能の実装" --due 2026-03-31

# 進捗確認
zeus status
```

#### 優先順位調整
```bash
# AI提案の確認
zeus suggest

# 提案を適用
zeus apply <suggestion-id>
```

### 2.4 経営層/ステークホルダー向け

#### ダッシュボード
```bash
# Web ダッシュボードで可視化
zeus dashboard

# 簡潔なステータス
zeus status
```

#### レポート
```bash
# HTMLレポート生成
zeus report --format=html -o report.html

# Markdown レポート
zeus report --format=markdown -o report.md
```

## 3. コマンドリファレンス

### 3.1 Core コマンド

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus init` | プロジェクト初期化 | - |
| `zeus status` | ステータス表示 | - |
| `zeus add` | エンティティ追加 | `<entity> <name>` |
| `zeus list` | 一覧表示 | `[entity] [--status]` |
| `zeus doctor` | システム診断 | - |
| `zeus fix` | 自動修復 | `--dry-run` |

### 3.2 AI コマンド

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus suggest` | AI提案一覧 | `--detail <id>`, `--category` |
| `zeus apply` | 提案適用 | `<id>`, `--dry-run` |
| `zeus explain` | AI解説 | `<entity-id>` |

### 3.3 承認コマンド

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus pending` | 承認待ち一覧 | - |
| `zeus approve` | 承認 | `<id>` |
| `zeus reject` | 却下 | `<id> [--reason ""]` |
| `zeus snapshot` | スナップショット管理 | `create\|list\|restore` |
| `zeus history` | 履歴表示 | `-n <limit>` |

### 3.4 分析コマンド（Phase 4）

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus graph` | 依存関係グラフ表示 | `--format text\|dot\|mermaid`, `-o <file>` |
| `zeus predict` | 予測分析 | `completion\|risk\|velocity\|all` |
| `zeus report` | レポート生成 | `--format text\|html\|markdown`, `-o <file>` |

### 3.5 ダッシュボードコマンド（Phase 5）

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus dashboard` | Web ダッシュボード起動 | `--port <port>`, `--no-open` |

### 3.6 ユーティリティコマンド

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus update-claude` | Claude Code 連携ファイル再生成 | - |

## 4. 分析機能の運用（Phase 4）

### 4.1 依存関係グラフ（graph コマンド）

#### 基本的な使い方
```bash
# テキスト形式で表示（CLI確認用）
zeus graph

# Graphviz DOT形式で出力
zeus graph --format dot -o dependencies.dot
dot -Tpng dependencies.dot -o dependencies.png

# Mermaid形式でMarkdownに出力
zeus graph --format mermaid -o docs/dependencies.md
```

#### 出力形式の選択ガイド

| 形式 | 用途 | 出力例 |
|------|------|--------|
| text | CLI での簡易確認 | `TASK-001 --> TASK-002` |
| dot | 画像生成（Graphviz） | digraph G {...} |
| mermaid | ドキュメント埋め込み | ```mermaid graph TD ...``` |

#### 循環依存の検出
```bash
# グラフに循環依存がある場合、警告が表示される
zeus graph
# 出力例:
# Warning: Circular dependency detected!
#   TASK-001 -> TASK-002 -> TASK-003 -> TASK-001
```

### 4.2 予測分析（predict コマンド）

#### 完了日予測
```bash
zeus predict completion
# 出力例:
# Completion Prediction
# =====================
# Estimated completion: 2024-03-15
# Confidence interval: 2024-03-10 ~ 2024-03-20
# Remaining tasks: 12
# Average velocity: 2.5 tasks/day
```

#### リスク分析
```bash
zeus predict risk
# 出力例:
# Risk Analysis
# =============
# Overall risk level: MEDIUM
#
# Risk factors:
#   [HIGH] Dependency complexity - 3 tasks have 5+ dependencies
#   [MEDIUM] Estimation accuracy - 30% of tasks exceeded estimates
#   [LOW] Scope creep - 2 new tasks added this week
```

#### ベロシティ分析
```bash
zeus predict velocity
# 出力例:
# Velocity Analysis
# =================
# Current velocity: 2.5 tasks/day
# Trend: +0.3 (improving)
# 7-day average: 2.2 tasks/day
# 30-day average: 2.0 tasks/day
```

#### 全予測を一度に表示
```bash
zeus predict
# または
zeus predict all
```

### 4.3 レポート生成（report コマンド）

#### テキストレポート
```bash
zeus report
# プロジェクト概要、進捗、タスク一覧を標準出力に表示
```

#### HTMLレポート
```bash
# HTMLファイルとして出力
zeus report --format html -o report.html

# ブラウザで確認
open report.html
```

#### Markdownレポート
```bash
# Markdownファイルとして出力
zeus report --format markdown -o docs/STATUS.md

# GitHubなどで表示
```

#### レポート内容
1. プロジェクト概要（名前、説明、開始日）
2. 進捗サマリー（完了率、残タスク数）
3. タスク一覧（ステータス別）
4. 依存関係グラフ（Mermaid形式、HTML/Markdownのみ）
5. 予測分析結果
6. リスク・課題

## 5. Webダッシュボードの運用（Phase 5）

### 5.1 ダッシュボードの起動

#### 基本起動
```bash
# デフォルトポート(8080)で起動し、ブラウザを自動で開く
zeus dashboard
```

#### カスタムポート
```bash
# ポート3000で起動
zeus dashboard --port 3000
```

#### ブラウザ自動起動の無効化
```bash
# ヘッドレスサーバーとして起動（CIなど）
zeus dashboard --no-open
```

### 5.2 ダッシュボードの停止

```bash
# Ctrl+C でサーバーを停止
# または、別ターミナルから
kill $(lsof -t -i:8080)
```

### 5.3 アクセス方法

ダッシュボードは **ローカルホストのみ** からアクセス可能です:
- URL: `http://localhost:8080` （または指定ポート）
- 外部ネットワークからのアクセスは不可（セキュリティ対策）

### 5.4 ダッシュボード機能

| 機能 | 説明 |
|------|------|
| プロジェクト概要 | 名前、説明、進捗率、健全性をカード表示 |
| タスク統計 | 完了/進行中/保留の内訳 |
| タスク一覧 | テーブル形式、ステータス色分け |
| 依存関係グラフ | Mermaid.js でインタラクティブ表示 |
| 予測分析 | 完了日、リスク、ベロシティ |
| 自動更新 | 5秒間隔で最新データを取得 |

### 5.5 REST API の利用

プログラムからダッシュボードAPIを利用:

```bash
# プロジェクト状態を取得
curl http://localhost:8080/api/status | jq

# タスク一覧を取得
curl http://localhost:8080/api/tasks | jq

# 依存関係グラフ（Mermaid形式）を取得
curl http://localhost:8080/api/graph

# 予測分析結果を取得
curl http://localhost:8080/api/predict | jq
```

### 5.6 ダッシュボード計測ログ（Graph View）

Graph View の描画/更新メトリクスを自動保存できます。

- 開発時: `http://localhost:5173/?metricsAutoSave`
- 本番時: `http://localhost:8080/?metricsAutoSave`
- テストモード（`import.meta.env.MODE === 'test'`）では自動記録が常時有効

保存先:

- `.zeus/metrics/dashboard-metrics-<session>.jsonl`（JSON Lines）

補足:

- 手動ダウンロードは `?metrics` で有効化後、Graph View 右上の `DL` ボタン
- 自動保存は `/api/metrics` に送信され、サーバー側で追記保存される

### 5.7 トラブルシューティング

#### ポートが使用中の場合
```bash
# 別のポートを指定
zeus dashboard --port 3000

# または、使用中のプロセスを終了
lsof -i:8080
kill <PID>
```

#### ブラウザが開かない場合
```bash
# 手動でブラウザを開く
zeus dashboard --no-open
# 別ターミナルで
open http://localhost:8080
```

## 6. ワークフロー

### 6.1 標準的な1日の流れ

```
Morning Check (朝)
├── zeus status          # 全体状況確認
├── zeus pending         # 承認待ち確認
└── zeus suggest         # AI提案確認

Work Session (作業中)
├── zeus approve/reject  # 提案への対応
├── zeus add task        # タスク追加
└── zeus apply           # AI提案の適用

End of Day (終業時)
├── zeus status          # 最終確認
└── zeus report          # 日次レポート（オプション）
```

### 6.2 週次レビューフロー

```bash
# Step 1: 依存関係の確認
zeus graph --format mermaid

# Step 2: 予測分析の確認
zeus predict

# Step 3: レポート生成
zeus report --format=html -o weekly_report.html
```

### 6.3 分析ワークフロー

```bash
# Step 1: 依存関係の確認
zeus graph --format mermaid

# Step 2: 循環依存のチェック
zeus graph | grep -i "circular"

# Step 3: 予測の確認
zeus predict

# Step 4: レポート生成
zeus report --format html -o analysis_report.html
```

### 6.4 ダッシュボードを使った運用

```bash
# Step 1: ダッシュボードを起動（バックグラウンド）
zeus dashboard &

# Step 2: ブラウザで確認しながら作業
#   - プロジェクト概要を確認
#   - タスク進捗を監視
#   - 依存関係グラフでボトルネックを特定

# Step 3: 作業完了後に停止
fg
# Ctrl+C
```

### 6.5 問題発生時のフロー

```bash
# Step 1: 診断
zeus doctor

# Step 2: 問題確認
# → 修復可能な問題が表示される

# Step 3: 修復（プレビュー）
zeus fix --dry-run

# Step 4: 修復実行
zeus fix

# Step 5: 状態確認
zeus status
```

## 7. 承認レベルの理解

### 7.1 auto（自動実行）
人間の確認なしで実行される操作:
- 読み取り操作
- 計算処理
- レポート生成
- 完了タスクのアーカイブ

### 7.2 notify（通知付き実行）
実行後に通知される操作:
- ステータス更新
- 見積もり更新（20%以内の変更）
- 依存関係追加

### 7.3 approve（事前承認必須）
実行前に承認が必要な操作:
- マイルストーン変更
- リソースアサイン
- スコープ変更
- 3タスク以上に影響する変更
- 信頼度70%未満のAI提案

## 8. トラブルシューティング

### 8.1 よくある問題と解決方法

#### zeus.yaml が見つからない
```bash
# 原因: 初期化されていない
# 解決:
zeus init
```

#### YAMLシンタックスエラー
```bash
# 診断
zeus doctor

# 結果例:
# [FAIL] yaml_syntax: 2 YAML syntax errors found
#   - tasks/active.yaml: invalid indentation

# 解決: エディタで修正
zeus edit tasks
```

#### 状態の不整合
```bash
# 診断
zeus doctor

# 結果例:
# [WARN] state_integrity: State may be out of sync

# 解決
zeus fix
```

#### バックアップがない
```bash
# 診断
zeus doctor

# 結果例:
# [WARN] backup_health: No recent backups found

# 解決
zeus fix  # 自動でバックアップを作成
```

### 8.2 グレースフルデグラデーション

Zeusは問題発生時に段階的に機能を制限します：

| レベル | 状態 | 利用可能機能 |
|--------|------|-------------|
| Normal | 正常 | 全機能 |
| Limited | 一部制限 | AI提案停止、読み書き可能 |
| Read-only | 読み取り専用 | 閲覧のみ |
| Safe | セーフモード | 復旧機能のみ |

```bash
# 現在のモード確認
zeus status

# 問題の診断と修復
zeus doctor
zeus fix
```

## 9. ベストプラクティス

### 9.1 効果的な運用のコツ

1. **毎日のチェック習慣化**
   - 朝一番に `zeus status` と `zeus pending` を確認

2. **AI提案の活用**
   - `zeus suggest` で改善提案を取得
   - `zeus apply` で提案を適用

3. **週次レビューの実施**
   - `zeus predict` で予測分析
   - `zeus report` でレポート生成

4. **直接編集の活用**
   - 複雑な変更は `.zeus/tasks/active.yaml` を直接編集
   - Gitでの差分管理が容易

5. **分析機能の活用**
   - `zeus graph` で依存関係を可視化
   - `zeus predict` でリスクを早期発見
   - `zeus dashboard` でリアルタイム監視

### 9.2 避けるべきこと

1. **長期間の承認放置**
   - 7日以上放置すると警告が表示される

2. **バックアップの削除**
   - `.zeus/backups/` は手動で削除しない

3. **YAMLの直接編集でのシンタックスエラー**
   - 編集後は `zeus doctor` で確認

4. **ダッシュボードの外部公開**
   - セキュリティリスクがあるため、常にローカルアクセスのみ

## 10. 設定カスタマイズ

### 10.1 自動化ポリシーの調整
```yaml
# .zeus/config/automation.yaml
automation:
  overrides:
    # notifyをautoに昇格
    promote_to_auto:
      - "update_status"

    # notifyをapproveに降格
    demote_to_approve:
      - "add_dependency"
```

### 10.2 通知設定
```yaml
# .zeus/config/notifications.yaml
notifications:
  cli_output: true      # CLIへの出力
  log_file: true        # ログファイルへの記録
  # 将来の拡張
  # slack: false
  # email: false
```

### 10.3 ビュー設定
```yaml
# .zeus/config/views.yaml
views:
  default: "manager"
  quick:
    show_risks: false
    max_tasks: 5
  detailed:
    show_history: true
    depth: 3
```

### 10.4 ダッシュボード設定
```yaml
# .zeus/config/dashboard.yaml
dashboard:
  default_port: 8080
  auto_open: true
  refresh_interval: 5000  # ミリ秒
```

## 11. 用語集

| 用語 | 説明 |
|------|------|
| Objective | プロジェクトの目標・マイルストーン |
| Task | 具体的な作業項目 |
| Entity | Zeus管理対象（Task, Objective, Resource等） |
| Approval | 承認プロセス |
| Override | 人間によるAI提案の上書き |
| Snapshot | 特定時点の状態保存 |
| Health | プロジェクトの健全性指標 |
| Graph | 依存関係グラフ |
| Velocity | タスク完了速度 |
| Dashboard | Webベースの管理画面 |

## 12. サポート

### 12.1 ヘルプの確認
```bash
zeus --help
zeus <command> --help
```

### 12.2 バージョン確認
```bash
zeus --version
```

### 12.3 問題報告
- GitHub Issues: https://github.com/biwakonbu/zeus/issues

---

*Zeus Operations Manual v1.2*
*作成日: 2026-01-14*
*更新日: 2026-01-17（未実装コマンドを削除、実装済み機能のみに整理）*
