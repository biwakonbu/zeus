# Affinity Canvas 設計書

## 概要

WBS Health ビューを刷新し、機能間の「横のつながり」を可視化する新しいビュー。

### 背景

現在の Health ビューの問題点：
- タスクがリストとして並んでいるだけ
- 選択時のサイドペインが情報の羅列
- 「神の視点」の哲学が欠如

### 目的

- 機能間の関連性（弱い依存、参照、影響）を可視化
- 機能軸・カテゴリ軸での隣接関係を表現
- 選択時に関連機能が浮かび上がる体験を提供

---

## 設計方針

| 項目 | 決定 |
|------|------|
| ビュー統合 | **レイヤード** - 同一キャンバスに重ね合わせ |
| 配置 | **フォースダイレクテッド** - 関連が強いものが自然に近づく |
| API | **バックエンド計算** - 類似度マトリクスを事前計算 |
| パラメータ | **自動推定** - プロジェクト特性から自動計算 |

---

## レイヤー構成

```
┌─────────────────────────────────────────────────────────┐
│  Layer 3: Impact Ripple（選択時のみ表示）              │
│  - 波紋アニメーション                                   │
│  - 影響範囲のハイライト                                │
├─────────────────────────────────────────────────────────┤
│  Layer 2: Affinity Edges（常時表示、透明度で強弱）     │
│  - 兄弟関係: オレンジ破線                              │
│  - 参照関係: 青い点線                                  │
│  - カテゴリ類似: 薄いグレー                            │
├─────────────────────────────────────────────────────────┤
│  Layer 1: Nodes（フォースダイレクテッド配置）          │
│  - 親子関係の引力: 強                                  │
│  - 兄弟関係の引力: 中                                  │
│  - カテゴリ類似の引力: 弱                              │
└─────────────────────────────────────────────────────────┘
```

---

## 関連タイプと重み付け

### 関連タイプ

| タイプ | 説明 | 視覚表現 |
|--------|------|----------|
| parent-child | 直接の親子関係 | 太い実線（白） |
| sibling | 同じ親を持つ兄弟 | オレンジ破線 |
| wbs-adjacent | WBSコードが隣接（1.1 と 1.2） | 薄いオレンジ点線 |
| reference | 参照関係（Quality→Deliverable等） | 青い点線 |
| category | 同じカテゴリに属する | 薄いグレー |

### 自動重み計算

```go
// プロジェクト特性を分析して重みを自動調整
func CalculateWeights(project *Project) AffinityWeights {
    // 参照関係が多い → reference の重みを上げる
    refRatio := countReferences(project) / totalEntities(project)

    // WBS が深い → wbs-adjacent の重みを上げる
    wbsDepth := maxWBSDepth(project)

    // 兄弟が多い → sibling の重みを下げる（差別化のため）
    avgSiblings := averageSiblingCount(project)

    return AffinityWeights{
        ParentChild: 1.0,                        // 常に最強（固定）
        Sibling:     0.7 - (avgSiblings * 0.05), // 0.5-0.8
        WBSAdjacent: 0.3 + (wbsDepth * 0.1),     // 0.3-0.6
        Reference:   0.4 + (refRatio * 0.3),     // 0.4-0.7
        Category:    0.3,                         // ベース（固定）
    }
}
```

---

## インタラクション設計

| 状態 | 表示 |
|------|------|
| **デフォルト** | フォース配置。関連が強いノードが自然にクラスタを形成 |
| **ホバー** | そのノードへの関連線がハイライト（他は透明度を下げる） |
| **クリック** | Impact Ripple 発動。1st/2nd/3rd リングが波紋で広がる |
| **ドラッグ** | ノードを手動配置可。離すとゆっくり元の位置へ戻る |
| **ダブルクリック** | 詳細パネルを開く |

### Impact Ripple 詳細

選択したノードを中心に、関連度に応じて3段階のリングで表示：

- **1st Ring**: 直接関連（親子、兄弟）
- **2nd Ring**: 間接関連（参照、WBS隣接）
- **3rd Ring**: 弱い関連（カテゴリ）

波紋アニメーションで広がり、関連ノードがハイライトされる。

---

## API 設計

### エンドポイント

```
GET /api/affinity
```

### レスポンス

```json
{
  "nodes": [
    {
      "id": "obj-023",
      "title": "Webダッシュボード",
      "type": "objective",
      "wbs_code": "3.4",
      "progress": 100,
      "status": "not_started"
    }
  ],
  "edges": [
    {
      "source": "del-018",
      "target": "del-019",
      "score": 0.85,
      "types": ["sibling", "category"],
      "reason": "同じ obj-023 に属する"
    }
  ],
  "clusters": [
    {
      "id": "cluster-1",
      "name": "可視化機能",
      "members": ["del-018", "del-019", "del-020", "del-021"]
    }
  ],
  "weights": {
    "parent_child": 1.0,
    "sibling": 0.65,
    "wbs_adjacent": 0.55,
    "reference": 0.60,
    "category": 0.30
  },
  "stats": {
    "total_nodes": 73,
    "total_edges": 156,
    "cluster_count": 8,
    "avg_connections": 4.3
  }
}
```

---

## ビジュアルイメージ

