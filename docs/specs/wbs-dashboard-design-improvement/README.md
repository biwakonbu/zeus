# WBS Dashboard Design Improvement

## 概要

Zeus WBS ダッシュボードの可視化ビューを改善し、プロジェクトの健全性・進捗・作業量を直感的に把握できる 3 つの視点を提供する。

## 背景

従来の WBS ダッシュボードでは D3.js による Treemap 表示を使用していたが、以下の課題があった:

1. 情報の階層構造が把握しにくい
2. 時間軸での進捗比較ができない
3. Factorio テーマとの統一感が不足

## 3 ビュー設計

### Health View (P1)

プロジェクト全体の健全性を一目で把握するビュー。

**主要コンポーネント**:
- **MetricsPanel**: Coverage, Balance, Overall Health の 3 指標
- **ObjectiveList**: 折りたたみ可能な階層リスト

**メトリクス計算**:
- Coverage: Vision → Objective → Deliverable の連携率
- Balance: 進捗の標準偏差（偏りが小さいほど高スコア）
- Overall Health: Coverage * 0.6 + Balance * 0.4

### Timeline View (P2)

計画 vs 実績の時間的乖離を可視化するビュー。

**主要コンポーネント**:
- **TimelineScale**: W/M/Q スケール切替ヘッダー
- **TimelineBar**: 計画（薄色）と実績（濃色）の 2 段バー

**ステータス判定**:
- ON TRACK: 計画通り進行
- DELAYED: 遅延あり
- AHEAD: 計画より前倒し
- COMPLETED: 完了

### Density View (P3)

作業量の分布をヒートマップで可視化するビュー。

**主要コンポーネント**:
- **HeatmapGrid**: CSS Grid によるセル配置
- 色は進捗率に基づく（赤 → 黄 → 緑）

**サイズ指標**:
- Tasks: タスク数
- Hours: 工数（タスク数 * 4 で概算）

## 技術スタック

| 項目 | 採用技術 |
|------|----------|
| フレームワーク | SvelteKit + Svelte 5 |
| 構文 | $state, $derived, $props, $effect |
| スタイリング | CSS 変数（Factorio テーマ） |
| 状態管理 | Svelte Stores |
| 描画 | Svelte + CSS（D3.js 不使用） |

## ファイル構成

```
zeus-dashboard/src/lib/viewer/wbs/
├── WBSViewer.svelte              # メインコンテナ
├── health/
│   ├── HealthView.svelte
│   ├── MetricsPanel.svelte
│   ├── MetricCard.svelte
│   └── ObjectiveList.svelte
├── timeline/
│   ├── TimelineView.svelte
│   ├── TimelineScale.svelte
│   └── TimelineBar.svelte
├── density/
│   ├── DensityView.svelte
│   └── HeatmapGrid.svelte
├── shared/
│   ├── ProgressBar.svelte
│   └── StatusBadge.svelte
├── stores/
│   └── wbsStore.ts
├── EntityDetailPanel.svelte
└── WBSSummaryBar.svelte
```

## 使用方法

### ビュー切替

WBSViewer のタブをクリックして切り替え:
- Health（デフォルト）
- Timeline
- Density

### エンティティ選択

各ビューで項目をクリックすると:
1. 選択状態が wbsStore に保存
2. EntityDetailPanel が表示
3. 他ビューに切り替えても選択状態を維持

### 階層展開（Health View）

- ▶ をクリックで展開
- ▼ をクリックで折りたたみ
- 配下の Deliverable が表示される

## API

既存の `/api/wbs-aggregated` を使用。追加 API は不要。

```typescript
// レスポンス例
{
  progress: {
    objectives: [
      { id: "obj-001", title: "MVP", progress: 75, children: [...] }
    ],
    total_progress: 68
  },
  coverage: {
    coverage_score: 78
  },
  issues: {
    total_issues: 3
  }
}
```

## Factorio テーマ

CSS 変数を一貫して使用:

```css
--bg-primary: #1a1a1a;
--bg-secondary: #242424;
--bg-panel: #2d2d2d;
--accent-primary: #ff9533;
--border-metal: #4a4a4a;
--status-good: #44cc44;
--status-fair: #ffcc00;
--status-poor: #ee4444;
```

## 関連ドキュメント

- [Zeus ダッシュボード概要](../../.claude/rules/dashboard.md)
- [Zeus システム設計](../../docs/system-design.md)
