---
description: ダッシュボード（SvelteKit + PixiJS）の設計詳細。フロントエンド作業時に参照。
paths:
  - "zeus-dashboard/**"
---

# ダッシュボード

## バックエンドモジュール (internal/dashboard/)

| モジュール | 責務 |
|-----------|------|
| Server | HTTP サーバー管理、静的ファイル配信、CORS ミドルウェア |
| Handlers | REST API ハンドラー |
| SSEBroadcaster | SSE クライアント管理、リアルタイムイベント配信 |

**バックエンド設計ポイント:**
- Go 標準ライブラリのみ使用（net/http, embed）
- SvelteKit ビルド成果物は `//go:embed build/*` で埋め込み
- SSE で リアルタイム更新（ポーリングへのフォールバックあり）
- 開発モード (`--dev`) で CORS 有効化
- 127.0.0.1 にバインドしてローカルアクセスのみ許可
- SPA フォールバックルーティング対応

## フロントエンド (zeus-dashboard/)

**フロントエンド設計ポイント:**
- SvelteKit + TypeScript で型安全な開発
- Factorio 風インダストリアル UI テーマ
- PixiJS (WebGL) で高パフォーマンスなタスクグラフ描画
- Svelte Stores でリアクティブな状態管理
- SSE クライアントで自動再接続ロジック実装

## 統合戦略

| 環境 | 実行方法 | 静的ファイル配信 | API アクセス |
|------|----------|------------------|--------------|
| **開発時** | `make dashboard-dev` + `go run . dashboard --dev` | Vite Dev Server (:5173) | Go Server (:8080) + CORS |
| **本番時** | `make build-all` → `zeus dashboard` | Go embed (:8080) | 同一オリジン |

## API エンドポイント

- `GET /api/status` - プロジェクト状態
- `GET /api/tasks` - タスク一覧
- `GET /api/graph` - 依存関係グラフ（Mermaid形式）
- `GET /api/predict` - 予測分析結果
- `GET /api/wbs` - WBS 階層構造
- `GET /api/timeline` - タイムラインとクリティカルパス
- `GET /api/downstream?task_id=X` - 下流・上流タスク取得
- `GET /api/events` - SSE ストリーム（リアルタイム更新）

## ダッシュボード機能

| 機能 | 説明 |
|------|------|
| タスクグラフ | PixiJS によるインタラクティブな依存関係グラフ表示 |
| ミニマップ | 全体像の把握と素早いナビゲーション |
| フィルター | ステータス・優先度・担当者でフィルタリング |
| タスク詳細 | タスク選択時に詳細パネル表示 |
| リアルタイム更新 | SSE + ポーリングフォールバック |

## コマンドオプション

- `--port` - ポート番号（デフォルト: 8080）
- `--no-open` - ブラウザを自動で開かない
- `--dev` - 開発モード（CORS 有効）

## ビュー切り替え

- Graph View: 依存関係グラフ（Factorio 風）
- WBS View: 階層構造ツリー
- Timeline View: ガントチャート風表示

## 影響範囲可視化

- 選択タスクの下流タスクを黄色でハイライト
- 上流タスクを青色でハイライト
- 選択タスクはオレンジ色で強調
