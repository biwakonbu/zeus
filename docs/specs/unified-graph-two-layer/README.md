# UnifiedGraph 2層モデル仕様

> 文書メタデータ
> - 文書種別: 仕様
> - 実装状態: 完了
> - 正本ソース: `cmd/graph.go`, `internal/analysis/unified_graph.go`, `internal/dashboard/handlers_unified.go`
> - 最終検証日: `2026-02-07`
## 概要

UnifiedGraph は `structural` と `reference` の 2 層で関係を分離する。
本仕様は CLI・Dashboard・API の共通契約を定義する。

## 参照バージョン

- Zeus: 現行 main（2026-02-07 時点）
- Go: `1.24.x`（`go.mod` 基準）
- Dashboard: Svelte `5.45.x`（`package.json` 基準）

## データモデル

### レイヤー

- `structural`: 階層・実装・貢献など、構造を形成する関係
- `reference`: 依存・参照など、横断参照を表す関係

### relation と許容行列

| relation | layer | from | to |
|----------|-------|------|----|
| `parent` | `structural` | `objective` | `objective` |
| `depends_on` | `reference` | `activity` | `activity` |
| `implements` | `structural` | `activity` | `usecase` |
| `contributes` | `structural` | `usecase` | `objective` |

### 方向規約

- `structural` は **child -> parent** の向きで統一
- `reference` は関係意味に従う向きで保持

## CLI 仕様（`zeus graph --unified`）

### オプション

- `--layers structural,reference`
- `--relations parent,depends_on,implements,contributes`
- `--types activity,usecase,objective`
- `--focus <id>` + `--depth <n>`（`--depth` 未指定時は `3`）
- `--hide-completed`
- `--hide-draft`

### 出力仕様

- Text: `Structural Relations` と `Reference Relations` を分離表示
- DOT/Mermaid: relation/layer に応じた線種・色を分離

## API 仕様（`GET /api/unified-graph`）

### 破壊的変更

| 旧 | 新 |
|----|----|
| `edges[*].type` | `edges[*].layer` |
| `edges[*].label` | `edges[*].relation` |
| `nodes[*].depth` | `nodes[*].structural_depth` |
| `nodes[*].parents` | `nodes[*].structural_parents` |
| `nodes[*].children` | `nodes[*].structural_children` |
| `stats.max_depth` | `stats.max_structural_depth` |
| `stats.edges_by_type` | `stats.edges_by_layer`, `stats.edges_by_relation` |

### クエリ

- `focus`, `depth`（未指定時 `3`）, `types`, `layers`, `relations`, `hide-completed`, `hide-draft`

## Dashboard 仕様

### レイアウト

- 深さ決定は `structural_depth` を優先し、欠損時のみ `structural` から層計算
- ノード座標は 50px グリッドにスナップ（`LAYOUT_GRID_UNIT=50`）
- 層内順序は barycenter sweep（重み: `structural=1.0`, `reference=0.35`）
- エッジは 10px サブグリッド（`EDGE_ROUTING_GRID_UNIT=10`）で直交配線
- 接点はノード辺に対して垂直（port normal に沿った stub を使用）
- グループ境界は `structural` の無向連結成分単位（ラベルは代表ノード `title`）

### 依存フィルター（Alt+クリック / 右クリック）

- 対象ノードと関連ノードのみを表示
- 関連ノード集合は、探索対象エッジ上の **無向連結成分** として算出
- 既定探索対象: `reference` のみ
- `reference` で関連 0 件の場合、`structural` をフォールバック探索
- `Include structural edges in impact filter` ON で `structural` も常時探索
- 同一ノードを再度 Alt+クリック/右クリックするとフィルター解除

## 非互換ポリシー

- 互換レイヤーは提供しない
- 旧フィールドは返却しない
- フロント・API・CLI を同時更新する

## 実装参照

- `internal/analysis/types.go`
- `internal/analysis/unified_graph.go`
- `internal/dashboard/handlers_unified.go`
- `cmd/graph.go`
- `zeus-dashboard/src/lib/types/api.ts`
- `zeus-dashboard/src/lib/viewer/FactorioViewer.svelte`
- `zeus-dashboard/src/lib/viewer/engine/LayoutEngine.ts`

*更新日: 2026-02-10（Deliverable削除・SimpleMode廃止対応）*
