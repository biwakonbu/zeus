# WBS Dashboard UX Improvement

WBS Dashboard のユーザーエクスペリエンス改善仕様。
Factorio 風インダストリアル UI を維持しつつ、Rich Tooltip と Drill-Down Mode による二層構造でエンティティ情報にアクセスする。

## 概要

| 項目 | 値 |
|------|-----|
| バージョン | 1.0.0 |
| ステータス | 実装完了 |
| 対象モジュール | zeus-dashboard |

## 設計方針

1. **Factorio 風の洗練**
   - 角丸: 4px（シャープすぎず、丸すぎず）
   - シャドウ: レイヤード表現でデプス感を強調
   - CSS 変数: デザイントークンとして一元管理

2. **固定サイドバー排除**
   - 画面を狭くする固定サイドバーは採用しない
   - 代替: Rich Tooltip + Drill-Down Mode の二層構造

3. **統一インタラクション**
   - 全ビュー（Health, Timeline, Density, Affinity, Graph）で同一操作

## インタラクション仕様

| 操作 | 結果 |
|------|------|
| ホバー（500ms） | Rich Tooltip 表示（320x220px） |
| クリック | 選択状態（ハイライト） |
| ダブルクリック | Drill-Down Mode へ遷移 |
| Cmd+クリック | 複数選択 |
| Escape | 選択解除 / Drill-Down 終了 |

## 関連ドキュメント

- [デザイントークン](./design-tokens.md)
- [Rich Tooltip 仕様](./rich-tooltip.md)
- [Drill-Down Mode 仕様](./drill-down.md)
