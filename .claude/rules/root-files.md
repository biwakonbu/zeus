---
description: ルートディレクトリのファイル編集時のガイドライン。ビルド設定・依存関係管理時に参照。
paths:
  - "main.go"
  - "Makefile"
  - "go.mod"
  - "go.sum"
---

# ルートファイル

## ファイル一覧

| ファイル | 役割 |
|----------|------|
| `main.go` | エントリーポイント（cmd.Execute() 呼び出しのみ） |
| `Makefile` | ビルド・テスト・開発タスクの自動化 |
| `go.mod` | Go モジュール定義・依存関係管理 |
| `go.sum` | 依存関係のチェックサム（自動生成） |

## main.go

- **最小限の実装を維持**: `cmd.Execute()` 呼び出しのみ
- ロジックは `cmd/root.go` または `internal/` に配置
- グローバル変数・初期化処理は避ける

## Makefile

主要ターゲット:

```makefile
make build          # Go バイナリビルド
make test           # 全テスト実行
make dashboard-dev  # フロントエンド開発サーバー
make build-all      # Go + SvelteKit 統合ビルド
```

**編集時の注意:**
- 新コマンド追加時は対応するターゲットを追加
- `dashboard-deps` は `zeus-dashboard/node_modules` 存在チェック付き

## go.mod / go.sum

- **直接編集禁止**: `go get`, `go mod tidy` で管理
- 依存追加時は最小限のパッケージのみ
- `go.sum` は自動生成のためコミット前に `go mod tidy` を実行
