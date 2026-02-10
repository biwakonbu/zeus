# Zeus API リファレンス

> 文書メタデータ
> - 文書種別: 正本
> - 実装状態: 完了
> - 正本ソース: `cmd/*.go`, `internal/dashboard/server.go`
> - 最終検証日: `2026-02-07`
> 正本判定: `docs/README.md` を参照。CLI は `cmd/*.go`、HTTP API は `internal/dashboard/server.go` を正本とする。

## 1. 概要

Zeus は以下の 2 系統の公開インターフェースを提供する。

1. CLI (`zeus ...`)
2. Dashboard HTTP API (`/api/...`)

## 2. CLI リファレンス

## 2.1 共通構文

```bash
zeus <command> [subcommand] [arguments] [flags]
```

## 2.2 グローバルフラグ

| フラグ | 短縮 | デフォルト | 説明 |
|---|---|---|---|
| `--verbose` | `-v` | `false` | 詳細出力 |
| `--format` | `-f` | `text` | 出力形式（text/json） |

## 2.3 コマンド一覧

| カテゴリ | コマンド | 概要 |
|---|---|---|
| コア | `init` | プロジェクト初期化 |
| コア | `status` | 現在状態表示 |
| コア | `add` | エンティティ追加 |
| コア | `list` | エンティティ一覧 |
| コア | `doctor` | 整合性診断 |
| コア | `fix` | 自動修復 |
| 承認 | `pending` | 承認待ち一覧 |
| 承認 | `approve <id>` | 承認 |
| 承認 | `reject <id>` | 却下 |
| 履歴 | `snapshot create [label]` | スナップショット作成 |
| 履歴 | `snapshot list [-n N]` | スナップショット一覧 |
| 履歴 | `snapshot restore <timestamp>` | スナップショット復元 |
| 履歴 | `history [-n N]` | 履歴表示 |
| AI支援 | `suggest` | 提案生成 |
| AI支援 | `apply` | 提案適用 |
| AI支援 | `explain` | エンティティ解説 |
| AI支援 | `update-claude` | Claude 連携ファイル更新 |
| 可視化 | `graph` | 依存グラフ |
| 可視化 | `report` | レポート生成 |
| 可視化 | `dashboard` | Web ダッシュボード起動 |
| UML | `uml show usecase` | UseCase 図出力 |
| UML | `usecase add-actor` | UseCase と Actor の関連付け |
| UML | `usecase link` | UseCase 関係追加 |

## 2.4 `add` 対応エンティティ

`zeus add <entity> <name>` の `<entity>`:

- `vision`
- `objective`
- `consideration`
- `decision`
- `problem`
- `risk`
- `assumption`
- `constraint`
- `quality`
- `actor`
- `usecase`
- `subsystem`
- `activity`

## 2.5 重要コマンド仕様

### graph

```bash
zeus graph [--format text|dot|mermaid] [-o FILE]
zeus graph --unified [--focus ID] [--depth N]
zeus graph --unified --types activity,usecase,objective
zeus graph --unified --layers structural,reference
zeus graph --unified --relations parent,depends_on,implements,contributes
zeus graph --unified --hide-completed --hide-draft
```

### dashboard

```bash
zeus dashboard [--port 8080] [--no-open] [--dev]
```

### report

```bash
zeus report [--format text|html|markdown] [-o FILE]
```

### suggest / apply

```bash
zeus suggest [--limit N] [--impact high|medium|low] [--force]
zeus apply <suggestion-id> [--dry-run]
zeus apply --all [--dry-run]
```

### uml show usecase

```bash
zeus uml show usecase [--boundary NAME] [--format text|mermaid] [-o FILE]
```

### usecase add-actor

```bash
zeus usecase add-actor <usecase-id> <actor-id> [--role primary|secondary]
```

### usecase link

```bash
zeus usecase link <usecase-id> --include <target-id>
zeus usecase link <usecase-id> --extend <target-id> [--condition TEXT] [--extension-point TEXT]
zeus usecase link <usecase-id> --generalize <target-id>
```

