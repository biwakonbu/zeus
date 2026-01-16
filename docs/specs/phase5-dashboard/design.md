# Zeus Dashboard - 設計書

## 概要

Zeus ダッシュボードは **SvelteKit** で実装された Web UI。**Factorio 風インダストリアルデザイン** と **SSE リアルタイム更新**を採用。

## アーキテクチャ

### パッケージ構造

```
zeus/
├── internal/dashboard/           # Go バックエンド
│   ├── server.go                 # HTTP サーバー + SSE + CORS
│   ├── handlers.go               # API ハンドラー + SSE エンドポイント
│   ├── sse.go                    # SSE Broadcaster
│   ├── build/                    # SvelteKit ビルド成果物（embed）
│   └── dashboard_test.go
│
└── zeus-dashboard/               # SvelteKit フロントエンド
    ├── src/
    │   ├── lib/
    │   │   ├── api/              # API クライアント + SSE
    │   │   │   ├── client.ts     # Fetch ベースの API クライアント
    │   │   │   └── sse.ts        # SSE クライアント（自動再接続）
    │   │   ├── stores/           # Svelte ストア
    │   │   │   ├── connection.ts # 接続状態
    │   │   │   ├── status.ts     # プロジェクト状態
    │   │   │   ├── tasks.ts      # タスク
    │   │   │   ├── graph.ts      # グラフ
    │   │   │   └── prediction.ts # 予測
    │   │   ├── components/
    │   │   │   ├── layout/       # Header, Footer
    │   │   │   ├── panels/       # Overview, Stats, Tasks, Graph, Prediction
    │   │   │   ├── ui/           # Badge, ProgressBar, Table, Stat, Panel
    │   │   │   └── graph/        # MermaidGraph
    │   │   ├── theme/            # Factorio デザインシステム
    │   │   │   ├── variables.css # CSS 変数
    │   │   │   └── factorio.css  # グローバルスタイル
    │   │   └── types/            # TypeScript 型定義
    │   │       └── api.ts
    │   └── routes/
    │       ├── +layout.svelte
    │       └── +page.svelte
    ├── svelte.config.js          # adapter-static
    ├── vite.config.ts            # Proxy 設定
    └── package.json
```

### コンポーネント構成

```
cmd/dashboard.go (Cobra コマンド)
         │
         ▼
internal/dashboard/Server
  - zeus: *core.Zeus
  - server: *http.Server
  - devMode: bool
  - broadcaster: *SSEBroadcaster
  + Start(ctx) error
  + Shutdown(ctx) error
  + BroadcastAllUpdates()
         │
         ├─────────────────────────────────┐
         ▼                                 ▼
internal/dashboard/handlers         internal/dashboard/SSEBroadcaster
  - handleIndex()                     - clients: map[string]*SSEClient
  - handleAPIStatus()                 + AddClient(id, w, r)
  - handleAPITasks()                  + RemoveClient(id)
  - handleAPIGraph()                  + Broadcast(event SSEEvent)
  - handleAPIPredict()                + BroadcastStatus/Task/Graph/Prediction
  - handleSSE()
         │
         ▼
internal/core/Zeus
  - Status()
  - List()
  - BuildDependencyGraph()
  - Predict()
```

### フロントエンドアーキテクチャ

```
+page.svelte (メインページ)
         │
         ├─ onMount: refreshAllData() + connectSSE()
         │
         ▼
┌─────────────────────────────────────────────────┐
│ Dashboard Layout (Grid)                         │
│  ┌─────────────┬─────────────┐                 │
│  │ OverviewPanel│ StatsPanel │                 │
│  └─────────────┴─────────────┘                 │
│  ┌───────────────────────────┐                 │
│  │       TasksPanel          │                 │
│  └───────────────────────────┘                 │
│  ┌─────────────┬─────────────┐                 │
│  │ GraphPanel  │PredictionPanel│               │
│  └─────────────┴─────────────┘                 │
└─────────────────────────────────────────────────┘
         │
         ▼
Svelte Stores (リアクティブ状態管理)
  - statusStore → OverviewPanel, StatsPanel
  - tasksStore → TasksPanel
  - graphStore → GraphPanel
  - predictionStore → PredictionPanel
  - connectionStore → Header (接続状態表示)
```

## API レスポンス形式

### GET /api/status
```json
{
  "project": {
    "id": "zeus-xxx",
    "name": "Project Name",
    "description": "Description",
    "start_date": "2026-01-01"
  },
  "state": {
    "health": "good",
    "summary": {
      "total_tasks": 10,
      "completed": 3,
      "in_progress": 2,
      "pending": 5
    }
  },
  "pending_approvals": 0
}
```

### GET /api/tasks
```json
{
  "tasks": [
    {
      "id": "task-xxx",
      "title": "Task Title",
      "status": "in_progress",
      "priority": "high",
      "assignee": "ai",
      "dependencies": ["task-yyy"]
    }
  ],
  "total": 10
}
```

