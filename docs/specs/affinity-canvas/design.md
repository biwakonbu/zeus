# Affinity Canvas 設計書

## アーキテクチャ概要

```
+------------------------------------------------------------------+
|                      Affinity Canvas                              |
+------------------------------------------------------------------+
|  Frontend (SvelteKit + SVG)                                       |
|  +----------------+  +----------------+  +----------------+       |
|  | AffinityView   |  | Force Layout   |  | Interaction    |       |
|  | .svelte        |  | (内蔵)         |  | Handlers       |       |
|  +-------+--------+  +-------+--------+  +-------+--------+       |
|          |                   |                   |                |
|  +-------v---------+  +------v--------+  +-------v--------+       |
|  | SVG Canvas      |  | Node/Edge     |  | Pan/Zoom       |       |
|  | (viewBox制御)   |  | 描画          |  | Drag           |       |
|  +-----------------+  +---------------+  +----------------+       |
+------------------------------------------------------------------+
|  API Layer                                                        |
|  +--------------------------------------------------------------+ |
|  | fetchAffinity() -> GET /api/affinity                         | |
|  +--------------------------------------------------------------+ |
+------------------------------------------------------------------+
|  Backend (Go)                                                     |
|  +----------------+  +----------------+  +----------------+       |
|  | handlers.go    |  | affinity.go    |  | bottleneck.go  |       |
|  | /api/affinity  |  | Calculator     |  | 共有型定義     |       |
|  +----------------+  +----------------+  +----------------+       |
+------------------------------------------------------------------+
```

---

## バックエンド設計

### 1. AffinityCalculator

**ファイル**: `internal/analysis/affinity.go`

```go
// AffinityCalculator は類似度を計算
type AffinityCalculator struct {
    vision       VisionInfo
    objectives   []ObjectiveInfo
    deliverables []DeliverableInfo
    tasks        []TaskInfo
    quality      []QualityInfo
    risks        []RiskInfo
}

// Calculate はアフィニティを計算
func (ac *AffinityCalculator) Calculate(ctx context.Context) (*AffinityResult, error)
```

**主要メソッド:**

| メソッド | 役割 |
|---------|------|
| `buildNodes()` | 全エンティティからノードを構築 |
| `detectParentChild()` | 親子関係を検出 |
| `detectSibling()` | 兄弟関係を検出 |
| `detectWBSAdjacent()` | WBS 隣接関係を検出 |
| `detectReference()` | Quality/Risk からの参照関係を検出 |
| `CalculateWeights()` | プロジェクト特性から重みを計算 |
| `buildClusters()` | Objective ベースでクラスタを構築 |

### 2. 関連検出アルゴリズム

**親子関係 (parent-child):**
- Vision -> トップレベル Objective
- Objective -> 子 Objective
- Objective -> Deliverable
- Deliverable -> Task (ParentID)

**兄弟関係 (sibling):**
- 同じ Objective に属する Deliverable 同士
- 同じ親を持つ Task 同士

**WBS 隣接 (wbs-adjacent):**
- WBS コードが連続するノード
- 例: `1.1` と `1.2`、`1.1.1` と `1.1.2`

**参照関係 (reference):**
- Quality -> Deliverable
- Risk -> Objective / Deliverable

### 3. 重み自動計算

```go
func (ac *AffinityCalculator) CalculateWeights() AffinityWeights {
    // 参照関係の比率
    refRatio := float64(refCount) / float64(totalEntities)

    // WBS の深さ
    maxDepth := maxWBSDepth()

    // 平均兄弟数
    avgSiblings := averageSiblingCount()

    return AffinityWeights{
        ParentChild: 1.0,                        // 常に最強
        Sibling:     0.7 - (avgSiblings * 0.05), // 0.5-0.8
        WBSAdjacent: 0.3 + (maxDepth * 0.1),     // 0.3-0.6
        Reference:   0.4 + (refRatio * 0.3),     // 0.4-0.7
        Category:    0.3,                         // 固定
    }
}
```

### 4. 共有型定義

`bottleneck.go` に共有型を配置:

```go
// RiskInfo はリスク情報（ボトルネック分析・アフィニティ分析用）
type RiskInfo struct {
    ID            string
    Title         string
    Probability   string
    Impact        string
    Score         int
    Status        string
    ObjectiveID   string   // Affinity 用
    DeliverableID string   // Affinity 用
}

// QualityInfo は Quality エンティティ情報
type QualityInfo struct {
    ID            string
    Title         string
    DeliverableID string
    Status        string
}
```

