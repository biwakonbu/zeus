---
description: Zeus プロジェクトのディレクトリ構造。必要時に手動参照。
---

# プロジェクト構造

```
zeus/
├── cmd/                      # Cobra コマンド
│   ├── root.go               # ルートコマンド（DI対応）
│   ├── init.go               # zeus init
│   ├── status.go             # zeus status
│   ├── pending.go            # zeus pending
│   ├── approve.go            # zeus approve
│   ├── reject.go             # zeus reject
│   ├── snapshot.go           # zeus snapshot
│   ├── history.go            # zeus history
│   ├── graph.go              # zeus graph（依存関係グラフ）
│   ├── predict.go            # zeus predict（予測分析）
│   ├── report.go             # zeus report（レポート生成）
│   ├── dashboard.go          # zeus dashboard（Webダッシュボード）
│   └── ...
├── internal/                 # 内部パッケージ
│   ├── core/                 # コアロジック
│   │   ├── zeus.go           # メインロジック（DI対応、分析機能統合）
│   │   ├── interfaces.go     # FileStore, StateStore, ApprovalStore インターフェース
│   │   ├── entity.go         # EntityHandler, EntityRegistry
│   │   ├── activity_handler.go # Activity エンティティハンドラー
│   │   ├── subsystem_handler.go # サブシステムハンドラー
│   │   ├── types.go          # 型定義
│   │   ├── state.go          # 状態・スナップショット管理（Context対応）
│   │   ├── approval.go       # 3段階承認システム（Context対応）
│   │   ├── errors.go         # エラー定義
│   │   ├── lint.go           # Lint チェック（ID形式、status/progress整合性）
│   │   └── mocks/            # テスト用モック
│   │       └── mock_file_store.go
│   ├── analysis/             # 分析機能（Phase 4）
│   │   ├── types.go          # 分析用型定義（独立）
│   │   ├── graph.go          # 依存関係グラフ構築・可視化
│   │   ├── wbs.go            # WBS 階層構築
│   │   └── predict.go        # 予測分析（完了日、リスク、ベロシティ）
│   ├── report/               # レポート生成（Phase 4）
│   │   ├── generator.go      # レポート生成ロジック
│   │   └── templates.go      # 出力テンプレート（TEXT/HTML/Markdown）
│   ├── dashboard/            # Web ダッシュボード（Phase 5）
│   │   ├── server.go         # HTTP サーバー + SSE + CORS
│   │   ├── handlers.go       # API ハンドラー + SSE エンドポイント
│   │   ├── sse.go            # SSE Broadcaster
│   │   ├── build/            # SvelteKit ビルド成果物（embed）
│   │   └── dashboard_test.go
│   ├── yaml/                 # YAML 操作
│   │   ├── parser.go
│   │   ├── writer.go
│   │   ├── file_manager.go   # パス検証・セキュリティ（Context対応）
│   │   └── filelock.go       # ファイルロック機構
│   ├── doctor/               # 診断・修復（Context対応）
│   │   └── doctor.go
│   └── generator/            # Claude Code 連携ファイル生成（Context対応）
│       └── generator.go
├── zeus-dashboard/           # SvelteKit ダッシュボード
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/          # API クライアント + SSE
│   │   │   ├── stores/       # Svelte ストア
│   │   │   ├── components/   # UI コンポーネント
│   │   │   │   ├── layout/   # Header, Footer
│   │   │   │   └── ui/       # Badge, ProgressBar, Panel
│   │   │   ├── viewer/       # Factorio風ビューワー
│   │   │   │   ├── FactorioViewer.svelte  # メインコンポーネント
│   │   │   │   ├── engine/   # 描画エンジン
│   │   │   │   │   ├── ViewerEngine.ts    # PixiJS 初期化・管理
│   │   │   │   │   ├── LayoutEngine.ts    # 自動レイアウト
│   │   │   │   │   └── SpatialIndex.ts    # Quadtree 空間インデックス
│   │   │   │   ├── rendering/# 描画クラス
│   │   │   │   │   ├── TaskNode.ts        # ノード描画（LOD対応）
│   │   │   │   │   └── TaskEdge.ts        # エッジ描画
│   │   │   │   ├── interaction/# インタラクション
│   │   │   │   │   ├── SelectionManager.ts # 選択管理
│   │   │   │   │   └── FilterManager.ts    # フィルター管理
│   │   │   │   └── ui/       # UI コンポーネント
│   │   │   │       ├── Minimap.svelte      # ミニマップ
│   │   │   │       └── FilterPanel.svelte  # フィルターパネル
│   │   │   ├── theme/        # Factorio デザインシステム
│   │   │   └── types/        # TypeScript 型定義
│   │   └── routes/
│   │       ├── +layout.svelte
│   │       └── +page.svelte  # Factorio風ビューワー
│   ├── svelte.config.js      # adapter-static 設定
│   ├── vite.config.ts        # Proxy 設定（開発時）
│   └── package.json
├── main.go
├── go.mod
└── Makefile
```
