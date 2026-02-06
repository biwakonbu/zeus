# CLAUDE.md

Zeus は「神の視点」でプロジェクト構造を扱う AI 駆動 CLI/ダッシュボード基盤。実装は Go + Cobra + YAML ストレージ。

## 正本ルール

- 正本入口: `docs/README.md`
- CLI 契約正本: `cmd/*.go` の `cobra.Command`
- HTTP 契約正本: `internal/dashboard/server.go` の `mux.HandleFunc`
- 履歴資料は現行仕様の判断根拠に使わない

## 技術スタック

- Backend: Go 1.21+, Cobra, YAML
- Frontend: SvelteKit + TypeScript + PixiJS
- 配信: REST API + SSE

## 開発コマンド

```bash
make build
make test
go test ./...

# ダッシュボード開発
make dashboard-deps
make dashboard-dev
go run . dashboard --dev
```

## 実装フェーズ

### 機能フェーズ

| Phase | 内容 | 状態 |
|---|---|---|
| Phase 1 | init, status, add, list, doctor, fix | 完了 |
| Phase 2 | pending, approve, reject, snapshot, history | 完了 |
| Phase 3 | suggest, apply, explain, update-claude | 完了 |
| Phase 4 | graph, report | 完了 |
| Phase 5 | dashboard, REST API, SSE | 完了 |
| Phase 7 | Affinity Canvas API | 完了 |
| 概念モデル Phase 1-3 | Vision〜Quality + UML + Activity | 完了 |

### 外部連携フェーズ

| 項目 | 状態 | 備考 |
|---|---|---|
| Git 自動連携 | 未実装 | 手動運用前提 |
| Slack/Email 通知 | 未実装 | SSE/CLI で代替 |
| 認証・認可 | 未実装 | 127.0.0.1 バインド運用 |

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

## 参照先

- 正本入口: `docs/README.md`
- 運用手順: `docs/operations-manual.md`
- 設計: `docs/system-design.md`
- API 契約: `docs/api-reference.md`
- ユーザー向け: `docs/user-guide.md`
- 履歴資料: `docs/implementation-guide.md`, `docs/specs/remove-progress-features/README.md`

*更新日: 2026-02-06（実装同期版）*