```
┌─────────────────────────────────────────────────────────────┐
│  ZEUS AFFINITY CANVAS                    [Filter▼] [Reset]  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│         ┌─────┐                                            │
│    ┌────┤obj-9├────┐         ┌─────┐                       │
│    │    └──┬──┘    │    ┌────┤obj-7├────┐                  │
│    │       │       │    │    └──┬──┘    │                  │
│  ┌─┴─┐   ┌─┴─┐   ┌─┴─┐  │    ┌─┴─┐    ┌─┴─┐               │
│  │023│───│024│   │025│  │    │013│    │014│               │
│  │ ★ │   │   │   │   │  │    │   │    │   │               │
│  └─┬─┘   └───┘   └───┘  │    └───┘    └───┘               │
│    │                     │                                  │
│  ┌─┴─┐ ┌───┐ ┌───┐      │       [クラスタ: AI機能]         │
│  │018│─│019│─│020│      │         ┌───┐ ┌───┐             │
│  │   │ │   │ │   │      │         │017│─│018│             │
│  └───┘ └───┘ └───┘      │         └───┘ └───┘             │
│   [クラスタ: 可視化]     │                                  │
│                          │                                  │
├─────────────────────────────────────────────────────────────┤
│  選択中: obj-023 Webダッシュボード                          │
│  直接関連: 4件  間接関連: 8件  影響範囲: 12件              │
└─────────────────────────────────────────────────────────────┘
```

---

## 実装フェーズ

### Phase 1: バックエンド API

| タスク | ファイル | 内容 |
|--------|----------|------|
| 1.1 | `internal/analysis/affinity.go` | AffinityCalculator 実装 |
| 1.2 | `internal/analysis/affinity.go` | 重み自動計算ロジック |
| 1.3 | `internal/analysis/affinity.go` | 関連検出（兄弟、WBS、参照、カテゴリ） |
| 1.4 | `internal/dashboard/handlers.go` | /api/affinity エンドポイント |
| 1.5 | `internal/analysis/affinity_test.go` | テスト |

### Phase 2: フロントエンド

| タスク | ファイル | 内容 |
|--------|----------|------|
| 2.1 | `zeus-dashboard/src/lib/viewer/engine/ForceLayoutEngine.ts` | フォースダイレクテッドレイアウト |
| 2.2 | `zeus-dashboard/src/lib/viewer/rendering/AffinityEdge.ts` | 関連線描画 |
| 2.3 | `zeus-dashboard/src/lib/viewer/effects/RippleEffect.ts` | Impact Ripple アニメーション |
| 2.4 | `zeus-dashboard/src/lib/api/affinity.ts` | API クライアント |
| 2.5 | `zeus-dashboard/src/lib/types/affinity.ts` | 型定義 |

### Phase 3: UI 統合

| タスク | ファイル | 内容 |
|--------|----------|------|
| 3.1 | `zeus-dashboard/src/lib/viewer/affinity/AffinityCanvas.svelte` | メインコンポーネント |
| 3.2 | - | インタラクション実装（ホバー、クリック、ドラッグ） |
| 3.3 | - | WBS View への統合（タブ追加） |

### Phase 4: 仕上げ

| タスク | 内容 |
|--------|------|
| 4.1 | パフォーマンス最適化（100+ノード対応） |
| 4.2 | LOD 実装（ズームレベル対応） |
| 4.3 | ドキュメント更新 |

---

## 依存関係

```
Phase 1.1-1.3 (AffinityCalculator)
    ↓
Phase 1.4 (API)
    ↓
Phase 2.1 (ForceLayout) ←── 並行可能 ──→ Phase 2.2-2.3 (Edge, Ripple)
    ↓                                           ↓
Phase 2.4-2.5 (API Client, Types)
    ↓
Phase 3.1 (AffinityCanvas)
    ↓
Phase 3.2 (Interaction) → Phase 3.3 (Integration)
    ↓
Phase 4 (Polish)
```

---

## ファイル構成

```
internal/
├── analysis/
│   ├── affinity.go          # NEW: 類似度計算
│   └── affinity_test.go     # NEW: テスト
└── dashboard/
    └── handlers.go          # 修正: /api/affinity 追加

zeus-dashboard/src/lib/
├── api/
│   └── affinity.ts          # NEW: API クライアント
├── viewer/
│   ├── engine/
│   │   └── ForceLayoutEngine.ts  # NEW
│   ├── rendering/
│   │   └── AffinityEdge.ts       # NEW
│   ├── effects/
│   │   └── RippleEffect.ts       # NEW
│   └── affinity/
│       ├── AffinityCanvas.svelte # NEW
│       └── index.ts              # NEW
└── types/
    └── affinity.ts          # NEW: 型定義
```

---

## 参考: ラウンドテーブル議論の経緯

### 議論テーマ

「WBS の Health 画面を既存の固定概念に捉われず、神の視点を表現する」

### 検討したアプローチ

1. **生態系メタファー** - プロジェクトを生命体として表現（却下: 抽象的すぎる）
2. **時空間マップ** - 過去-現在-未来の統合（一部採用: トレンド表示に活用可能）
3. **ミッションコントロール** - NASA風監視システム（却下: 現状と大差ない）

### 採用したアプローチ

- **Proximity Map** + **Affinity View** + **Impact Ripple** の複合

### 決定事項

1. レイヤード統合（シームレスな体験）
2. フォースダイレクテッド配置（関連が強いものが自然に近づく）
3. バックエンドAPI（類似度の事前計算）
4. 自動パラメータ推定（プロジェクト特性から計算）

---

## 更新履歴

| 日付 | 内容 |
|------|------|
| 2026-01-20 | 初版作成（ラウンドテーブル議論の成果物） |
