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
│   ├── root.go               # ルートコマンド
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
│   │   ├── zeus.go           # メインロジック
│   │   ├── types.go          # 型定義
│   │   ├── state.go          # 状態・スナップショット管理
│   │   ├── approval.go       # 3段階承認システム
│   │   └── errors.go         # エラー定義
│   ├── yaml/                 # YAML 操作
│   │   ├── parser.go
│   │   ├── writer.go
│   │   └── file_manager.go
│   ├── doctor/               # 診断・修復
│   │   └── doctor.go
│   └── generator/            # Claude Code 連携ファイル生成
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
| ApprovalManager | 3段階承認フロー (auto/notify/approve) |

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

### Claude Code 連携

zeus init --level=standard/advanced は .claude/ ディレクトリも生成:

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

## 実装フェーズ

| Phase | 内容 | 状態 |
|-------|------|------|
| **Phase 1 (MVP)** | init, status, add, list, doctor, fix | 完了 |
| **Phase 2 (Standard)** | pending, approve, reject, snapshot, history, Claude Code 連携 | 完了 |
| **Phase 3 (AI統合)** | suggest, apply, explain | 未実装 |

## ドキュメント

- `docs/SYSTEM_DESIGN.md` - システム設計書（必読）
- `docs/IMPLEMENTATION_GUIDE.md` - Go 実装ガイド
- `docs/OPERATIONS_MANUAL.md` - 運用マニュアル

## 現在の状態

**Phase 2 (Standard) 完了** - 2026-01-14

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
```

### 承認レベル

3段階の承認レベルをサポート:

| レベル | 説明 | 動作 |
|--------|------|------|
| auto | 自動承認 | 低リスク操作、即時実行 |
| notify | 通知のみ | 中リスク操作、ログ記録して実行 |
| approve | 明示的承認必要 | 高リスク操作、承認待ちキューに追加 |

### 次のステップ

- Phase 3: AI 提案機能 (suggest, apply, explain)
- テストカバレッジの向上
