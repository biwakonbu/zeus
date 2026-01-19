# 要件定義 - WBS Dashboard Design Improvement

## 概要

Zeus WBS ダッシュボードの可視化ビューを改善し、Factorio 風テーマとの統一感を持たせつつ、情報の読み取りやすさを向上させる。

## 機能要件

### FR-01: Health View

**目的**: プロジェクト全体の健全性を一目で把握

**表示要素**:
1. **メトリクスパネル** - 3 つの主要指標
   - Coverage（網羅度）: Vision → Objective → Deliverable の連携率
   - Balance（バランス）: 進捗の偏り（標準偏差ベース）
   - Overall Health: 総合評価（%）

2. **階層リスト** - Objective 単位の進捗
   - 折りたたみ可能
   - 各 Objective に水平プログレスバー
   - 配下 Deliverable も表示可能

**インタラクション**:
- 項目クリックで詳細パネル表示
- 折りたたみトグル
- ホバーでハイライト

### FR-02: Timeline View

**目的**: 計画 vs 実績の時間的乖離を可視化

**表示要素**:
1. **時間軸スケール** - 週/月/四半期切り替え
2. **計画/実績バー** - 2 段表示
   - 計画（薄色）: start_date → due_date
   - 実績（濃色）: actual_start → 現在または actual_end
3. **進捗インジケータ** - ON TRACK / DELAYED / AHEAD / COMPLETED

**インタラクション**:
- スケール切り替え（W/M/Q）
- Today ボタンで現在日にスクロール
- バークリックで詳細パネル表示

### FR-03: Density View

**目的**: 作業量の分布を可視化

**表示要素**:
1. **ヒートマップグリッド** - Objective 単位
   - セルサイズ: 固定（CSS Grid）
   - 色: 進捗率に基づく（赤→黄→緑）
   - 数値: タスク数または工数

**インタラクション**:
- セルクリックで詳細パネル表示
- サイズ指標切り替え（タスク数/工数）

### FR-04: ビュー切り替え

**方式**: タブ UI

**タブ構成**:
- Health（デフォルト）
- Timeline
- Density

**状態共有**:
- selectedEntityId: 選択中エンティティ
- expandedIds: 展開済み項目リスト

## 非機能要件

### NFR-01: パフォーマンス

- 100 ノード以下でスムーズ動作
- 初期表示: 1 秒以内
- インタラクション: 100ms 以内

### NFR-02: デザイン統一

- Factorio テーマ CSS 変数を 100% 使用
- 既存コンポーネント（WBSSummaryBar, EntityDetailPanel）との視覚的一貫性
- D3.js 依存を排除（Svelte + CSS のみ）

### NFR-03: 後方互換性

- 既存 API（/api/wbs/aggregated）を維持
- 新フィールドはオプショナル
- 既存ビュー（Progress, Issues, Coverage, Resources）は削除可

### NFR-04: アクセシビリティ

- キーボードナビゲーション対応
- 十分な色コントラスト比
- aria 属性による支援技術対応

## 受け入れ基準

### AC-01: Health View

- [x] 3 メトリクスが正しく計算・表示される
- [x] 階層リストが折りたたみ可能
- [x] クリックで詳細パネルが開く

### AC-02: Timeline View

- [x] 計画/実績の 2 段バーが表示される
- [x] W/M/Q スケール切り替えが動作
- [x] 乖離状態が正しく表示される

### AC-03: Density View

- [x] ヒートマップグリッドが表示される
- [x] 色が進捗率を反映
- [x] セルクリックで詳細パネルが開く

### AC-04: 統合

- [x] タブ切り替えが正常動作
- [x] 選択状態がビュー間で保持される
- [x] Factorio テーマと統一感がある

## 除外事項

- Momentum メトリクス（履歴機能実装後）
- Packed Bubble（Density View 将来拡張）
- リアルタイム更新（SSE 対応は別タスク）
- 実データ連携（Timeline の actual_start/due_date）
