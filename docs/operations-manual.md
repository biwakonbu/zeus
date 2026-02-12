# Zeus 運用マニュアル

> 文書メタデータ
> - 文書種別: 正本
> - 実装状態: 完了
> - 正本ソース: `cmd/*.go`, `internal/dashboard/server.go`
> 正本判定: `docs/README.md` を参照。CLI は `cmd/*.go`、HTTP API は `internal/dashboard/server.go` を正本とする。

## 1. 目的

本書は、現行実装に基づく Zeus の運用手順を定義する。日次運用、障害一次対応、CLI/API の確認手順を対象とする。

## 2. 運用開始

### 2.1 初期化

```bash
zeus init
```

### 2.2 状態確認

```bash
zeus status
zeus doctor
```

### 2.3 ダッシュボード起動

```bash
zeus dashboard --port 8080
```

開発モード（Vite + Go API 分離）:

```bash
zeus dashboard --dev --port 8080
```

## 3. CLI 運用リファレンス

### 3.1 コア操作

| コマンド | 用途 |
|---|---|
| `zeus init` | プロジェクト初期化 |
| `zeus status` | 状態確認 |
| `zeus add <entity> <name>` | エンティティ追加 |
| `zeus list [entity]` | 一覧確認 |
| `zeus doctor` | 整合性診断 |
| `zeus fix [--dry-run]` | 自動修復 |

### 3.2 AI 支援

| コマンド | 用途 |
|---|---|
| `zeus suggest [--limit N] [--impact high|medium|low]` | 提案生成 |
| `zeus apply [suggestion-id] [--all] [--dry-run]` | 提案適用 |
| `zeus explain <entity-id> [--context]` | エンティティ解説 |

### 3.3 承認・履歴

| コマンド | 用途 |
|---|---|
| `zeus pending` | 承認待ち一覧 |
| `zeus approve <id>` | 承認 |
| `zeus reject <id> --reason "..."` | 却下 |
| `zeus snapshot create [label]` | スナップショット作成 |
| `zeus snapshot list [-n N]` | スナップショット一覧 |
| `zeus snapshot restore <timestamp>` | スナップショット復元 |
| `zeus history [-n N]` | 履歴表示 |

### 3.4 可視化・レポート

| コマンド | 用途 |
|---|---|
| `zeus graph [--format text|dot|mermaid] [-o file]` | 依存グラフ |
| `zeus graph --unified [--focus ID] [--depth N]` | 統合グラフ |
| `zeus graph --unified --layers structural,reference` | 2層フィルタ |
| `zeus graph --unified --relations ...` | 関係種別フィルタ |
| `zeus report [--format text|html|markdown] [-o file]` | レポート出力 |
| `zeus dashboard [--port N] [--no-open] [--dev]` | Web ダッシュボード |

### 3.5 UML 操作

| コマンド | 用途 |
|---|---|
| `zeus uml show usecase [--boundary NAME] [--format text|mermaid] [-o file]` | ユースケース図出力 |
| `zeus usecase add-actor <usecase-id> <actor-id> [--role primary|secondary]` | UseCase と Actor の関連付け |
| `zeus usecase link <usecase-id> --include|--extend|--generalize ...` | UseCase 関係追加 |

## 4. API 運用チェック

ダッシュボード起動後、以下で API 契約の実在確認を行う。

```bash
curl -s http://127.0.0.1:8080/api/status | jq '.state.health'
curl -s http://127.0.0.1:8080/api/graph | jq '.stats'
curl -s "http://127.0.0.1:8080/api/unified-graph?layers=structural" | jq '.stats'
curl -s "http://127.0.0.1:8080/api/affinity?max_siblings=20&min_score=0.2" | jq '.stats'
curl -s http://127.0.0.1:8080/api/actors | jq '.total'
curl -s http://127.0.0.1:8080/api/usecases | jq '.total'
curl -s http://127.0.0.1:8080/api/subsystems | jq '.total'
curl -s "http://127.0.0.1:8080/api/uml/usecase?boundary=System" | jq '.boundary'
curl -s http://127.0.0.1:8080/api/activities | jq '.total'
curl -s "http://127.0.0.1:8080/api/uml/activity?id=act-001" | jq '.activity.id'
```

SSE 接続確認:

```bash
curl -N http://127.0.0.1:8080/api/events
```

## 5. 日次運用手順

1. `zeus status` で全体状態を確認する。
2. `zeus list vision` で Vision の整合性を確認する。
3. `zeus list objectives` で Objective の進捗を確認する。
4. `zeus pending` を確認し、承認・却下を処理する。
5. `zeus doctor` を実行し、必要なら `zeus fix --dry-run` で修復内容を確認する。
6. `zeus graph --unified --layers structural,reference` で関係変化を確認する。
7. 必要に応じて `zeus report --format markdown -o report.md` を出力する。

## 6. 週次運用手順

1. `zeus snapshot create "weekly-review"` を実行する。
2. `zeus history -n 20` で推移を確認する。
3. `zeus dashboard` で API と可視化を目視確認する。
4. 変更内容を Git にコミットし、差分レビューを行う。

## 7. 障害一次対応

### 7.1 ポート競合

症状:
- `zeus dashboard` 起動時にバインド失敗。

対応:

```bash
zeus dashboard --port 18080
```

### 7.2 UML Activity API が 400 を返す

症状:
- `/api/uml/activity` が `id パラメータが必要です` を返す。

対応:

```bash
curl -s "http://127.0.0.1:8080/api/uml/activity?id=act-001" | jq
```

### 7.3 グラフが空になる

症状:
- `/api/unified-graph` が空配列を返す。

確認:

```bash
zeus list activities
zeus uml show usecase --format text
zeus list objectives
```

### 7.4 承認待ちが滞留する

症状:
- `zeus pending` に項目が残り続ける。

対応:

```bash
zeus approve <id>
# または
zeus reject <id> --reason "判断理由"
```

## 8. 運用上の禁止事項

- 実装に存在しない CLI/API を運用手順へ記載しない。
- 履歴資料を現行仕様の判断根拠にしない。

## 9. 関連文書

- 正本入口: `docs/README.md`
- 設計: `docs/system-design.md`
- 契約: `docs/api-reference.md`
- 利用者向け: `docs/user-guide.md`
- 開発要約: `CLAUDE.md`
