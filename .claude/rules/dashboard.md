---
description: ダッシュボード（SvelteKit + PixiJS）の設計詳細。フロントエンド作業時に参照。
paths:
  - "zeus-dashboard/**"
---

# ダッシュボード

## バックエンド（internal/dashboard）

- `server.go`: ルーティング、静的配信、CORS、起動制御
- `handlers_core.go`: `/api/status`, `/api/graph`, `/api/activities`
- `handlers_uml.go`: `/api/actors`, `/api/usecases`, `/api/subsystems`, `/api/uml/*`
- `handlers_unified.go`: `/api/unified-graph`
- `handlers_affinity.go`: `/api/affinity`
- `sse.go`: `/api/events` 向け SSE ブロードキャスト

## フロントエンド（zeus-dashboard）

- SvelteKit + TypeScript + PixiJS
- レイアウト/描画の中核: `src/lib/viewer/FactorioViewer.svelte`
- レイアウトエンジン: `src/lib/viewer/engine/LayoutEngine.ts`
- API 型: `src/lib/types/api.ts`

## ビュー種別（現行実装）

- `graph`
- `usecase`
- `activity`

根拠: `zeus-dashboard/src/lib/viewer/ui/types.ts`

## API エンドポイント

- `GET /api/status`
- `GET /api/graph`
- `GET /api/affinity`
- `GET /api/actors`
- `GET /api/usecases`
- `GET /api/subsystems`
- `GET /api/uml/usecase`
- `GET /api/activities`
- `GET /api/uml/activity`
- `GET /api/unified-graph`
- `GET /api/events`

## 開発コマンド

```bash
make dashboard-deps
make dashboard-dev
go run . dashboard --dev
```
