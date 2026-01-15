# Zeus Dashboard - 設計書

## アーキテクチャ

### パッケージ構造

```
internal/dashboard/
├── server.go         # Server 構造体、Start/Shutdown
├── handlers.go       # API ハンドラー群
├── static/
│   ├── index.html    # ダッシュボード HTML
│   ├── styles.css    # スタイルシート
│   └── app.js        # フロントエンドロジック
└── dashboard_test.go # テスト
```

### コンポーネント構成

```
cmd/dashboard.go (Cobra コマンド)
         │
         ▼
internal/dashboard/Server
  - zeus: *core.Zeus
  - server: *http.Server
  + Start(ctx) error
  + Shutdown(ctx) error
         │
         ▼
internal/dashboard/handlers
  - handleIndex()
  - handleAPIStatus()
  - handleAPITasks()
  - handleAPIGraph()
  - handleAPIPredict()
         │
         ▼
internal/core/Zeus
  - Status()
  - List()
  - BuildDependencyGraph()
  - Predict()
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

## 静的ファイル設計

### CSS カラーテーマ
| ステータス | 色コード |
|-----------|---------|
| Good/Completed | #22c55e |
| Fair/InProgress | #f59e0b |
| Poor/Blocked | #ef4444 |
| Pending | #6b7280 |

### JavaScript 構成
- 5秒間隔ポーリング
- Mermaid.js 動的レンダリング
- エラーハンドリング付き fetch
- DOM 更新関数群

## エラーハンドリング

### サーバーサイド
- ErrorResponse 構造体で統一
- 適切な HTTP ステータスコード

### クライアントサイド
- エラーバナー表示
- 接続状態インジケータ
- コンソールログ出力

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
| TestHandle404 | 404 処理 |
