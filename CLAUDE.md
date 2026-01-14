# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Zeus は「神の視点」でプロジェクト管理を支援する AI 駆動型 CLI システム。Go + Cobra で実装。

**設計哲学:**
- ファイルベース: 外部 DB 不要、YAML で人間可読
- 人間中心: AI は提案者、人間が最終決定者
- 段階的複雑化: Simple → Standard → Advanced の3レベル
- Git 親和性: 全データがテキストで差分追跡可能

## 技術スタック

- **言語**: Go 1.21+
- **CLI フレームワーク**: Cobra
- **YAML 処理**: gopkg.in/yaml.v3
- **カラー出力**: fatih/color
- **UUID 生成**: github.com/google/uuid
- **配布形式**: スタンドアロン CLI + Claude Code Plugin 連携

## 開発コマンド

```bash
# ビルド
make build

# テスト実行
make test

# 単一パッケージテスト
go test -v ./internal/core/...

# 開発実行
go run . <command>
make dev ARGS="init --level=simple"

# インストール
make install
```

## プロジェクト構造

```
zeus/
├── cmd/                      # Cobra コマンド
│   ├── root.go               # ルートコマンド（DI対応）
│   ├── init.go               # zeus init
│   ├── status.go             # zeus status
│   ├── pending.go            # zeus pending
│   ├── approve.go            # zeus approve
│   ├── reject.go             # zeus reject
│   ├── snapshot.go           # zeus snapshot
│   ├── history.go            # zeus history
│   └── ...
├── internal/                 # 内部パッケージ
│   ├── core/                 # コアロジック
│   │   ├── zeus.go           # メインロジック（DI対応）
│   │   ├── interfaces.go     # FileStore, StateStore, ApprovalStore インターフェース
│   │   ├── entity.go         # EntityHandler, EntityRegistry
│   │   ├── task_handler.go   # タスクエンティティハンドラー
│   │   ├── types.go          # 型定義
│   │   ├── state.go          # 状態・スナップショット管理（Context対応）
│   │   ├── approval.go       # 3段階承認システム（Context対応）
│   │   ├── errors.go         # エラー定義
│   │   └── mocks/            # テスト用モック
│   │       └── mock_file_store.go
│   ├── yaml/                 # YAML 操作
│   │   ├── parser.go
│   │   ├── writer.go
│   │   ├── file_manager.go   # パス検証・セキュリティ（Context対応）
│   │   └── filelock.go       # ファイルロック機構
│   ├── doctor/               # 診断・修復（Context対応）
│   │   └── doctor.go
│   └── generator/            # Claude Code 連携ファイル生成（Context対応）
│       └── generator.go
├── main.go
├── go.mod
└── Makefile
```

## アーキテクチャ

### コアモジュール (internal/core/)

| モジュール | 責務 |
|-----------|------|
| Zeus | メインロジック、プロジェクト初期化、コマンド実行 |
| StateManager | 状態スナップショット管理、履歴追跡 |
| ApprovalManager | 3段階承認フロー (auto/notify/approve)、ファイルロック |
| TaskHandler | タスクエンティティの CRUD 操作 |
| EntityRegistry | エンティティハンドラーの登録・取得 |

### DI パターン

Zeus は Option パターンによる依存性注入をサポート:

```go
// 本番環境
z := core.New(projectPath)

// テスト環境（モック注入）
z := core.New(projectPath,
    core.WithFileStore(mockFS),
    core.WithStateStore(mockSS),
    core.WithApprovalStore(mockAS),
)
```

### Context 対応

全ての公開 API が `context.Context` を第一引数として受け取る:
- タイムアウト制御
- キャンセル伝播
- 非同期処理のコントロール

### セキュリティ対策

| 対策 | 実装箇所 | 説明 |
|------|----------|------|
| ディレクトリトラバーサル防止 | file_manager.go | ValidatePath でパス検証 |
| ID 衝突防止 | zeus.go, approval.go | UUID v4 ベースの ID 生成 |
| 承認フロー原子性 | approval.go | flock ベースのファイルロック |

### データ構造 (.zeus/)

zeus init 実行後、ターゲットプロジェクトに生成される構造:

```
.zeus/
├── zeus.yaml              # プロジェクト定義（メイン）
├── tasks/
│   ├── active.yaml        # 進行中タスク
│   └── backlog.yaml       # バックログ
├── state/
│   ├── current.yaml       # 現在の状態
│   └── snapshots/         # 履歴スナップショット
├── approvals/             # 承認管理 (standard/advanced)
│   ├── pending/           # 承認待ち
│   ├── approved/          # 承認済み
│   └── rejected/          # 却下済み
└── backups/               # 自動バックアップ
```

