# 設計書 - WBS Dashboard Design Improvement

## アーキテクチャ概要

```
+---------------------------------------------------------------------+
|                          WBSViewer.svelte                           |
+---------------------------------------------------------------------+
|  [Health]    [Timeline]    [Density]        (Tab Navigation)        |
+---------------------------------------------------------------------+
|                                                                     |
|  +-----------------------------------+ +---------------------------+ |
|  |                                   | |                           | |
|  |       Active View Area            | |   EntityDetailPanel       | |
|  |                                   | |                           | |
|  |   - HealthView.svelte             | |   (Shown on selection)    | |
|  |   - TimelineView.svelte           | |                           | |
|  |   - DensityView.svelte            | |                           | |
|  |                                   | |                           | |
|  +-----------------------------------+ +---------------------------+ |
+---------------------------------------------------------------------+
|                         WBSSummaryBar.svelte                        |
+---------------------------------------------------------------------+
```

## コンポーネント仕様

### 1. WBSViewer.svelte

メインコンテナ。3 ビューの切り替えとデータ取得を管理。

```typescript
interface Props {
  onNodeSelect?: (nodeId: string, nodeType: string) => void;
}

// 内部状態
let activeView: ViewTab = $state('health');
let aggregatedData: WBSAggregatedResponse | null = $state(null);
let loading = $state(true);
let error: string | null = $state(null);
let showDetailPanel = $state(false);
```

### 2. HealthView.svelte

```typescript
interface Props {
  data: WBSAggregatedResponse | null;
  onNodeSelect: (nodeId: string, nodeType: string) => void;
}

// 派生状態
const coverage = $derived(data?.coverage?.coverage_score ?? 0);
const objectives = $derived(data?.progress?.objectives ?? []);
const balance = $derived(calculateBalance(objectives));
const overallHealth = $derived(Math.round(coverage * 0.6 + balance * 0.4));
```

### 3. TimelineView.svelte

```typescript
interface Props {
  data: WBSAggregatedResponse | null;
  onNodeSelect: (nodeId: string, nodeType: string) => void;
}

// 内部状態
let scale: 'week' | 'month' | 'quarter' = $state('month');

// 派生状態
const objectives = $derived(data?.progress?.objectives ?? []);
const timelineRange = $derived(calculateTimelineRange(objectives));
```

### 4. DensityView.svelte

```typescript
interface Props {
  data: WBSAggregatedResponse | null;
  onNodeSelect: (nodeId: string, nodeType: string) => void;
}

// 内部状態
let sizeMetric: 'tasks' | 'hours' = $state('tasks');

// 派生状態
const items = $derived<DensityItem[]>(
  (data?.progress?.objectives ?? []).map((obj) => ({
    id: obj.id,
    title: obj.title,
    taskCount: obj.children_count,
    progress: obj.progress
  }))
);
```

## 状態管理

### wbsStore.ts

```typescript
import { writable, derived } from 'svelte/store';

// 選択状態
export const selectedEntityId = writable<string | null>(null);
export const selectedEntityType = writable<string | null>(null);

// 展開状態
export const expandedIds = writable<Set<string>>(new Set());

// アクティブビュー
export type ViewType = 'health' | 'timeline' | 'density';
export const activeView = writable<ViewType>('health');

// 派生状態
export const hasSelection = derived(selectedEntityId, ($id) => $id !== null);

// アクション
export function selectEntity(id: string | null, type: string | null = null): void;
export function clearSelection(): void;
export function toggleExpand(id: string): void;
export function expand(id: string): void;
export function collapse(id: string): void;
export function expandAll(ids: string[]): void;
export function collapseAll(): void;
export function setActiveView(view: ViewType): void;
```

## メトリクス計算

### Coverage（網羅度）

Vision → Objective → Deliverable の連携率を計算。API から取得した `coverage_score` を使用。

### Balance（バランス）

進捗の均一度を標準偏差ベースで計算:

```typescript
function calculateBalance(objs: ProgressNode[]): number {
  if (objs.length === 0) return 0;
  const progresses = objs.map((o) => o.progress);
  const mean = progresses.reduce((a, b) => a + b, 0) / progresses.length;
  const variance =
    progresses.reduce((sum, p) => sum + Math.pow(p - mean, 2), 0) / progresses.length;
  const stdDev = Math.sqrt(variance);
  // stdDev が 50 以上なら 0、0 なら 100
  return Math.max(0, Math.min(100, Math.round(100 - stdDev * 2)));
}
```

### Overall Health（総合健全性）

Coverage と Balance の加重平均:

```typescript
const overallHealth = Math.round(coverage * 0.6 + balance * 0.4);
```

## スタイリング

### 共通プログレスバー

```css
.progress-bar {
  position: relative;
  background: var(--bg-secondary, #242424);
  border: 1px solid var(--border-metal, #4a4a4a);
  border-radius: 2px;
  overflow: hidden;
}

.progress-bar__fill--low { background: var(--status-poor, #ee4444); }
.progress-bar__fill--mid { background: var(--status-fair, #ffcc00); }
.progress-bar__fill--high { background: var(--status-good, #44cc44); }
```

### 階層リストアイテム

```css
.objective-item {
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--border-dark, #333333);
  transition: background-color 0.15s ease;
}

.objective-item:hover {
  background-color: var(--bg-hover, #3a3a3a);
}

.objective-item.selected {
  background-color: var(--bg-secondary, #242424);
  border-left: 3px solid var(--accent-primary, #ff9533);
}
```

### ヒートマップセル

```css
.heatmap-cell {
  aspect-ratio: 1;
  padding: 12px;
  background-color: var(--bg-panel, #2d2d2d);
  border: 2px solid var(--border-metal, #4a4a4a);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
}

.heatmap-cell::before {
  content: '';
  position: absolute;
  inset: 0;
  background-color: var(--cell-color);
  opacity: var(--cell-opacity);
  z-index: 0;
}

.heatmap-cell:hover {
  border-color: var(--accent-primary, #ff9533);
  transform: scale(1.02);
}

.heatmap-cell.selected {
  border-color: var(--accent-primary, #ff9533);
  box-shadow: 0 0 12px rgba(255, 149, 51, 0.4);
}
```

## アクセシビリティ

### キーボードナビゲーション

- `Tab`: フォーカス移動
- `Enter` / `Space`: 選択 / 展開トグル
- `aria-label`: スクリーンリーダー対応
- `aria-expanded`: 展開状態の通知
- `aria-pressed`: 選択状態の通知

### 色コントラスト

Factorio テーマの CSS 変数は WCAG AA 準拠のコントラスト比を確保。

## 将来の拡張

1. **Momentum メトリクス**: 履歴機能実装後に追加
2. **リアルタイム更新**: SSE 対応
3. **Timeline 実データ連携**: actual_start, actual_end, estimated_hours
4. **レスポンシブ対応**: モバイル表示の最適化
