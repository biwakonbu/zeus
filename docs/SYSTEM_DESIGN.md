# Zeus システム設計書

## 1. システム概要

### 1.1 ビジョン
Zeusは、AIによるプロジェクトマネジメントを「神の視点」で俯瞰し、上流工程（方針立案からWBS化、タイムライン設計、仕様作成まで）を支援するCLIベースのシステムです。

### 1.2 コアコンセプト
- **神の視点（Zeus View）**: プロジェクト全体を俯瞰し、依存関係、リソース配分、進捗を一元的に把握
- **ファイルベース**: 依存ミドルウェアゼロ、YAMLで可読性とGit親和性を確保
- **人間中心**: AIは助言者、人間が最終決定者
- **段階的複雑化**: Simple → Standard → Advancedの3段階構成

### 1.3 対象ユーザー
1. プロジェクトマネージャー
2. 技術リーダー/アーキテクト
3. プロダクトマネージャー
4. 経営層/ステークホルダー

## 2. アーキテクチャ

### 2.1 システム構成図

```
┌─────────────────────────────────────────────────────────────┐
│                         User Interface                       │
│                    ┌─────────────────────┐                  │
│                    │    Zeus CLI          │                  │
│                    └──────────┬──────────┘                  │
├───────────────────────────────┼─────────────────────────────┤
│                         Core Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Command     │  │  Approval   │  │  AI Engine  │        │
│  │  Processor   │  │  Manager    │  │  Interface  │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│         └────────────┬────┴────────────────┘                │
│                      │                                       │
│  ┌─────────────────────────────────────────────────┐       │
│  │               State Manager                       │       │
│  └─────────────────────────────────────────────────┘       │
├───────────────────────────────┼─────────────────────────────┤
│                         Data Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  zeus.yaml   │  │  Tasks      │  │  State      │        │
│  │  (Core)      │  │  Store      │  │  Snapshots  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Entities    │  │  Approvals  │  │  Analytics  │        │
│  │  (Standard)  │  │  Queue      │  │  Tracking   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 ディレクトリ構造

#### 2.2.1 Core（必須）
```
.zeus/
├── zeus.yaml        # メインプロジェクト定義
├── tasks/           # タスク管理
├── state/           # 状態管理
└── backups/         # 自動バックアップ
```

#### 2.2.2 Standard（標準）
```
.zeus/
├── config/          # 設定ファイル
├── entities/        # エンティティ定義
├── approvals/       # 承認管理
├── logs/            # ログ記録
└── analytics/       # 分析データ
```

#### 2.2.3 Advanced（高度）
```
.zeus/
├── graph/           # 関係性グラフ
├── views/           # カスタムビュー
└── .local/          # ローカル設定
```

### 2.3 データフォーマット

#### 2.3.1 メイン設定（zeus.yaml）
```yaml
# .zeus/zeus.yaml
version: "1.0"
project:
  id: "project-zeus-001"
  name: "Zeus Development"
  description: "AI-driven project management system"
  start_date: "2024-01-14"

objectives:
  - id: "obj-001"
    title: "MVP Release"
    deadline: "2024-03-31"
    priority: "high"

settings:
  automation_level: "standard"  # simple|standard|advanced
  approval_mode: "default"      # default|strict|loose
  ai_provider: "claude-code"    # claude-code|gemini|codex
```

#### 2.3.2 タスク定義
```yaml
# .zeus/tasks/active.yaml
tasks:
  - id: "task-001"
    title: "Design core data structure"
    status: "in_progress"
    assignee: "ai"
    estimate_hours: 8
    actual_hours: null
    dependencies: []
    approval_level: "auto"
```

#### 2.3.3 状態スナップショット
```yaml
# .zeus/state/current.yaml
snapshot:
  timestamp: "2024-01-14T16:00:00Z"
  summary:
    total_tasks: 42
    completed: 15
    in_progress: 10
    pending: 17
  health: "good"
  risks: []