### Claude Code 連携（Phase 3 で実装予定）

現時点では Zeus CLI と Claude Code の連携方法が未定義のため、.claude/ ディレクトリ生成は無効化されています。
Phase 3 で適切な連携設計を行った上で有効化予定です。

**Phase 3 で生成予定の構造:**
```
.claude/
├── agents/                # Zeus 用エージェント
│   ├── zeus-orchestrator.md
│   ├── zeus-planner.md
│   └── zeus-reviewer.md
└── skills/                # Zeus 用スキル
    ├── zeus-project-scan/SKILL.md
    ├── zeus-task-suggest/SKILL.md
    └── zeus-risk-analysis/SKILL.md
```

**設計方針:**
- Zeus CLI はスタンドアロンで動作（外部依存なし）
- Claude Code との連携は Plugin として別途実装
- AI 提案は現在ルールベース、Phase 3 で AI ベースに拡張

## 実装フェーズ

| Phase | 内容 | 状態 |
|-------|------|------|
| **Phase 1 (MVP)** | init, status, add, list, doctor, fix | 完了 |
| **Phase 2 (Standard)** | pending, approve, reject, snapshot, history | 完了 |
| **Phase 2.5 (Security)** | パス検証、UUID ID、ファイルロック | 完了 |
| **Phase 2.6 (DI/Context)** | DI対応、Context対応、テスト強化 | 完了 |
| **Phase 2.7 (Suggest)** | suggest, apply (ルールベース提案) | 完了 |
| **Phase 3 (AI統合)** | Claude Code 連携、AI 提案、explain | 未実装 |

## ドキュメント

- `docs/SYSTEM_DESIGN.md` - システム設計書（必読）
- `docs/IMPLEMENTATION_GUIDE.md` - Go 実装ガイド
- `docs/OPERATIONS_MANUAL.md` - 運用マニュアル

## 現在の状態

**Phase 2.7 (Suggest) 完了** - 2026-01-15

### 実装済みコマンド

```bash
# Phase 1 (MVP)
zeus init [--level=simple|standard|advanced]   # プロジェクト初期化
zeus status [--detail]                          # 状態表示
zeus add <entity> <name>                        # エンティティ追加
zeus list [entity]                              # 一覧表示
zeus doctor                                     # 診断
zeus fix [--dry-run]                            # 修復

# Phase 2 (Standard)
zeus pending                                    # 承認待ち一覧
zeus approve <id>                               # 承認
zeus reject <id> [--reason ""]                  # 却下
zeus snapshot create [label]                    # スナップショット作成
zeus snapshot list [-n limit]                   # スナップショット一覧
zeus snapshot restore <timestamp>               # スナップショットから復元
zeus history [-n limit]                         # プロジェクト履歴表示

# Phase 2.7 (Suggest)
zeus suggest [--limit N] [--impact high|medium|low]  # 提案生成（ルールベース）
zeus apply <suggestion-id>                      # 提案適用
zeus apply --all [--dry-run]                    # 全提案適用
```

### 承認レベル

3段階の承認レベルをサポート:

| レベル | 説明 | 動作 |
|--------|------|------|
| auto | 自動承認 | 低リスク操作、即時実行 |
| notify | 通知のみ | 中リスク操作、ログ記録して実行 |
| approve | 明示的承認必要 | 高リスク操作、承認待ちキューに追加 |

**現在の実装状態:**
- Simple レベル: 全操作が auto（承認フローなし）
- Standard/Advanced: 承認基盤は実装済み、Add との連携は Phase 3 で実装予定
- 手動で `zeus approve/reject` コマンドは使用可能

### テスト

```bash
# 全テスト実行
go test ./...

# 詳細出力
go test -v ./internal/core/...

# カバレッジ計測
go test -cover ./...
```

テストカテゴリ:
- DI テスト: モック注入の検証
- Context タイムアウトテスト: キャンセル処理の検証
- 統合テスト: Init から List までのフロー
- TaskHandler 単体テスト: CRUD 操作の検証

### 次のステップ

- Phase 3: AI 統合
  - Claude Code との連携設計・実装
  - AI ベースの提案機能（現在はルールベース）
  - explain コマンドの実装
  - Add コマンドと承認フローの連携
- E2E テストの整備
- テストカバレッジ 80% 達成
