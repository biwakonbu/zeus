---
description: テスト戦略と E2E テストの詳細。テスト関連作業時に参照。
paths:
  - "tests/**"
  - "scripts/e2e/**"
  - "**/*_test.go"
---

# テスト

## 基本コマンド

```bash
go test ./...                    # 全テスト
go test -v ./internal/core/...   # 詳細出力
go test -cover ./...             # カバレッジ
```

## E2E テスト

### CLI テスト（Go）

```bash
go test -v ./tests/e2e/...       # E2E テスト実行
```

E2E テストは実バイナリをビルドして実行するため、事前の `go build` が必要。

### Web テスト（agent-browser）

```bash
./scripts/e2e/run-web-test.sh     # Web E2E テスト実行
./scripts/e2e/update-golden.sh    # ゴールデンファイル更新
```

**特性:**
- State-First アプローチ: `window.__ZEUS__` API で内部状態を直接検証
- 座標除外: x, y, id, viewport を比較から除外（安定性重視）
- agent-browser 統合: ヘッドレスブラウザで自動化
- jq 構造比較: JSON フィルタリングで正規化→ハッシュ比較
- エラーハンドリング強化:
  * agent-browser レスポンス検証 (JSON 形式チェック、成功フィールド確認)
  * jq フィルタ null 値チェック (ID→名前変換失敗検出)
  * window.__ZEUS__ API 存在確認 (型チェック + 関数検証)

**ファイル構成:**
- `scripts/e2e/run-web-test.sh` - メインテストスクリプト
- `scripts/e2e/run-parallel-tests.sh` - 並列テスト実行（3ジョブ同時）
- `scripts/e2e/update-golden.sh` - ゴールデン更新
- `scripts/e2e/lib/common.sh` - ユーティリティ・検証関数
- `scripts/e2e/lib/verify.sh` - jq 構造比較ロジック
- `scripts/e2e/lib/report.sh` - レポート生成（MD/HTML/Text）
- `scripts/e2e/golden/` - ゴールデンファイル格納

## ゴールデンテスト

ゴールデンファイルは `.claude/skills/zeus-e2e-tester/resources/golden/` に配置:
- `cli-init.golden.json` - zeus init 出力検証
- `cli-graph.golden.json` - zeus graph 出力検証
- `graph-state.golden.json` - Web グラフ状態検証
- `integration-graph-state.golden.json` - 統合テスト検証

詳細は `golden/README.md` を参照。
