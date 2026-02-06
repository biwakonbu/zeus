# Zeus ドキュメント正本マトリクス

本ファイルは、Zeus の運用設計ドキュメントにおける正本判定の唯一の入口です。

## 1. 最終同期情報

- 最終同期日: `2026-02-06`
- 同期方針: 実装を正本として文書のみ同期
- 対象ブランチ: リポジトリ `HEAD`

## 2. 正本判定ルール

1. CLI 契約は `cmd/*.go` の `cobra.Command` 定義を正本とする。
2. HTTP API 契約は `internal/dashboard/server.go` の `mux.HandleFunc` 定義を正本とする。
3. 文書間で記載が衝突した場合は、必ず実装定義を優先する。
4. 履歴資料は参照可能だが、現行運用判断の根拠としては使わない。

## 3. 現行正本（運用判断に使用可）

| 区分 | 文書 | 役割 |
|---|---|---|
| 運用 | `docs/operations-manual.md` | 日次運用手順、障害時対応、CLI/API運用 |
| 設計 | `docs/system-design.md` | 現行アーキテクチャ、データフロー、フェーズ定義 |
| 契約 | `docs/api-reference.md` | 公開 CLI/API 契約 |
| 利用 | `docs/user-guide.md` | 利用者向け導入・利用手順 |
| 開発 | `CLAUDE.md` | 開発者向け実装・運用要約 |

## 4. 履歴資料（非正本）

以下は履歴保全目的の資料です。現行仕様判断には使わず、必ず本ファイルから現行正本へ遷移してください。

- `docs/implementation-guide.md`
- `docs/specs/remove-progress-features/README.md`

## 5. 文書更新ポリシー

- コード変更を伴わない文書同期では、まず本ファイルの最終同期日を更新する。
- 新規公開コマンド/APIの追加時は、`cmd/*.go` / `internal/dashboard/server.go` への実装反映を確認後、正本文書へ同期する。
- 廃止済み仕様は履歴資料へ退避し、現行正本からは除外する。

*更新日: 2026-02-06（実装同期版）*
