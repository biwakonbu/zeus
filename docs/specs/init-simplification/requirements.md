# zeus init 簡素化 - 要件定義書

## 概要

### 目的

`zeus init` コマンドから `--level` オプションを廃止し、単一の初期化動作に統一する。

### 背景

- 3段階のレベル（simple/standard/advanced）は複雑さを増加させていた
- シンプルな UX を実現（デフォルトは即時実行、承認不要）
- 承認システムは `zeus.yaml` で後から設定変更可能にする

## 機能要件

### FR-001: init コマンドの簡素化

- `zeus init` は引数なしで実行可能
- `--level` オプションを削除
- 実行後、全機能が使用可能な状態で初期化される

### FR-002: ディレクトリ構造の統一

単一のディレクトリ構造を採用し、全機能（承認、スナップショット、分析等）に対応:

```
.zeus/
├── config/
├── tasks/
│   └── _archive/
├── state/
│   └── snapshots/
├── entities/
├── approvals/
│   ├── pending/
│   ├── approved/
│   └── rejected/
├── logs/
├── analytics/
└── backups/
```

### FR-003: デフォルト設定

| 設定項目 | デフォルト値 | 説明 |
|----------|--------------|------|
| automation_level | auto | 承認なし、即時実行 |
| approval_mode | default | 標準承認モード |
| ai_provider | claude-code | AI プロバイダー |

### FR-004: Claude Code 連携

`zeus init` 実行時に `.claude/` を常に生成:

- **エージェント**: orchestrator, planner, reviewer
- **スキル**: project-scan, task-suggest, risk-analysis

### FR-005: 承認システムの独立性

- 承認機能は `zeus.yaml` の `automation_level` で制御
- 設定値: `auto` | `notify` | `approve`
- ユーザーが手動で変更可能

## 非機能要件

### NFR-001: 後方互換性

- 既存の `.zeus/` ディレクトリを持つプロジェクトに影響なし
- 既存の `zeus.yaml` 設定は維持される

### NFR-002: パフォーマンス

- 初期化処理: 1秒以内
- ディレクトリ作成とファイル生成のみ

### NFR-003: エラーハンドリング

- 既に初期化済みの場合はエラーメッセージを表示
- ディスク容量不足等のシステムエラーを適切に処理

## インターフェース変更

### コマンドライン

```bash
# 変更前
zeus init [--level=simple|standard|advanced]

# 変更後
zeus init
```

### Go API

```go
// 変更前
func (z *Zeus) Init(ctx context.Context, level string) (*InitResult, error)

// 変更後
func (z *Zeus) Init(ctx context.Context) (*InitResult, error)
```

### InitResult 構造体

```go
// 変更前
type InitResult struct {
    Success    bool
    Level      string
    ZeusPath   string
    ClaudePath string
}

// 変更後
type InitResult struct {
    Success    bool
    ZeusPath   string
    ClaudePath string
}
```

## 受け入れ基準

| ID | 基準 |
|----|------|
| AC-001 | `zeus init` を引数なしで実行できること |
| AC-002 | 初期化後、`zeus status` が正常に動作すること |
| AC-003 | 初期化後、`zeus add task <name>` が即時実行されること（承認なし） |
| AC-004 | 初期化後、`.claude/` ディレクトリにエージェントとスキルが生成されていること |
| AC-005 | `zeus.yaml` の `automation_level` を `approve` に変更後、`zeus add task` で承認待ちキューに追加されること |
| AC-006 | 全テストがパスすること |
