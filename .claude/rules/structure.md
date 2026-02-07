---
description: Zeus プロジェクトのディレクトリ構造。必要時に手動参照。
---

# プロジェクト構造

```text
zeus/
├── cmd/                      # Cobra コマンド
├── internal/
│   ├── core/                 # ドメイン・状態・承認・各種ハンドラー
│   ├── analysis/             # graph/unified/affinity 分析
│   ├── report/               # レポート生成
│   ├── dashboard/            # HTTP サーバー/REST/SSE
│   │   ├── server.go
│   │   ├── handlers_core.go
│   │   ├── handlers_uml.go
│   │   ├── handlers_unified.go
│   │   ├── handlers_affinity.go
│   │   └── sse.go
│   ├── yaml/                 # YAML I/O とファイルロック
│   ├── doctor/               # 診断/修復
│   ├── generator/            # .claude assets 生成
│   └── testing/              # テスト補助
├── zeus-dashboard/           # SvelteKit + PixiJS
├── docs/                     # 文書（正本/仕様/設計/履歴）
├── .claude/                  # agents / skills / rules
├── CLAUDE.md
├── go.mod
└── Makefile
```

## 重要ポイント

- dashboard handler は機能別 `handlers_*.go` を編集する（単一ハンドラーファイルを前提にしない）。
- `docs/archive/` は履歴凍結領域。現行仕様は `docs/README.md` から正本へ辿る。
- `.claude/agents/*.md` は `internal/generator/assets/agents/*.md` の生成物。直接編集よりテンプレート編集を優先する。
