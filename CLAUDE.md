# CLAUDE.md

Zeus は「神の視点」でプロジェクト管理を支援する AI 駆動型 CLI システム。Go + Cobra で実装。

## 設計哲学

- ファイルベース: 外部 DB 不要、YAML で人間可読
- 人間中心: AI は提案者、人間が最終決定者
- シンプルな初期化: 単一の `zeus init` コマンドで全機能を利用可能
- Git 親和性: 全データがテキストで差分追跡可能

## 技術スタック

**バックエンド:** Go 1.21+, Cobra, gopkg.in/yaml.v3, fatih/color, github.com/google/uuid

**フロントエンド:** SvelteKit + TypeScript, PixiJS (WebGL), SSE, Factorio 風デザイン

## コーディング規約

- **Go**: 標準規約（gofmt, go vet）に準拠
- **TypeScript/Svelte**: ESLint + Prettier
- **コメント**: 日本語
- **変数・関数名**: 英語（国際標準）

## 開発コマンド

```bash
make build              # ビルド
make test               # テスト実行
go test -v ./internal/core/...  # 単一パッケージテスト
go run . <command>      # 開発実行

# ダッシュボード開発
make dashboard-deps     # npm 依存関係インストール（初回のみ）
make dashboard-dev      # Vite 開発サーバー起動
go run . dashboard --dev  # Go サーバー起動（CORS 有効）
make build-all          # 統合ビルド
```

## 実装フェーズ

| Phase | 内容 | 状態 |
|-------|------|------|
| Phase 1 (MVP) | init, status, add, list, doctor, fix | 完了 |
| Phase 2 (Standard) | pending, approve, reject, snapshot, history | 完了 |
| Phase 2.5-2.7 | セキュリティ、DI/Context、suggest/apply | 完了 |
| Phase 3 (AI統合) | Claude Code 連携、explain | 完了 |
| Phase 4 (分析) | graph, predict, report | 完了 |
| Phase 5 (ダッシュボード) | Factorio風ビューワー、SSE | 完了 |
| Phase 6 (WBS・タイムライン) | WBS階層、クリティカルパス、影響範囲可視化 | 完了 |

## 実装済みコマンド

```bash
# コア操作
zeus init                                       # プロジェクト初期化
zeus status                                     # 状態表示
zeus add <entity> <name> [options]              # エンティティ追加
  # --parent <id>  --start <date>  --due <date>  --progress <0-100>  --wbs <code>
zeus list [entity]                              # 一覧表示
zeus doctor                                     # 診断
zeus fix [--dry-run]                            # 修復

# 承認管理
zeus pending                                    # 承認待ち一覧
zeus approve <id>                               # 承認
zeus reject <id> [--reason ""]                  # 却下
zeus snapshot create|list|restore              # スナップショット管理
zeus history [-n limit]                         # 履歴表示

# AI 機能
zeus suggest [--limit N] [--impact level]       # 提案生成
zeus apply <suggestion-id>                      # 提案適用
zeus explain <entity-id> [--context]            # 詳細説明

# 分析・可視化
zeus graph [--format text|dot|mermaid] [-o file]    # 依存関係グラフ
zeus predict [completion|risk|velocity|all]         # 予測分析
zeus report [--format text|html|markdown] [-o file] # レポート生成
zeus dashboard [--port 8080] [--no-open] [--dev]    # Web ダッシュボード

# ユーティリティ
zeus update-claude                              # Claude Code ファイル再生成
```

## 承認レベル

| レベル | 説明 | デフォルト |
|--------|------|-----------|
| auto | 自動承認（即時実行） | ✓ |
| notify | 通知のみ（ログ記録して実行） | |
| approve | 明示的承認必要 | |

`zeus.yaml` の `automation_level` で変更可能。

## Claude Code 連携

`zeus init` で `.claude/` ディレクトリに連携ファイルを生成。
既存プロジェクトの更新: `zeus update-claude`

**生成ファイル:**
- `agents/zeus-orchestrator.md` - 全コマンド一覧
- `agents/zeus-planner.md` - WBS・タイムライン設計
- `agents/zeus-reviewer.md` - 分析・レビュー
- `skills/zeus-project-scan/SKILL.md` - プロジェクト状態取得
- `skills/zeus-task-suggest/SKILL.md` - タスク提案
- `skills/zeus-risk-analysis/SKILL.md` - リスク分析

## ドキュメント

- `docs/SYSTEM_DESIGN.md` - システム設計書（必読）
- `docs/IMPLEMENTATION_GUIDE.md` - Go 実装ガイド
- `docs/OPERATIONS_MANUAL.md` - 運用マニュアル

## 詳細情報

詳細なアーキテクチャ、プロジェクト構造、ダッシュボード設計は `.claude/rules/` を参照:
- `architecture.md` - コアモジュール、DI パターン、セキュリティ対策
- `dashboard.md` - フロントエンド/バックエンド設計、API エンドポイント、メトリクス計測
- `structure.md` - ディレクトリ構造の詳細
- `testing.md` - E2E テスト、ゴールデンテストの詳細

## テスト

```bash
go test ./...                    # 全テスト
go test -v ./internal/core/...   # 詳細出力
go test -cover ./...             # カバレッジ
go test -v ./tests/e2e/...       # E2E テスト
```

E2E テスト詳細は `.claude/rules/testing.md` を参照。
