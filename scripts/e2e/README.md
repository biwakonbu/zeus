# Zeus E2E テストスイート

agent-browser を使用した Web ダッシュボードの E2E テストスイートです。
State-First アプローチを採用し、PixiJS Canvas の内部状態を `window.__ZEUS__` API 経由で検証します。

## Quick Start

### 前提条件

```bash
# 1. Zeus バイナリをビルド
make build

# 2. ダッシュボードをビルド
cd zeus-dashboard && npm ci && npm run build && cd ..

# 3. agent-browser をインストール
npm install -g agent-browser

# 4. jq をインストール（macOS）
brew install jq
# または Ubuntu
# apt install jq
```

### テスト実行

```bash
# E2E テスト実行
./scripts/e2e/run-web-test.sh
```

### ゴールデンファイル更新

```bash
# 更新スクリプト実行
./scripts/e2e/update-golden.sh

# 差分確認
git diff scripts/e2e/golden/state/basic-tasks.json

# 意図的な変更の場合のみコミット
git add scripts/e2e/golden/
git commit -m 'chore: update E2E golden files'
```

## ディレクトリ構造

```
scripts/e2e/
├── run-web-test.sh           # メインテストスクリプト
├── update-golden.sh          # ゴールデン更新ユーティリティ
├── lib/
│   ├── common.sh             # ログ、設定、ユーティリティ
│   └── verify.sh             # jq を使った構造比較
├── golden/
│   ├── state/
│   │   └── basic-tasks.json  # 構造ゴールデン（座標なし）
│   └── performance/
│       └── README.md         # 将来用プレースホルダー
└── README.md                 # このファイル
```

## 検証方式

### State-First アプローチ

DOM ベースの検証ではなく、アプリケーション内部状態を直接検証します:

1. `window.__ZEUS__.isReady()` で描画完了を待機
2. `window.__ZEUS__.getGraphState()` で状態取得
3. ゴールデンファイルと構造比較

### 座標除外版の構造比較

以下のフィールドは比較から除外されます:

- `nodes[*].x`, `nodes[*].y` - 座標（レイアウトアルゴリズム依存）
- `nodes[*].id` - UUID（動的生成）
- `viewport` - ビューポート状態

比較対象:
- `nodes[*].name` - タスク名
- `nodes[*].status` - ステータス
- `nodes[*].progress` - 進捗率
- `edges` - 名前ベースの依存関係

### エッジの名前ベース変換

ゴールデンファイルではエッジを名前で定義:

```json
{
  "edges": [
    { "from": "Task A", "to": "Task C" }
  ]
}
```

実際の状態（ID ベース）から名前ベースに変換して比較します。

## タイムアウト設定

| 設定 | デフォルト | 環境変数 |
|------|-----------|----------|
| サーバー起動 | 30秒 | `TIMEOUT_SERVER_START` |
| API Ready | 10秒 | `TIMEOUT_API_READY` |
| アプリ Ready | 20秒 | `TIMEOUT_APP_READY` |
| 状態キャプチャ | 5秒 | `TIMEOUT_CAPTURE` |

## 環境変数

| 変数 | 説明 | デフォルト |
|------|------|-----------|
| `DASHBOARD_PORT` | ダッシュボードポート | `18080` |
| `KEEP_ARTIFACTS` | アーティファクト保持 | `false` |
| `ARTIFACTS_DIR` | アーティファクト保存先 | `/tmp/zeus-e2e-artifacts` |

## エラー時のアーティファクト

テスト失敗時に以下が自動収集されます:

- `actual-state.json` - 取得したグラフ状態
- `server.log` - ダッシュボードサーバーログ
- `zeus-data.tar.gz` - .zeus ディレクトリのアーカイブ
- `screenshot.png` - ブラウザスクリーンショット
- `metrics.json` - パフォーマンスメトリクス

保存先: `/tmp/zeus-e2e-artifacts/`

## トラブルシューティング

### agent-browser が見つからない

```bash
npm install -g agent-browser
```

### ブラウザが起動しない

Playwright の依存関係をインストール:

```bash
npx playwright install chromium --with-deps
```

### ポートが使用中

別のポートを指定:

```bash
DASHBOARD_PORT=18081 ./scripts/e2e/run-web-test.sh
```

### タイムアウト

タイムアウトを延長:

```bash
TIMEOUT_APP_READY=60 ./scripts/e2e/run-web-test.sh
```

## CI/CD

GitHub Actions での実行は `.github/workflows/e2e.yml` を参照してください。

ヘッドレスモードで実行されます。
