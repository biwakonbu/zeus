---
description: ドキュメント編集時のガイドライン。設計書・仕様書・ガイド編集時に参照。
paths:
  - "docs/**"
---

# ドキュメント

## ディレクトリ構成

| パス | 内容 |
|------|------|
| `docs/system-design.md` | システム設計書（アーキテクチャ概要） |
| `docs/detailed-design.md` | 詳細設計書（10概念モデル） |
| `docs/implementation-guide.md` | 実装ガイド（Go コーディング指針） |
| `docs/operations-manual.md` | 運用マニュアル |
| `docs/security.md` | セキュリティ実装ガイド |
| `docs/api-spec.md` | API 仕様書 |
| `docs/api-reference.md` | API リファレンス |
| `docs/user-guide.md` | ユーザーガイド |
| `docs/specs/` | 機能別仕様書（requirements.md, design.md） |

## 記載基準

- **技術的正確性を優先**: コード例は実際に動作するものを使用
- **バージョン整合性**: 実装済み機能と仕様書の記述を一致させる
- **相互参照**: 関連ドキュメント間のリンクを維持

## 編集時の注意

1. **コード例の更新**: API 変更時は対応するドキュメントのコード例も更新
2. **フェーズ記載**: 実装フェーズの状態は CLAUDE.md で管理（docs は詳細のみ）
3. **specs ディレクトリ**: 新機能は `docs/specs/<feature>/` に requirements.md と design.md を配置

## 命名規則

- **小文字ケバブケース**: トップレベルドキュメント（system-design.md）
- **小文字ハイフン**: specs 内のディレクトリ名（phase5-dashboard）
- **小文字**: specs 内のファイル名（requirements.md, design.md）
- **例外**: README.md は GitHub 表示慣例のため大文字維持
