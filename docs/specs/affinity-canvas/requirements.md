# Affinity Canvas 要件定義書

## 概要

WBS ダッシュボードの 4 番目のビュー「Affinity Canvas」は、プロジェクト内のエンティティ（Vision, Objective, Deliverable, Task）間の関連性をフォースグラフで可視化する機能です。

**主な目的:**
- 機能間の横のつながりを発見
- 依存関係の全体像を把握
- 影響範囲の可視化

---

## 機能要件

### FR-1: Affinity API

| ID | 要件 | 実装状態 |
|----|------|----------|
| FR-1.1 | `/api/affinity` エンドポイントを提供 | 実装済み |
| FR-1.2 | ノード情報（id, title, type, wbs_code, progress, status）を返却 | 実装済み |
| FR-1.3 | エッジ情報（source, target, score, types, reason）を返却 | 実装済み |
| FR-1.4 | クラスタ情報（id, name, members）を返却 | 実装済み |
| FR-1.5 | 重み情報（parent_child, sibling, wbs_adjacent, reference, category）を返却 | 実装済み |
| FR-1.6 | 統計情報（total_nodes, total_edges, cluster_count, avg_connections）を返却 | 実装済み |

### FR-2: 関連検出

| ID | 要件 | 実装状態 |
|----|------|----------|
| FR-2.1 | 親子関係（parent-child）を検出 | 実装済み |
| FR-2.2 | 兄弟関係（sibling）を検出 | 実装済み |
| FR-2.3 | WBS 隣接関係（wbs-adjacent）を検出 | 実装済み |
| FR-2.4 | 参照関係（reference: Quality -> Deliverable 等）を検出 | 実装済み |
| FR-2.5 | カテゴリ類似（category）を検出 | 将来実装 |

### FR-3: 重み自動計算

| ID | 要件 | 実装状態 |
|----|------|----------|
| FR-3.1 | プロジェクト特性から重みを自動調整 | 実装済み |
| FR-3.2 | 参照関係が多い場合、reference の重みを上げる | 実装済み |
| FR-3.3 | WBS が深い場合、wbs-adjacent の重みを上げる | 実装済み |
| FR-3.4 | 兄弟が多い場合、sibling の重みを下げる | 実装済み |

### FR-4: フロントエンド - キャンバス表示

| ID | 要件 | 実装状態 |
|----|------|----------|
| FR-4.1 | フォースダイレクテッドレイアウトでノードを配置 | 実装済み（SVG） |
| FR-4.2 | 関連タイプに応じたスタイルでエッジを描画 | 実装済み |
| FR-4.3 | WBSViewer のタブとして統合 | 実装済み |
| FR-4.4 | 既存コンポーネントとデザイン統一 | 実装済み |

### FR-5: フロントエンド - インタラクション

| ID | 要件 | 実装状態 |
|----|------|----------|
| FR-5.1 | ホバー時に関連線をハイライト | 実装済み |
| FR-5.2 | クリック時にノードを選択 | 実装済み |
| FR-5.3 | ドラッグでノードを手動配置可能 | 実装済み |
| FR-5.4 | パン・ズーム操作 | 実装済み |

---

## 非機能要件

### NFR-1: パフォーマンス

| ID | 要件 | 目標値 | 実装状態 |
|----|------|--------|----------|
| NFR-1.1 | 初期レンダリング（100 ノード） | < 1000ms | 達成 |
| NFR-1.2 | フォースレイアウト安定化 | < 2000ms | 達成 |
| NFR-1.3 | ホバーレスポンス | < 50ms | 達成 |
| NFR-1.4 | アニメーション | 60fps | 達成 |

### NFR-2: デザイン

| ID | 要件 | 実装状態 |
|----|------|----------|
| NFR-2.1 | Factorio 風インダストリアルデザインに準拠 | 達成 |
| NFR-2.2 | Lucide Icons を使用（Unicode Emoji 禁止） | 達成 |
| NFR-2.3 | アニメーション duration は 200ms 以下 | 達成 |

### NFR-3: 保守性

| ID | 要件 | 実装状態 |
|----|------|----------|
| NFR-3.1 | テスト全パス（バックエンド） | 達成 |
| NFR-3.2 | TypeScript 型定義完備（フロントエンド） | 達成 |
| NFR-3.3 | コメントは日本語で記載 | 達成 |

---

## データモデル

### AffinityResponse

```typescript
interface AffinityResponse {
  nodes: AffinityNode[];
  edges: AffinityEdge[];
  clusters: AffinityCluster[];
  weights: AffinityWeights;
  stats: AffinityStats;
}
```

### AffinityNode

```typescript
interface AffinityNode {
  id: string;
  title: string;
  type: 'vision' | 'objective' | 'deliverable' | 'task';
  wbs_code: string;
  progress: number;
  status: string;
}
```

### AffinityEdge

```typescript
interface AffinityEdge {
  source: string;
  target: string;
  score: number;  // 0.0 - 1.0
  types: AffinityEdgeType[];
  reason: string;
}

type AffinityEdgeType =
  | 'parent-child'
  | 'sibling'
  | 'wbs-adjacent'
  | 'reference'
  | 'category';
```

### AffinityWeights

```typescript
interface AffinityWeights {
  parent_child: number;   // 常に 1.0
  sibling: number;        // 0.5 - 0.8
  wbs_adjacent: number;   // 0.3 - 0.6
  reference: number;      // 0.4 - 0.7
  category: number;       // 常に 0.3
}
```

---

## 関連ドキュメント

- [設計書](./design.md)
- [API 仕様](../../api-spec.md)
- [ダッシュボード規約](../../.claude/rules/dashboard.md)
