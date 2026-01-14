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
│   └── ...
├── internal/                 # 内部パッケージ
│   ├── core/                 # コアロジック (zeus.go, state.go, approval.go)
│   ├── yaml/                 # YAML 操作 (parser.go, writer.go)
│   ├── doctor/               # 診断・修復
│   └── generator/            # Claude Code 連携ファイル生成
├── templates/                # embed 用テンプレート
│   ├── agents/               # agent テンプレート
│   └── skills/               # skill テンプレート
├── main.go
├── go.mod
└── Makefile
```

## アーキテクチャ

### コアモジュール (internal/core/)

| モジュール | 責務 |
|-----------|------|
| Zeus | メインロジック、プロジェクト初期化 |
| StateManager | 状態スナップショット管理 |
| ApprovalManager | 3段階承認フロー (auto/notify/approve) |

### データ構造 (.zeus/)

zeus init 実行後、ターゲットプロジェクトに生成される構造:

```
.zeus/
├── zeus.yaml          # プロジェクト定義（メイン）
├── tasks/
│   ├── active.yaml    # 進行中タスク
│   └── backlog.yaml   # バックログ
├── state/
│   ├── current.yaml   # 現在の状態
│   └── snapshots/     # 履歴スナップショット
└── backups/           # 自動バックアップ
```

### Claude Code 連携

zeus init は .claude/ ディレクトリも生成:

```
.claude/
├── agents/            # Zeus 用エージェント
│   ├── zeus-orchestrator.md
│   ├── zeus-planner.md
│   └── zeus-reviewer.md
└── skills/            # Zeus 用スキル
    ├── zeus-project-scan/
    ├── zeus-task-suggest/
    └── zeus-risk-analysis/
```

## 実装フェーズ

| Phase | 内容 |
|-------|------|
| **Phase 1 (MVP)** | init, status, add, list, doctor, fix |
| **Phase 2 (AI統合)** | suggest, apply, explain |
| **Phase 3 (承認)** | pending, approve, reject |

## ドキュメント

- `docs/SYSTEM_DESIGN.md` - システム設計書（必読）
- `docs/IMPLEMENTATION_GUIDE.md` - Go 実装ガイド
- `docs/OPERATIONS_MANUAL.md` - 運用マニュアル

## 現在の状態

**Phase 1 (MVP) 完了** - 2026-01-14

### 実装済みコマンド

```bash
zeus init [--level=simple|standard|advanced]   # プロジェクト初期化
zeus status [--detail]                          # 状態表示
zeus add <entity> <name>                        # エンティティ追加
zeus list [entity]                              # 一覧表示
zeus doctor                                     # 診断
zeus fix [--dry-run]                            # 修復
```

### 次のステップ

- Phase 2: 承認システム (auto/notify/approve)、スナップショット機能
- Phase 3: Claude Code 連携ファイル生成 (.claude/agents/, .claude/skills/)
