# Zeus ユーザーガイド

## 1. はじめに

### 1.1 Zeus とは

Zeus は「神の視点」でプロジェクト管理を支援する AI 駆動型 CLI システムです。タスク管理、進捗追跡、AI による提案機能を提供し、プロジェクトの上流工程（方針立案から WBS 化、タイムライン設計まで）を支援します。

### 1.2 主な機能

- **ファイルベース管理**: 外部データベース不要、YAML ファイルで人間可読
- **3段階承認フロー**: auto / notify / approve の承認レベル
- **スナップショット**: プロジェクト状態の保存と復元
- **AI 提案**: タスク追加、優先度変更、リスク対策の提案
- **予測分析**: 完了日予測、リスク予測、ベロシティ分析
- **依存関係グラフ**: タスク間の依存関係を可視化
- **Web ダッシュボード**: ブラウザでプロジェクト状態をリアルタイム可視化

### 1.3 対象ユーザー

- プロジェクトマネージャー
- 技術リーダー / アーキテクト
- プロダクトマネージャー
- 個人開発者

## 2. インストール

### 2.1 前提条件

- Go 1.21 以上
- Git（ソースコードの取得用）

### 2.2 ソースからのビルド

```bash
# リポジトリをクローン
git clone https://github.com/biwakonbu/zeus.git
cd zeus

# ビルド
make build

# インストール（$GOPATH/bin にコピー）
make install
```

### 2.3 動作確認

```bash
# バージョン確認
zeus --help
```

出力例:
```
Zeus は AI によるプロジェクトマネジメントを「神の視点」で
俯瞰するシステムです。上流工程（方針立案からWBS化、タイムライン設計、
仕様作成まで）を支援します。

Usage:
  zeus [command]

Available Commands:
  add         エンティティを追加
  apply       AI提案を適用
  approve     アイテムを承認
  doctor      システムの健全性を診断
  ...
```

## 3. クイックスタート

### 3.1 5分で始める Zeus

```bash
# 1. プロジェクトディレクトリに移動
cd your-project

# 2. Zeus を初期化
zeus init

# 3. タスクを追加
zeus add task "ドキュメント作成"

# 4. 状態を確認
zeus status
```

### 3.2 出力例

```
Zeus Project Status
═══════════════════════════════════════════════════════════
Project: New Zeus Project
Health:  good

Tasks Summary:
  Total:       1
  Completed:   0
  In Progress: 0
  Pending:     1
═══════════════════════════════════════════════════════════
```

## 4. 基本的な使い方

### 4.1 プロジェクト初期化

`zeus init` コマンドでプロジェクトを初期化します。

```bash
# プロジェクトを初期化
zeus init
```

初期化により以下が作成されます:

- `.zeus/` ディレクトリ（タスク、状態、承認管理用）
- `.claude/` ディレクトリ（Claude Code 連携用）
- デフォルトの `automation_level` は `auto`（即時実行、承認不要）

**承認フローの設定:**

承認フローは `zeus.yaml` の `automation_level` で設定できます:

| 値 | 説明 | 動作 |
|---|------|------|
| auto | 自動承認（デフォルト） | 即時実行、承認不要 |
| notify | 通知のみ | 実行時に通知、ログ記録 |
| approve | 事前承認必須 | 承認待ちキューに追加 |

### 4.2 状態確認

`zeus status` コマンドでプロジェクトの状態を確認できます。

```bash
# 基本的な状態表示
zeus status

# 詳細表示
zeus status --detail
```

### 4.3 タスク管理

#### タスクの追加

```bash
zeus add task "新機能の実装"
```

出力例:
```
✓ Added task: 新機能の実装 (ID: task-abc123)
```

Note: `automation_level` が `approve` の場合、タスク追加時に承認フローが適用されます。

#### タスク一覧の表示

```bash
# 全タスクを表示
zeus list

# タスクのみ表示
zeus list task
```

出力例:
```
Tasks (3 items)
────────────────────────────────────────
[pending] task-001 - ドキュメント作成
[in_progress] task-002 - 新機能の実装
[completed] task-003 - バグ修正
```

### 4.4 診断と修復

#### システム診断

```bash
zeus doctor
```

出力例:
```
Zeus Doctor - System Diagnosis
═══════════════════════════════════════════════════════════
✓ zeus_yaml: zeus.yaml exists and is valid
✓ tasks_dir: tasks directory exists
⚠ backup_health: No recent backups found
═══════════════════════════════════════════════════════════
Overall: fair

1 issue(s) can be fixed automatically. Run 'zeus fix' to repair.
```

