# Zeus ユーザーガイド

> 文書メタデータ
> - 文書種別: 正本
> - 実装状態: 完了
> - 正本ソース: `cmd/*.go`, `internal/dashboard/server.go`
> - 最終検証日: `2026-02-07`
> 現行仕様の正本入口: `docs/README.md`

## 1. Zeus とは

Zeus は、YAML ベースでプロジェクト構造を管理する CLI + Web ダッシュボードです。Activity、UseCase、Objective を統合的に可視化し、日次運用と設計レビューを支援します。

## 2. はじめ方

## 2.1 インストール（ソースから）

```bash
git clone https://github.com/biwakonbu/zeus.git
cd zeus
make build
```

## 2.2 初期化

```bash
cd your-project
zeus init
```

## 2.3 最初の確認

```bash
zeus status
zeus list activities
```

## 3. 基本操作

## 3.1 エンティティ追加

```bash
zeus add vision "AIで設計品質を上げる"
zeus add objective "設計レビュー自動化"
zeus add activity "API一覧を整備"
```

## 3.2 一覧確認

```bash
zeus list
zeus list objectives
zeus list activities
```

## 3.3 品質確認

```bash
zeus doctor
zeus fix --dry-run
```

## 4. 承認フロー

## 4.1 承認待ち確認

```bash
zeus pending
```

## 4.2 承認/却下

```bash
zeus approve <id>
zeus reject <id> --reason "理由"
```

## 5. AI 支援

## 5.1 提案生成

```bash
zeus suggest --limit 5 --impact high
```

## 5.2 提案適用

```bash
zeus apply <suggestion-id>
# または
zeus apply --all --dry-run
```

## 5.3 詳細解説

```bash
zeus explain act-001 --context
```

## 6. 可視化とレポート

## 6.1 依存グラフ

```bash
zeus graph
zeus graph --format mermaid -o docs/deps.md
```

## 6.2 統合グラフ

```bash
zeus graph --unified
zeus graph --unified --layers structural
zeus graph --unified --relations implements
zeus graph --unified --focus act-001 --depth 2
```

## 6.3 レポート出力

```bash
zeus report --format markdown -o report.md
zeus report --format html -o report.html
```

## 6.4 ダッシュボード

```bash
zeus dashboard --port 8080
```

開発モード:

```bash
zeus dashboard --dev --port 8080
```

## 7. UML 操作

## 7.1 UseCase 図を出力

```bash
zeus uml show usecase --format mermaid -o docs/usecase.md
```

## 7.2 UseCase と Actor の関連付け

```bash
zeus usecase add-actor uc-001 actor-001 --role primary
```

## 7.3 UseCase 関係追加

```bash
zeus usecase link uc-001 --include uc-002
zeus usecase link uc-001 --extend uc-003 --condition "任意機能選択時" --extension-point "決済方式"
zeus usecase link uc-001 --generalize uc-004
```

## 8. API を使った確認

```bash
curl -s http://127.0.0.1:8080/api/status | jq '.state.health'
curl -s http://127.0.0.1:8080/api/unified-graph | jq '.stats'
curl -s http://127.0.0.1:8080/api/affinity | jq '.stats'
curl -s http://127.0.0.1:8080/api/actors | jq '.total'
curl -s http://127.0.0.1:8080/api/usecases | jq '.total'
curl -s http://127.0.0.1:8080/api/subsystems | jq '.total'
curl -s "http://127.0.0.1:8080/api/uml/usecase?boundary=System" | jq '.boundary'
curl -s http://127.0.0.1:8080/api/activities | jq '.total'
curl -s "http://127.0.0.1:8080/api/uml/activity?id=act-001" | jq '.activity.id'
```

SSE:

```bash
curl -N http://127.0.0.1:8080/api/events
```

## 9. よくある問題

## 9.1 ダッシュボードが起動しない

```bash
zeus dashboard --port 18080
```

## 9.2 整合性エラーが出る

```bash
zeus doctor
zeus fix --dry-run
```

## 9.3 グラフが空になる

```bash
zeus list activities
zeus uml show usecase --format text
zeus list objectives
```

## 10. 関連文書

- 正本入口: `docs/README.md`
- 運用手順: `docs/operations-manual.md`
- 設計: `docs/system-design.md`
- CLI/API 契約: `docs/api-reference.md`
- 開発要約: `CLAUDE.md`

*更新日: 2026-02-10（Deliverable削除・SimpleMode廃止対応）*