## 3. HTTP API リファレンス

Base URL:

```text
http://127.0.0.1:8080
```

## 3.1 Core API

### GET /api/status

プロジェクト状態を返す。

```bash
curl -s http://127.0.0.1:8080/api/status | jq
```

主なレスポンス項目:
- `project`
- `state.health`
- `state.summary.total_activities`
- `pending_approvals`

### GET /api/graph

依存グラフ（Mermaid + 統計）を返す。

```bash
curl -s http://127.0.0.1:8080/api/graph | jq '.stats'
```

主なレスポンス項目:
- `mermaid`
- `stats`
- `cycles`
- `isolated`

## 3.2 Affinity API

### GET /api/affinity

Affinity 計算結果を返す。

クエリ:
- `max_siblings` (int)
- `min_score` (float)
- `max_edges` (int)

```bash
curl -s "http://127.0.0.1:8080/api/affinity?max_siblings=20&min_score=0.2&max_edges=300" | jq '.stats'
```

主なレスポンス項目:
- `nodes`
- `edges`
- `clusters`
- `weights`
- `stats`

## 3.3 UML/Activity API

### GET /api/actors

```bash
curl -s http://127.0.0.1:8080/api/actors | jq
```

レスポンス:
- `actors`
- `total`

### GET /api/usecases

```bash
curl -s http://127.0.0.1:8080/api/usecases | jq '.total'
```

レスポンス:
- `usecases`
- `total`

### GET /api/subsystems

```bash
curl -s http://127.0.0.1:8080/api/subsystems | jq '.total'
```

レスポンス:
- `subsystems`
- `total`

### GET /api/uml/usecase

クエリ:
- `boundary` (string, optional)

```bash
curl -s "http://127.0.0.1:8080/api/uml/usecase?boundary=System" | jq '.mermaid'
```

レスポンス:
- `actors`
- `usecases`
- `boundary`
- `mermaid`

### GET /api/activities

```bash
curl -s http://127.0.0.1:8080/api/activities | jq '.total'
```

レスポンス:
- `activities`
- `total`

### GET /api/uml/activity

クエリ:
- `id` (必須)

```bash
curl -s "http://127.0.0.1:8080/api/uml/activity?id=act-001" | jq
```

レスポンス:
- `activity`
- `mermaid`

## 3.4 Unified Graph API

### GET /api/unified-graph

Activity / UseCase / Objective を統合したグラフを返す。

クエリ:
- `focus`
- `depth`
- `types`
- `layers`
- `relations`
- `hide-completed`
- `hide-draft`

```bash
curl -s http://127.0.0.1:8080/api/unified-graph | jq '.stats'
curl -s "http://127.0.0.1:8080/api/unified-graph?layers=structural" | jq '.stats'
curl -s "http://127.0.0.1:8080/api/unified-graph?relations=depends_on,contributes" | jq '.filter'
curl -s "http://127.0.0.1:8080/api/unified-graph?focus=act-001&depth=2" | jq '.filter'
```

主なレスポンス項目:
- `nodes`
- `edges`
- `stats`
- `cycles`
- `isolated`
- `mermaid`
- `filter`

## 3.5 SSE API

### GET /api/events

SSE ストリームを開く。

```bash
curl -N http://127.0.0.1:8080/api/events
```

イベントタイプ:
- `connected`
- `status`
- `graph`
- `approval`

## 4. エラーレスポンス

- 不正メソッド: `405 Method Not Allowed`
- 必須パラメータ不足: `400 Bad Request`
- 対象不在: `404 Not Found`
- 内部エラー: `500 Internal Server Error`

## 5. ドキュメント運用ルール

- CLI/API の公開契約変更時は本書を同時更新する。
- 契約差異が疑われる場合は `cmd/*.go` と `internal/dashboard/server.go` を優先確認する。
- 正本/履歴の分類は `docs/README.md` を参照する。

*更新日: 2026-02-10（Deliverable削除・SimpleMode廃止対応）*