#### 自動修復

```bash
# プレビュー（実行せずに確認）
zeus fix --dry-run

# 実行
zeus fix
```

## 5. 承認フロー

### 5.1 承認レベル

Zeus は3段階の承認レベルをサポートしています。

| レベル | 説明 | 適用例 |
|--------|------|--------|
| auto | 自動承認（即時実行） | 読み取り操作、レポート生成 |
| notify | 通知付き実行 | ステータス更新、軽微な変更 |
| approve | 事前承認必須 | 重要な変更、スコープ変更 |

### 5.2 承認待ちの管理

```bash
# 承認待ちアイテムを表示
zeus pending
```

出力例:
```
Pending Approvals
═══════════════════════════════════════════════════════════
[approve] appr-001 - タスク追加: 新機能の実装
    Type: add_task | Created: 2026-01-15T10:00:00Z
═══════════════════════════════════════════════════════════
Total: 1 item(s)

Use 'zeus approve <id>' to approve or 'zeus reject <id>' to reject.
```

### 5.3 承認と却下

```bash
# 承認
zeus approve appr-001

# 却下（理由付き）
zeus reject appr-001 --reason "優先度が低いため"
```

## 6. スナップショットと履歴

### 6.1 スナップショット

プロジェクトの状態をスナップショットとして保存できます。

```bash
# スナップショット作成
zeus snapshot create

# ラベル付きで作成
zeus snapshot create "リリース前"

# スナップショット一覧
zeus snapshot list

# スナップショットから復元
zeus snapshot restore 2026-01-15T10:00:00Z
```

### 6.2 履歴

プロジェクトの履歴を確認できます。

```bash
# 履歴表示（デフォルト10件）
zeus history

# 件数指定
zeus history -n 5
```

出力例:
```
Project History
═══════════════════════════════════════════════════════════
1. 2026-01-15T10:00:00Z [リリース前]
   Health: good | Tasks: 10 (Completed: 8, In Progress: 1, Pending: 1)

2. 2026-01-14T18:00:00Z
   Health: fair | Tasks: 10 (Completed: 5, In Progress: 3, Pending: 2)
═══════════════════════════════════════════════════════════
Use 'zeus snapshot restore <timestamp>' to restore a snapshot.
```

## 7. AI機能

### 7.1 提案機能

Zeus は現在のプロジェクト状態を分析し、改善提案を生成します。

```bash
# 提案を生成
zeus suggest

# 件数制限
zeus suggest --limit 3

# 影響度でフィルタ
zeus suggest --impact high
```

出力例:
```
[SUGGESTIONS] 2件の提案が生成されました:

1. [high] テストタスクの追加を推奨
   理由: テストカバレッジが不足しています
   ID: sug-001
   見積: 4.0時間

2. [medium] 優先度の見直しを推奨
   理由: ブロックされているタスクがあります
   ID: sug-002

[HINT] 提案を適用するには: zeus apply <suggestion-id>
```

#### 提案の適用

```bash
# 個別適用
zeus apply sug-001

# 全て適用
zeus apply --all

# プレビュー
zeus apply --all --dry-run
```

### 7.2 説明機能

エンティティの詳細説明を生成します。

```bash
# プロジェクト全体の説明
zeus explain project

# 特定タスクの説明
zeus explain task-001

# コンテキスト情報を含む
zeus explain task-001 --context
```

### 7.3 予測分析

プロジェクトの予測分析を表示します。

```bash
# 全ての予測
zeus predict

# 完了日予測のみ
zeus predict completion

# リスク予測のみ
zeus predict risk

# ベロシティ分析のみ
zeus predict velocity
```

出力例:
```
Zeus Prediction Analysis
============================================================

[COMPLETION PREDICTION]
  Estimated Completion: 2026-02-15
  Margin:               +/- 5 days
  Average Velocity:     3.5 tasks/week
  Remaining Tasks:      12
  Confidence:           75%

[RISK PREDICTION]
  Overall Risk Level:   Medium
  Risk Score:           45/100

  Risk Factors:
    - 依存関係の複雑さ (Impact: 6/10)
      複数のタスクが相互依存しています

[VELOCITY ANALYSIS]
  Last 7 days:          4 tasks completed
  Last 14 days:         7 tasks completed
  Last 30 days:         15 tasks completed
  Weekly Average:       3.5 tasks
  Trend:                Stable
============================================================
```

