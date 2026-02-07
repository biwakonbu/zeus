---
description: ドキュメント編集時のガイドライン。設計書・仕様書・ガイド編集時に参照。
paths:
  - "docs/**"
---

# ドキュメント

## 構造（2026-02-07 再編後）

| 区分 | パス | 役割 |
|------|------|------|
| 入口 | `docs/README.md` | 文書種別/実装状態の判定マトリクス |
| 正本 | `docs/operations-manual.md` | 運用手順 |
| 正本 | `docs/system-design.md` | 現行システム設計 |
| 正本 | `docs/api-reference.md` | 公開 CLI/API 契約 |
| 正本 | `docs/user-guide.md` | 利用手順 |
| 仕様 | `docs/specs/unified-graph-two-layer/README.md` | 実装完了仕様 |
| 設計 | なし | 現在は個別設計文書なし |
| 履歴 | `docs/archive/*.md` | 凍結文書（参照のみ） |

## 記載基準

- 技術的正確性を優先し、実装と一致しないコマンド例を残さない。
- 断片更新ではなく、`docs/README.md` の分類・状態列と同時に同期する。
- 履歴化した文書は `docs/archive/` へ移動し、旧パスには移動案内スタブを残す。

## 編集時の注意

1. CLI/API 契約変更時は `cmd/*.go` / `internal/dashboard/server.go` を確認してから記述する。
2. 正本・仕様・設計・履歴の分類変更時は `docs/README.md` を先に更新する。
3. `docs/archive/` 配下は凍結扱い。内容修正はリンク修復と注記追加に限定する。

## 命名規則

- トップレベル文書: 小文字ケバブケース（例: `system-design.md`）
- `docs/specs/` 配下: ディレクトリ単位で機能名を保持
- 移動案内スタブ: 元ファイル名を維持（互換導線のため）
