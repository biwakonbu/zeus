# Zeus E2E テストスイート

Zeus プロジェクトの E2E テストフレームワーク。`agent-browser` CLI を使用した Web UI の状態ベース検証と、複数形式のレポート生成機能を提供します。

## Quick Start

### 前提条件

```bash
# Zeus ビルド
cd /path/to/zeus
make build

# ダッシュボード依存関係インストール
cd zeus-dashboard
npm ci
npm run build
cd ..

# 必須ツール
npm install -g agent-browser
agent-browser install
brew install jq  # または apt install jq
```

### 単一テスト実行

```bash
./scripts/e2e/run-web-test.sh
```

**期待される出力:**
```
==> Zeus E2E テスト開始
...
==> テスト統計
[INFO] 実行時間: 2秒
[INFO] 成功ステップ: 9/9
[INFO] 成功率: 100%

[PASS] ============================================
[PASS] Zeus E2E テスト: 全て成功
[PASS] ============================================
```

### アーティファクト付きテスト実行

```bash
export KEEP_ARTIFACTS=true
./scripts/e2e/run-web-test.sh

# アーティファクトを確認
ls -lh /tmp/zeus-e2e-artifacts/
cat /tmp/zeus-e2e-artifacts/actual-state.json | jq .
```

**生成されるファイル:**
- `actual-state.json` - キャプチャした実際の状態
- `test-report.json` - JSON 形式のテスト結果
- `report.md` - Markdown レポート
- `report.html` - HTML レポート（ブラウザで表示可能）
- `report.txt` - テキストレポート
- `server.log` - ダッシュボードサーバーログ
- `screenshot.png` - UI スクリーンショット
- `zeus-data.tar.gz` - プロジェクトデータ

## ファイル構成

| ファイル | 説明 |
|---------|------|
| `run-web-test.sh` | メインテストスクリプト |
| `run-parallel-tests.sh` | 複数シナリオ並列実行 |
| `update-golden.sh` | ゴールデンファイル更新 |
| `lib/common.sh` | ログ、設定、ユーティリティ |
| `lib/verify.sh` | jq を使った構造比較 |
| `lib/report.sh` | 複数形式レポート生成 |
| `golden/` | ゴールデンファイル格納 |

## 機能

### 1. 状態ベース検証（座標除外）

**従来のピクセル比較ではなく、JSON 状態を検証。**

- ✅ **除外フィールド** - x, y, id, viewport（環境依存的な要素）
- ✅ **検証対象** - タスク名、ステータス、進捗度、依存関係
- ✅ **クロスプラットフォーム** - OS/ブラウザ依存性なし
- ✅ **Git フレンドリー** - 全ゴールデンファイルがテキスト形式

```bash
# ゴールデン比較の仕組み
jq -S '.nodes[] | {name, status, progress}' actual.json
# ↓ 名前順ソート + 座標除外 + 依存関係検証 → パス/フェイル
```

### 2. マルチシナリオ並列実行

複数の異なるプロジェクト構成でテストを並列実行：

```bash
./scripts/e2e/run-parallel-tests.sh
```

**特性:**
- **最大3ジョブ**を同時実行
- 各シナリオは**動的ポート割り当て**（衝突回避）
- シナリオ: `basic`, `complex`, `large`
- 統合レポートで全結果を集約

### 3. 複数形式レポート

テスト完了後、自動的に 3 形式のレポートを生成。

### 4. カスタムタイムアウト設定

環境変数で動的にタイムアウト値を変更：

```bash
TIMEOUT_APP_READY=40 ./scripts/e2e/run-web-test.sh
```

## 設定可能な環境変数

| 変数 | デフォルト | 説明 |
|------|-----------|------|
| `TIMEOUT_SERVER_START` | 30秒 | サーバー起動待機 |
| `TIMEOUT_API_READY` | 10秒 | API Ready 待機 |
| `TIMEOUT_APP_READY` | 20秒 | アプリケーション Ready 待機 |
| `TIMEOUT_CAPTURE` | 5秒 | 状態キャプチャタイムアウト |
| `DASHBOARD_PORT` | 18080 | ダッシュボードポート |
| `KEEP_ARTIFACTS` | false | アーティファクト保持 |
| `ARTIFACTS_DIR` | `/tmp/zeus-e2e-artifacts` | アーティファクト保存先 |

## ゴールデンファイル更新

```bash
./scripts/e2e/update-golden.sh

# 差分確認
git diff scripts/e2e/golden/

# コミット
git add scripts/e2e/golden/
git commit -m 'chore: update E2E golden files'
```

## トラブルシューティング

### ポート既に使用
```bash
lsof -i :18080
kill -9 <PID>
./scripts/e2e/run-web-test.sh
```

### タイムアウト
```bash
TIMEOUT_SERVER_START=60 TIMEOUT_APP_READY=40 ./scripts/e2e/run-web-test.sh
```

### テスト失敗時に詳細確認
```bash
cat /tmp/zeus-e2e-artifacts/server.log
open /tmp/zeus-e2e-artifacts/screenshot.png
cat /tmp/zeus-e2e-artifacts/actual-state.json | jq .
```

## CI/CD 統合

GitHub Actions で自動実行可能。詳細は `.github/workflows/e2e.yml` を参照。

