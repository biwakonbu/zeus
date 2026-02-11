# CLAUDE.md

Zeus は「神の視点」でプロジェクト構造を扱う AI 駆動 CLI/ダッシュボード基盤。実装は Go + Cobra + YAML ストレージ。

## 正本ルール

- 正本入口: `docs/README.md`
- CLI 契約正本: `cmd/*.go` の `cobra.Command`
- HTTP 契約正本: `internal/dashboard/server.go` の `mux.HandleFunc`
- 履歴資料（`docs/archive/`）は現行仕様の判断根拠に使わない

## 技術スタック

- Backend: Go `1.24.x`（`go.mod` 準拠）, Cobra, YAML
- Frontend: SvelteKit + TypeScript + PixiJS
- 配信: REST API + SSE

## ドキュメント導線

- 正本: `docs/operations-manual.md`, `docs/system-design.md`, `docs/api-reference.md`, `docs/user-guide.md`
- 仕様（実装完了）: `docs/specs/unified-graph-two-layer/README.md`
- 設計（実装中）: なし
- 履歴（凍結）: `docs/archive/*.md`

## 開発コマンド

```bash
make build
make test
go test ./...

# ダッシュボード開発
make dashboard-deps
make dashboard-dev
go run . dashboard --dev

# Claude 連携再生成
go run . update-claude
```

## 実装済み CLI（公開）

```bash
# Core
zeus init
zeus status
zeus add <entity> <name>
zeus list [entity]
zeus doctor
zeus fix [--dry-run]

# Approval / History
zeus pending
zeus approve <id>
zeus reject <id> [--reason TEXT]
zeus snapshot create|list|restore
zeus history [-n N]

# AI
zeus suggest [--limit N] [--impact high|medium|low]
zeus apply [suggestion-id] [--all] [--dry-run]
zeus explain <entity-id> [--context]
zeus update-claude

# Analysis / Visualization
zeus graph [--format text|dot|mermaid] [-o FILE]
zeus graph --unified [--focus ID] [--depth N] [--types ...] [--layers ...] [--relations ...]
zeus report [--format text|html|markdown] [-o FILE]
zeus dashboard [--port N] [--no-open] [--dev]

# UML
zeus uml show usecase [--boundary NAME] [--format text|mermaid] [-o FILE]
zeus usecase add-actor <usecase-id> <actor-id> [--role primary|secondary]
zeus usecase link <usecase-id> --include|--extend|--generalize ...
```

## 実装済み HTTP API（公開）

- `GET /api/status`
- `GET /api/graph`
- `GET /api/affinity`
- `GET /api/actors`
- `GET /api/usecases`
- `GET /api/subsystems`
- `GET /api/uml/usecase`
- `GET /api/activities`
- `GET /api/uml/activity`
- `GET /api/unified-graph`
- `GET /api/events` (SSE)

## 外部連携（未実装）

- Git 自動連携
- Slack/Email 通知
- 認証・認可（ローカルバインド前提運用）

*更新日: 2026-02-12（エージェント/スキル同期）*
