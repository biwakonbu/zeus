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
# 基本的な初期化
zeus init

# 詳細なセットアップ
zeus init --level=standard
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
zeus review                    # 対話的レビュー
zeus report --format=html      # レポート生成
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
# マネージャービュー（デフォルト）
zeus view manager

# エグゼクティブサマリー
zeus view executive

# 詳細ビュー
zeus status --detail
```

### 2.2 技術リーダー/アーキテクト向け

#### タスク分析
```bash
# タスク構造の確認
zeus list tasks --with-deps    # 依存関係付き
zeus explain <task-id>         # AI解説

# リスク分析
zeus health                    # プロジェクト健全性
```

#### 直接編集
```bash
# YAMLファイルを直接編集
zeus edit tasks                # $EDITORで編集
zeus edit config               # 設定を編集
```

#### 技術的調整
```bash
# 見積もり更新
zeus update <task-id> estimate_hours 16

# 依存関係追加
zeus add dependency <task-id> --depends-on <other-id>
```

### 2.3 プロダクトマネージャー向け

#### 目標管理
```bash
# 目標の確認
zeus list objectives

# 新規目標追加
zeus add objective "Q2 MVP Release" --deadline 2024-06-30

# 進捗確認
zeus status --objectives
```

#### 優先順位調整
```bash
# AI提案の確認
zeus suggest --category=priorities

# 優先順位変更
zeus update <task-id> priority high
```

### 2.4 経営層/ステークホルダー向け

#### ダッシュボード
```bash
# エグゼクティブビュー
zeus view executive

# 簡潔なステータス
zeus status
```

#### レポート
```bash
# HTMLレポート生成
zeus report --format=html

# JSON出力（BI連携用）
zeus report --format=json > report.json
```

## 3. コマンドリファレンス

### 3.1 Core コマンド

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus init` | プロジェクト初期化 | `--level=simple\|standard\|advanced` |
| `zeus status` | ステータス表示 | `--detail`, `--format=text\|json\|html` |
| `zeus scan` | プロジェクトスキャン | - |
| `zeus add` | エンティティ追加 | `<entity> <name>` |
| `zeus update` | エンティティ更新 | `<id> <field> <value>` |
| `zeus list` | 一覧表示 | `[entity] [--filter]` |
| `zeus report` | レポート生成 | `--format=md\|json\|html` |

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
| `zeus edit` | 直接編集 | `<entity>` |
| `zeus rollback` | 取り消し | `<override-id>` |

### 3.4 フィードバックコマンド

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus ok` | 成功報告 | `<id>` |
| `zeus ng` | 失敗報告 | `<id> [--reason "..."]` |
| `zeus review` | 週次レビュー | - |
| `zeus stats` | 精度統計 | `--detail`, `--json` |

### 3.5 復旧コマンド

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus doctor` | システム診断 | - |
| `zeus fix` | 自動修復 | `--dry-run` |
| `zeus restore` | バックアップ復元 | `[point]`, `--latest` |
| `zeus resume` | 通常モード復帰 | - |

### 3.6 自動化コマンド

| コマンド | 説明 | オプション |
|---------|------|-----------|
| `zeus automation status` | 自動化状態 | - |
| `zeus automation pause` | 一時停止 | - |
| `zeus automation resume` | 再開 | - |

## 4. ワークフロー

### 4.1 標準的な1日の流れ

```
Morning Check (朝)
├── zeus status          # 全体状況確認
├── zeus pending         # 承認待ち確認
└── zeus suggest         # AI提案確認

Work Session (作業中)
├── zeus approve/reject  # 提案への対応
├── zeus update          # タスク更新
└── zeus add             # 新規項目追加

End of Day (終業時)
├── zeus ok/ng           # フィードバック
├── zeus status          # 最終確認
└── zeus report          # 日次レポート（オプション）
```

### 4.2 週次レビューフロー

```bash
# Step 1: AI精度の確認
zeus stats

# Step 2: 対話的レビュー
zeus review

# Step 3: レポート生成
zeus report --format=html > weekly_report.html
```

### 4.3 問題発生時のフロー

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

## 5. 承認レベルの理解

### 5.1 auto（自動実行）
人間の確認なしで実行される操作:
- 読み取り操作
- 計算処理
- レポート生成
- 完了タスクのアーカイブ

### 5.2 notify（通知付き実行）
実行後に通知される操作:
- ステータス更新
- 見積もり更新（20%以内の変更）
- 依存関係追加

### 5.3 approve（事前承認必須）
実行前に承認が必要な操作:
- マイルストーン変更
- リソースアサイン
- スコープ変更
- 3タスク以上に影響する変更
- 信頼度70%未満のAI提案

## 6. トラブルシューティング

### 6.1 よくある問題と解決方法

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

### 6.2 グレースフルデグラデーション

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

# 通常モードへ復帰
zeus resume
```

## 7. ベストプラクティス

### 7.1 効果的な運用のコツ

1. **毎日のチェック習慣化**
   - 朝一番に `zeus status` と `zeus pending` を確認

2. **フィードバックを忘れずに**
   - タスク完了時に `zeus ok` / `zeus ng` を実行
   - これによりAIの精度が向上

3. **週次レビューの実施**
   - 毎週 `zeus review` で対話的レビュー
   - 問題点の早期発見

4. **直接編集の活用**
   - 複雑な変更は `zeus edit` で直接YAML編集
   - Gitでの差分管理が容易

### 7.2 避けるべきこと

1. **長期間の承認放置**
   - 7日以上放置すると警告が表示される

2. **フィードバックの省略**
   - AI精度向上の機会を逃す

3. **バックアップの削除**
   - `.zeus/backups/` は手動で削除しない

4. **YAMLの直接編集でのシンタックスエラー**
   - 編集後は `zeus doctor` で確認

## 8. 設定カスタマイズ

### 8.1 自動化ポリシーの調整
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

### 8.2 通知設定
```yaml
# .zeus/config/notifications.yaml
notifications:
  cli_output: true      # CLIへの出力
  log_file: true        # ログファイルへの記録
  # 将来の拡張
  # slack: false
  # email: false
```

### 8.3 ビュー設定
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

## 9. 用語集

| 用語 | 説明 |
|------|------|
| Objective | プロジェクトの目標・マイルストーン |
| Task | 具体的な作業項目 |
| Entity | Zeus管理対象（Task, Objective, Resource等） |
| Approval | 承認プロセス |
| Override | 人間によるAI提案の上書き |
| Snapshot | 特定時点の状態保存 |
| Health | プロジェクトの健全性指標 |

## 10. サポート

### 10.1 ヘルプの確認
```bash
zeus --help
zeus <command> --help
```

### 10.2 バージョン確認
```bash
zeus --version
```

### 10.3 問題報告
- GitHub Issues: https://github.com/biwakonbu/zeus/issues

---

*Zeus Operations Manual v1.0*
*作成日: 2026-01-14*