### GET /api/graph
```json
{
  "mermaid": "graph TD\n  task_xxx --> task_yyy",
  "stats": {
    "total_nodes": 10,
    "with_dependencies": 5,
    "isolated_count": 3,
    "cycle_count": 0,
    "max_depth": 3
  },
  "cycles": [],
  "isolated": ["task-zzz"]
}
```

### GET /api/predict
```json
{
  "completion": {
    "remaining_tasks": 7,
    "average_velocity": 2.5,
    "estimated_date": "2026-02-15",
    "confidence_level": 70,
    "margin_days": 5
  },
  "risk": {
    "overall_level": "Medium",
    "factors": [],
    "score": 40
  },
  "velocity": {
    "last_7_days": 2,
    "last_14_days": 5,
    "last_30_days": 10,
    "weekly_average": 2.5,
    "trend": "Stable"
  }
}
```

### GET /api/events (SSE)

**Content-Type:** `text/event-stream`

**イベント形式:**
```
event: connected
data: {"message": "SSE connection established"}

event: status
data: {"project": {...}, "state": {...}}

event: task
data: {"tasks": [...], "total": 10}

event: graph
data: {"mermaid": "...", "stats": {...}}

event: prediction
data: {"completion": {...}, "risk": {...}, "velocity": {...}}
```

## Factorio 風デザインシステム

### CSS 変数 (variables.css)

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

  /* テキスト */
  --text-primary: #ffffff;
  --text-secondary: #b8b8b8;

  /* ステータス色 */
  --status-good: #44cc44;
  --status-fair: #ffcc00;
  --status-poor: #ee4444;

  /* フォント */
  --font-family: 'IBM Plex Mono', monospace;
}
```

### コンポーネントスタイル

| コンポーネント | スタイル |
|---------------|----------|
| Panel | 金属フレーム効果、inset シャドウ |
| ProgressBar | オレンジ充填、インダストリアル感 |
| Badge | 角丸控えめ、ステータス別背景色 |
| Table | ホバー時に背景色変化、金属色ボーダー |
| Button | グラデーション、ホバー時にハイライト |

## SSE 実装

### Go 側 (SSEBroadcaster)

```go
type SSEBroadcaster struct {
    clients map[string]*SSEClient
    mu      sync.RWMutex
}

func (b *SSEBroadcaster) Broadcast(event SSEEvent) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    for _, client := range b.clients {
        client.Send(event)
    }
}
```

### フロントエンド側 (SSEClient)

```typescript
export class SSEClient {
    private eventSource: EventSource | null = null;
    private reconnectDelay: number = 3000;
    private maxReconnectAttempts: number = 10;

    connect(): void {
        this.eventSource = new EventSource('/api/events');

        ['status', 'task', 'graph', 'prediction'].forEach(type => {
            this.eventSource.addEventListener(type, (e) => {
                this.dispatchEvent(type, JSON.parse(e.data));
            });
        });

        this.eventSource.onerror = () => this.handleError();
    }

    private handleError(): void {
        // 自動再接続（exponential backoff）
    }
}
```

## 統合戦略

| 環境 | 実行方法 | 静的ファイル | API |
|------|----------|--------------|-----|
| **開発** | `make dashboard-dev` + `go run . dashboard --dev` | Vite :5173 (HMR) | Go :8080 (CORS) |
| **本番** | `make build-all` → `zeus dashboard` | Go embed :8080 | 同一オリジン |

### 開発ワークフロー

```bash
# ターミナル 1: Go サーバー（CORS 有効）
go run . dashboard --dev --port 8080

# ターミナル 2: Vite 開発サーバー（HMR）
cd zeus-dashboard && npm run dev
```

ブラウザで `http://localhost:5173` にアクセス。

### 本番ビルド

```bash
make build-all    # SvelteKit ビルド + Go ビルド
./zeus dashboard  # 127.0.0.1:8080 でサーバー起動
```

## テスト

| テスト | 内容 |
|--------|------|
| TestNewServer | サーバー初期化 |
| TestServerStartShutdown | 起動・停止サイクル |
| TestHandleAPIStatus | ステータス API |
| TestHandleAPITasks | タスク API |
| TestHandleAPIGraph | グラフ API |
| TestHandleAPIPredict | 予測 API |
| TestHandleIndex | インデックスページ |
| TestHandleMethodNotAllowed | メソッド不許可 |
| TestHandleSPAFallback | SPA フォールバック（200 OK + index.html） |

## エラーハンドリング

### サーバーサイド
- ErrorResponse 構造体で統一
- 適切な HTTP ステータスコード
- CORS ヘッダー（開発モード時）

### クライアントサイド
- 接続状態インジケータ（Header）
- SSE 自動再接続（最大 10 回）
- ポーリングへのフォールバック
- エラー時の Store 状態更新
