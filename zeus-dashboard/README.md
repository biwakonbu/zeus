# Zeus Dashboard - SvelteKit フロントエンド

Zeus の Web ダッシュボード。SvelteKit + TypeScript で実装。Factorio 風インダストリアル UI と SSE リアルタイム更新を採用。

## 技術スタック

| 技術           | バージョン | 用途                        |
| -------------- | ---------- | --------------------------- |
| SvelteKit      | 2.x        | フレームワーク              |
| Svelte         | 5.x        | UI ライブラリ（Runes 対応） |
| TypeScript     | 5.x        | 型システム                  |
| Vite           | 7.x        | ビルドツール                |
| Mermaid.js     | 11.x       | 依存関係グラフ              |
| adapter-static | 3.x        | SPA ビルド                  |
| oxlint         | 1.x        | TypeScript Linter（高速）   |
| ESLint         | 9.x        | Svelte Linter               |
| Prettier       | 3.x        | コードフォーマッター        |

## セットアップ

```bash
# 依存関係インストール
npm install

# 開発サーバー起動（ポート 5173）
npm run dev

# 型チェック
npm run check
```

## 開発ワークフロー

### 開発モード

**ターミナル 1 - Go バックエンド:**

```bash
# プロジェクトルートで実行
go run . dashboard --dev --port 8080
```

**ターミナル 2 - Vite 開発サーバー:**

```bash
cd zeus-dashboard
npm run dev
```

ブラウザで `http://localhost:5173` にアクセス。HMR（Hot Module Replacement）が有効。

### 本番ビルド

```bash
# プロジェクトルートで実行
make build-all
```

これにより:

1. SvelteKit をビルド（`zeus-dashboard/build/`）
2. ビルド成果物を Go embed 用にコピー（`internal/dashboard/build/`）
3. Go バイナリをビルド

## ディレクトリ構造

```
zeus-dashboard/
├── src/
│   ├── lib/
│   │   ├── api/              # API クライアント
│   │   │   ├── client.ts     # Fetch ベース API
│   │   │   └── sse.ts        # SSE クライアント（自動再接続）
│   │   ├── stores/           # Svelte ストア
│   │   │   ├── connection.ts # 接続状態管理
│   │   │   ├── status.ts     # プロジェクト状態
│   │   │   ├── tasks.ts      # タスク一覧
│   │   │   ├── graph.ts      # 依存関係グラフ
│   │   │   └── prediction.ts # 予測分析
│   │   ├── components/       # UI コンポーネント
│   │   │   ├── layout/       # Header, Footer
│   │   │   ├── viewer/       # Factorio 風ビューワー
│   │   │   ├── ui/           # 共通 UI
│   │   ├── theme/            # Factorio デザインシステム
│   │   │   ├── variables.css # CSS 変数
│   │   │   └── factorio.css  # グローバルスタイル
│   │   └── types/            # TypeScript 型定義
│   │       └── api.ts
│   ├── routes/
│   │   ├── +layout.svelte    # ルートレイアウト
│   │   └── +page.svelte      # メインページ
│   └── app.html              # HTML テンプレート
├── static/
│   └── fonts/                # IBM Plex Mono フォント
├── svelte.config.js          # SvelteKit 設定
├── vite.config.ts            # Vite 設定（API プロキシ）
├── tsconfig.json             # TypeScript 設定
└── package.json
```

## コンポーネント

### Factorio 風ビューワー

| コンポーネント | 説明                                          |
| -------------- | --------------------------------------------- |
| FactorioViewer | メインビューワーコンポーネント（PixiJS 描画） |
| ViewerEngine   | PixiJS 初期化・管理                           |
| LayoutEngine   | 自動レイアウト（50px 格子 + 構造連結成分境界） |
| OrthogonalRouter | 直交配線ルーター（10px サブグリッド）       |
| SpatialIndex   | Quadtree 空間インデックス                     |
| GraphNode      | タスクノード描画（LOD 対応）                  |
| GraphEdge      | エッジ描画（直交 polyline + flow dots）       |
| GraphGroupBoundary | 構造連結成分の境界描画                     |
| Minimap        | ミニマップ                                    |
| FilterPanel    | フィルターパネル                              |

### 共通 UI

| コンポーネント | 説明                               |
| -------------- | ---------------------------------- |
| Panel          | パネルコンテナ（金属フレーム効果） |
| Badge          | ステータスバッジ                   |
| ProgressBar    | プログレスバー                     |

## Svelte 5 Runes

このプロジェクトは Svelte 5 の Runes を使用:

```svelte
<script lang="ts">
	// リアクティブな状態
	let count = $state(0);

	// 派生値
	let doubled = $derived(count * 2);

	// 副作用
	$effect(() => {
		console.log(`Count is now ${count}`);
	});

	// Props
	let { title } = $props<{ title: string }>();
</script>
```

## API 連携

### エンドポイント

| エンドポイント        | 説明                           |
| --------------------- | ------------------------------ |
| `GET /api/status`     | プロジェクト状態               |
| `GET /api/activities` | Activity 一覧                  |
| `GET /api/unified-graph` | Unified Graph（2層モデル） |
| `GET /api/graph`      | 依存関係グラフ（Mermaid, 旧互換） |
| `GET /api/events`     | SSE ストリーム                 |

### Unified Graph（2層モデル）

- `layer`: `structural`
- `relation`: `parent` / `implements`
- Graph View のノード配置は `LAYOUT_GRID_UNIT=50` にスナップ
- エッジ配線は `EDGE_ROUTING_GRID_UNIT=10`（ノード格子の 1/5）で直交ルーティング
- エッジ接点はノード辺に対して垂直、流向は flow dots で可視化
- グループ境界は Objective ベースで描画（所属 UseCase/Activity を包含、ラベルは Objective `title`）
- `structural_depth` がある場合は深さ決定で優先、欠損時のみ構造層計算を使用
`/api/unified-graph` クエリ:

- `focus=<node_id>`
- `depth=<int>`（未指定時は `3`）
- `types=activity,usecase`
- `layers=structural`
- `relations=parent,implements`
- `group=<objective_id>`
- `hide-completed=true|false`
- `hide-draft=true|false`

### SSE イベント

| イベント     | 説明                 |
| ------------ | -------------------- |
| `connected`  | 接続確立             |
| `status`     | プロジェクト状態更新 |
| `task`       | タスク更新           |
| `graph`      | グラフ更新           |
| `prediction` | 予測更新             |

## デザインシステム

### CSS 変数

```css
:root {
	/* 背景色 */
	--bg-primary: #1a1a1a;
	--bg-secondary: #242424;
	--bg-panel: #2d2d2d;

	/* オレンジアクセント */
	--accent-primary: #ff9533;
	--accent-hover: #ffaa55;

	/* 金属フレーム */
	--border-metal: #4a4a4a;
	--border-highlight: #666666;

	/* ステータス色 */
	--status-good: #44cc44;
	--status-fair: #ffcc00;
	--status-poor: #ee4444;
}
```

### フォント

IBM Plex Mono を使用。`static/fonts/` に配置。

## npm スクリプト

| コマンド              | 説明                            |
| --------------------- | ------------------------------- |
| `npm run dev`         | 開発サーバー起動（ポート 5173） |
| `npm run build`       | 本番ビルド                      |
| `npm run preview`     | ビルド成果物プレビュー          |
| `npm run check`       | TypeScript 型チェック           |
| `npm run check:watch` | 型チェック（監視モード）        |
| `npm run lint`        | Lint 実行（oxlint + ESLint）    |
| `npm run lint:fix`    | Lint 自動修正                   |
| `npm run format`      | Prettier フォーマット           |
| `npm run clean`       | ビルド成果物削除                |

## Lint 構成

併用構成で高速かつ Svelte 対応の Lint を実現:

| ツール                      | 対象          | 役割                  |
| --------------------------- | ------------- | --------------------- |
| oxlint                      | `.ts` ファイル | TypeScript Lint（高速） |
| ESLint + eslint-plugin-svelte | `.svelte` ファイル | Svelte 固有 Lint      |
| Prettier                    | 全ファイル    | フォーマット          |

**設定ファイル:**

- `.oxlintrc.json` - oxlint 設定（`.svelte` を除外）
- `eslint.config.js` - ESLint 設定（`.svelte` のみ対象）
- `.prettierrc` - Prettier 設定

```bash
# Lint 実行
npm run lint

# 自動修正
npm run lint:fix

# フォーマット
npm run format
```

## トラブルシューティング

### API 接続エラー

開発モードで API にアクセスできない場合:

1. Go サーバーが起動しているか確認（`--dev` フラグ必須）
2. ポート 8080 が使用可能か確認
3. `vite.config.ts` のプロキシ設定を確認

### SSE 接続が切断される

- 自動再接続は最大 10 回まで
- 再接続失敗時は 5 秒間隔のポーリングにフォールバック
- Header の接続状態インジケータで確認可能

### Mermaid グラフが表示されない

- ブラウザコンソールでエラーを確認
- グラフデータが空でないか確認（`/api/graph` レスポンス）
