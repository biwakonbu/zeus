# Zeus Dashboard - SvelteKit フロントエンド

Zeus の Web ダッシュボード。SvelteKit + TypeScript で実装。Factorio 風インダストリアル UI と SSE リアルタイム更新を採用。

## 技術スタック

| 技術 | バージョン | 用途 |
|------|-----------|------|
| SvelteKit | 2.x | フレームワーク |
| Svelte | 5.x | UI ライブラリ（Runes 対応） |
| TypeScript | 5.x | 型システム |
| Vite | 7.x | ビルドツール |
| Mermaid.js | 11.x | 依存関係グラフ |
| adapter-static | 3.x | SPA ビルド |

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
│   │   │   ├── panels/       # 各パネル
│   │   │   ├── ui/           # 共通 UI
│   │   │   └── graph/        # Mermaid グラフ
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

### パネル

| コンポーネント | 説明 |
|---------------|------|
| OverviewPanel | プロジェクト概要（名前、説明、健全性、進捗） |
| StatsPanel | タスク統計（完了/進行中/保留/ブロック） |
| TasksPanel | タスク一覧テーブル |
| GraphPanel | Mermaid.js 依存関係グラフ |
| PredictionPanel | 予測分析（完了日、リスク、ベロシティ） |

### 共通 UI

| コンポーネント | 説明 |
|---------------|------|
| Panel | パネルコンテナ（金属フレーム効果） |
| Badge | ステータスバッジ |
| ProgressBar | プログレスバー |
| Stat | 統計アイテム |
| Table | Factorio 風テーブル |

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

| エンドポイント | 説明 |
|---------------|------|
| `GET /api/status` | プロジェクト状態 |
| `GET /api/tasks` | タスク一覧 |
| `GET /api/graph` | 依存関係グラフ（Mermaid 形式） |
| `GET /api/predict` | 予測分析結果 |
| `GET /api/events` | SSE ストリーム |

### SSE イベント

| イベント | 説明 |
|---------|------|
| `connected` | 接続確立 |
| `status` | プロジェクト状態更新 |
| `task` | タスク更新 |
| `graph` | グラフ更新 |
| `prediction` | 予測更新 |

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

| コマンド | 説明 |
|---------|------|
| `npm run dev` | 開発サーバー起動（ポート 5173） |
| `npm run build` | 本番ビルド |
| `npm run preview` | ビルド成果物プレビュー |
| `npm run check` | TypeScript 型チェック |
| `npm run check:watch` | 型チェック（監視モード） |
| `npm run clean` | ビルド成果物削除 |

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