### 7.4 依存関係グラフ

タスク間の依存関係を可視化します。

```bash
# テキスト形式で表示
zeus graph

# Graphviz DOT 形式
zeus graph --format=dot

# Mermaid 形式
zeus graph --format=mermaid

# ファイルに出力
zeus graph --format=mermaid --output=deps.md
```

### 7.5 レポート生成

プロジェクトの包括的なレポートを生成します。

```bash
# テキスト形式
zeus report

# HTML 形式
zeus report --format=html --output=report.html

# Markdown 形式
zeus report --format=markdown --output=report.md
```

### 7.6 Web ダッシュボード

ブラウザでプロジェクト状態をリアルタイム可視化します。

```bash
# デフォルト設定で起動（ポート8080、ブラウザ自動起動）
zeus dashboard

# ポート番号を指定
zeus dashboard --port 3000

# ブラウザを自動で開かない
zeus dashboard --no-open
```

出力例:
```
Zeus Dashboard
═══════════════════════════════════════════════════════════
Starting server on port 8080...

Dashboard: http://localhost:8080

Press Ctrl+C to stop the server
═══════════════════════════════════════════════════════════
```

**ダッシュボードの機能:**

- **プロジェクト概要**: プロジェクト名、健全性、進捗率を表示
- **タスク統計**: 完了/進行中/未着手/ブロック中のタスク数を集計
- **タスク一覧**: 全タスクをステータス別に表示
- **依存関係グラフ**: Mermaid.js を使用した依存関係の可視化
- **予測分析**: 完了日予測、リスク予測、ベロシティ分析を表示

**ユースケース:**

1. **チームミーティング**: ダッシュボードをプロジェクターで表示してプロジェクト状況を共有
2. **進捗確認**: 定期的にダッシュボードで全体の進捗を俯瞰
3. **依存関係の把握**: グラフでタスク間の依存関係を視覚的に確認
4. **リスク評価**: 予測分析で潜在的な問題を早期発見

## 8. ベストプラクティス

### 8.1 推奨ワークフロー

**毎日の作業:**
1. `zeus status` でプロジェクト状態を確認
2. `zeus pending` で承認待ちを確認
3. 作業完了時にタスクステータスを更新

**週次レビュー:**
1. `zeus predict` で進捗を分析
2. `zeus suggest` で改善提案を確認
3. `zeus snapshot create "週次バックアップ"` でスナップショット作成
4. `zeus dashboard` でチームと進捗を共有

### 8.2 Tips

- スナップショットは重要なマイルストーン前に作成する
- `zeus doctor` は定期的に実行して問題を早期発見する
- `--dry-run` オプションを活用して操作結果を事前確認する
- ダッシュボードはチームミーティングでの進捗共有に便利

### 8.3 避けるべきこと

- `.zeus/` ディレクトリ内のファイルを手動で編集しない（`zeus` コマンドを使用）
- バックアップディレクトリを削除しない
- 承認待ちアイテムを長期間放置しない

## 9. トラブルシューティング

### 9.1 よくある問題

#### 「zeus.yaml が見つからない」

**原因**: プロジェクトが初期化されていない

**解決方法**:
```bash
zeus init
```

#### 「YAML シンタックスエラー」

**原因**: YAML ファイルの形式が不正

**解決方法**:
```bash
# 診断
zeus doctor

# 修復
zeus fix
```

#### 「承認待ちアイテムが処理できない」

**原因**: 指定した ID が存在しない、または既に処理済み

**解決方法**:
```bash
# 承認待ち一覧を確認
zeus pending

# 正しい ID で再実行
zeus approve <正しいID>
```

#### 「ダッシュボードが起動しない」

**原因**: ポートが既に使用中

**解決方法**:
```bash
# 別のポートを指定
zeus dashboard --port 3001
```

### 9.2 エラーメッセージの読み方

Zeus のエラーメッセージは以下の形式で表示されます:

```
Error: <エラーの概要>
  <詳細情報>
```

詳細情報には、問題の原因と推奨される対処方法が含まれています。

### 9.3 サポート

問題が解決しない場合:

1. `zeus doctor` で診断結果を確認
2. GitHub Issues で報告: https://github.com/biwakonbu/zeus/issues

---

*Zeus User Guide v1.1*
*作成日: 2026-01-15*
*更新日: 2026-01-16（init コマンド簡略化）*