```

## 3. 機能仕様

### 3.1 コアコマンド

#### 3.1.1 初期化
```bash
zeus init [--level=simple|standard|advanced]
```
- プロジェクトディレクトリを初期化
- レベルに応じた構造を生成

#### 3.1.2 状態確認
```bash
zeus status [--detail] [--format=text|json|html]
```
- プロジェクト全体の状態を表示
- 3層ビュー（Quick/Detailed/Rich）に対応

#### 3.1.3 エンティティ操作
```bash
zeus add <entity> <name>
zeus update <entity-id> <field> <value>
zeus list [entity] [--filter]
```

### 3.2 AI機能

#### 3.2.1 提案システム
```bash
zeus suggest                  # AI提案一覧
zeus suggest --detail <id>    # 提案詳細
zeus apply <id> [--dry-run]   # 提案適用
```

#### 3.2.2 説明機能
```bash
zeus explain <entity-id>      # AIによる解説
```

### 3.3 承認フロー

#### 3.3.1 承認レベル
1. **auto**: 自動実行（読み取り、計算、レポート生成）
2. **notify**: 通知付き実行（ステータス更新、軽微な変更）
3. **approve**: 事前承認必須（重要な変更、外部連携）

#### 3.3.2 承認コマンド
```bash
zeus pending              # 承認待ち一覧
zeus approve <id>         # 承認
zeus reject <id> [reason] # 却下
```

### 3.4 フィードバックシステム

#### 3.4.1 自動追跡
- タスク完了時の見積もり精度
- オーバーライド時の差分
- 承認/却下の結果

#### 3.4.2 明示的フィードバック
```bash
zeus ok <id>              # 成功報告
zeus ng <id> [reason]     # 失敗報告
zeus review               # 週次レビュー
zeus stats                # 精度統計
```

### 3.5 エラー処理と復旧

#### 3.5.1 診断と修復
```bash
zeus doctor               # システム診断
zeus fix [--dry-run]      # 自動修復
```

#### 3.5.2 バックアップと復元
```bash
zeus restore [point]      # バックアップ復元
zeus restore --latest     # 最新から復元
```

## 4. CLIインターフェース設計

### 4.1 コマンド体系
- 動詞＋名詞の一貫した構造
- 短縮形とエイリアスのサポート
- --dry-runオプションの統一的提供

### 4.2 出力形式
- デフォルト: 人間可読なテキスト
- --format=json: プログラム連携用
- --format=html: レポート出力用

### 4.3 対話的操作
- 危険な操作には確認プロンプト
- zeus reviewでの対話的フィードバック
- $EDITORを使用した直接編集

## 5. Agent Skills設計

### 5.1 Zeus Agent
Claude Code Pluginとして実装し、以下のスキルを提供：

1. **project-scan**: プロジェクト構造の自動認識
2. **task-suggest**: タスク分割と優先順位提案
3. **risk-analysis**: リスク分析と対策提案
4. **timeline-optimize**: タイムライン最適化

### 5.2 実装パターン
```javascript
// skills/project-scan/skill.js
export async function projectScan(context) {
  // プロジェクトファイルを探索
  // 依存関係を分析
  // zeus.yamlを生成
}
```

## 6. 拡張性設計

### 6.1 プラグインシステム
- .zeus/plugins/ディレクトリでカスタムスキル定義
- JavaScript/TypeScriptでの実装
- 標準APIの提供

### 6.2 外部連携（将来）
- Git統合（Phase 2）
- Slack/Email通知（Phase 2）
- 他のCLIツールとの連携（Phase 3）

## 7. セキュリティとプライバシー

### 7.1 データ保護
- ローカルファイルのみ使用
- 外部送信は明示的承認が必要
- .localディレクトリはgitignore推奨

### 7.2 アクセス制御
- ファイルシステムの権限に依存
- 承認履歴の完全記録
- オーバーライドの監査証跡

## 8. パフォーマンス目標

### 8.1 レスポンスタイム
- zeus status: < 100ms
- zeus suggest: < 3s
- zeus doctor: < 5s

### 8.2 スケーラビリティ
- タスク数: ~10,000まで対応
- 履歴保持: 90日分
- ファイルサイズ: 各YAMLは10MB以下

## 9. 実装優先順位

### 9.1 Phase 1（MVP）- 完了
1. zeus.yaml構造定義
2. 基本CLIコマンド（init, status, add, list）
3. doctor/fix機能

### 9.2 Phase 2（標準機能）- 完了
1. 3段階承認レベル（pending, approve, reject）
2. スナップショット・履歴管理
3. セキュリティ強化（パス検証、UUID、ファイルロック）
4. DI/Context対応（テスタビリティ向上）

### 9.3 Phase 2.7（提案機能）- 完了
1. suggest コマンド（ルールベース提案）
2. apply コマンド（提案適用）
3. 影響度フィルタリング

### 9.4 Phase 3（AI統合）- 未実装
1. Claude Code 連携設計・実装
2. AI ベースの提案機能
3. explain コマンド
4. Add コマンドと承認フローの連携
5. 依存関係グラフ
6. 予測分析

## 10. 設計決定の根拠

### 10.1 ファイルベースを選択した理由
- 依存関係の最小化
- Git親和性の確保
- 可読性とデバッグの容易さ
- バックアップとリカバリの簡単さ

### 10.2 YAMLを選択した理由
- 人間可読性の高さ
- 構造化データの表現力
- コメント機能
- 広範なツールサポート

### 10.3 CLIを選択した理由
- 開発者フレンドリー
- 自動化との親和性
- リモート操作の容易さ
- Claude Code等との統合

---

*Zeus System Design Document v1.0*
*作成日: 2026-01-14*