---

## フロントエンド設計

### 1. コンポーネント構成

**ファイル**: `zeus-dashboard/src/lib/viewer/wbs/affinity/AffinityView.svelte`

```svelte
<script lang="ts">
  // 状態管理
  let layoutNodes: LayoutNode[] = $state([]);
  let hoveredNodeId: string | null = $state(null);
  let selectedNodeId: string | null = $state(null);

  // ビューポート
  let viewBox = $state({ x: 0, y: 0, width: 800, height: 600 });
  let zoom = $state(1);

  // フィルター
  let showEdges = $state(true);
  let minEdgeScore = $state(0.3);
</script>
```

### 2. フォースレイアウト

**アルゴリズム:**
1. ノードを円形に初期配置
2. 反発力・引力・中心引力の 3 フォースを適用
3. `requestAnimationFrame` でアニメーション
4. 減衰係数 0.6 でスムーズに収束

**パラメータ:**

| パラメータ | 値 | 役割 |
|-----------|-----|------|
| 反発係数 | 500 | ノード間の反発力 |
| 引力係数 | 0.1 * score | エッジの引力（スコアに比例） |
| 中心引力 | 0.01 | 中心への引力 |
| 減衰係数 | 0.6 | 速度の減衰 |
| 最大イテレーション | 300 | 収束上限 |

### 3. インタラクション

**ドラッグ:**
- `fx`, `fy` でノードを固定
- マウスアップで固定解除、シミュレーション再開

**パン:**
- SVG `viewBox` を操作
- マウスダウン + ムーブで移動

**ズーム:**
- wheel イベントで 0.5x - 3x
- マウス位置を中心にズーム

**ホバー:**
- 関連エッジをハイライト
- ノードラベルを表示

### 4. UI コンポーネント

| 要素 | 説明 |
|------|------|
| ヘッダー | タイトル、エッジ表示トグル、スコアフィルター |
| キャンバス | SVG フォースグラフ |
| 統計パネル | ノード数、エッジ数、クラスタ数、平均接続数 |
| クラスタリスト | クラスタ名とメンバー数 |
| 凡例 | ノードタイプと色の対応 |

### 5. スタイリング

**ノードカラー:**

| タイプ | 色 |
|--------|-----|
| vision | #f59e0b (オレンジ) |
| objective | #3b82f6 (ブルー) |
| deliverable | #10b981 (グリーン) |
| task | #8b5cf6 (パープル) |

**ノードサイズ:**

| タイプ | サイズ |
|--------|--------|
| vision | 24px |
| objective | 18px |
| deliverable | 14px |
| task | 10px |

**エッジカラー:**

| タイプ | 色 |
|--------|-----|
| parent-child | #f59e0b |
| sibling | #3b82f6 |
| wbs-adjacent | #10b981 |
| reference | #ec4899 |
| category | #8b5cf6 |

---

## ファイル構成

```
internal/
├── analysis/
│   ├── affinity.go          # AffinityCalculator
│   └── bottleneck.go        # 共有型（RiskInfo, QualityInfo）
└── dashboard/
    ├── handlers.go          # handleGetAffinity
    └── server.go            # /api/affinity ルート

zeus-dashboard/src/lib/
├── api/
│   └── client.ts            # fetchAffinity()
├── types/
│   └── api.ts               # Affinity 型定義
└── viewer/
    ├── wbs/
    │   └── affinity/
    │       └── AffinityView.svelte
    └── WBSViewer.svelte     # タブ統合
```

---

## 設計判断

### SVG vs PixiJS

**選択: SVG**

| 観点 | SVG | PixiJS |
|------|-----|--------|
| 学習コスト | 低 | 高 |
| 保守性 | 高 | 中 |
| 既存コードとの一貫性 | 高 | 低 |
| パフォーマンス（100 ノード） | 十分 | 過剰 |
| パフォーマンス（1000+ ノード） | 要検討 | 優位 |

現時点では SVG で十分なパフォーマンスが得られるため、保守性を優先。

### ForceLayout の分離

**選択: コンポーネント内蔵**

- 単一コンポーネントでの実装がシンプル
- 再利用の必要性が低い
- 将来的に分離が必要になった場合は容易に抽出可能

---

## 関連ドキュメント

- [要件定義書](./requirements.md)
- [ダッシュボード規約](../../.claude/rules/dashboard.md)
