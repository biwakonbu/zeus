# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Zeus は「神の視点」でプロジェクト管理を支援する AI 駆動型 CLI システム。Go + Cobra で実装。

**設計哲学:**
- ファイルベース: 外部 DB 不要、YAML で人間可読
- 人間中心: AI は提案者、人間が最終決定者
- 段階的複雑化: Simple → Standard → Advanced の3レベル
- Git 親和性: 全データがテキストで差分追跡可能

## 技術スタック

- **言語**: Go 1.21+
- **CLI フレームワーク**: Cobra
- **YAML 処理**: gopkg.in/yaml.v3
- **カラー出力**: fatih/color
- **UUID 生成**: github.com/google/uuid
- **配布形式**: スタンドアロン CLI + Claude Code Plugin 連携

## 開発コマンド

```bash
# ビルド
make build

# テスト実行
make test

# 単一パッケージテスト
go test -v ./internal/core/...

# 開発実行
go run . <command>
make dev ARGS="init --level=simple"

# インストール
make install
```

## プロジェクト構造

```
zeus/
├── cmd/                      # Cobra コマンド
│   ├── root.go               # ルートコマンド（DI対応）
│   ├── init.go               # zeus init
│   ├── status.go             # zeus status
│   ├── pending.go            # zeus pending
│   ├── approve.go            # zeus approve
│   ├── reject.go             # zeus reject
│   ├── snapshot.go           # zeus snapshot
│   ├── history.go            # zeus history
│   ├── graph.go              # zeus graph（依存関係グラフ）
│   ├── predict.go            # zeus predict（予測分析）
│   ├── report.go             # zeus report（レポート生成）
│   ├── dashboard.go          # zeus dashboard（Webダッシュボード）
│   └── ...
├── internal/                 # 内部パッケージ
│   ├── core/                 # コアロジック
│   │   ├── zeus.go           # メインロジック（DI対応、分析機能統合）
│   │   ├── interfaces.go     # FileStore, StateStore, ApprovalStore インターフェース
│   │   ├── entity.go         # EntityHandler, EntityRegistry
│   │   ├── task_handler.go   # タスクエンティティハンドラー
│   │   ├── types.go          # 型定義
│   │   ├── state.go          # 状態・スナップショット管理（Context対応）
│   │   ├── approval.go       # 3段階承認システム（Context対応）
│   │   ├── errors.go         # エラー定義
│   │   └── mocks/            # テスト用モック
│   │       └── mock_file_store.go
│   ├── analysis/             # 分析機能（Phase 4）
│   │   ├── types.go          # 分析用型定義（独立）
│   │   ├── graph.go          # 依存関係グラフ構築・可視化
│   │   └── predict.go        # 予測分析（完了日、リスク、ベロシティ）
│   ├── report/               # レポート生成（Phase 4）
│   │   ├── generator.go      # レポート生成ロジック
│   │   └── templates.go      # 出力テンプレート（TEXT/HTML/Markdown）
│   ├── dashboard/            # Web ダッシュボード（Phase 5）
│   │   ├── server.go         # HTTP サーバー
│   │   ├── handlers.go       # API ハンドラー
│   │   ├── static/           # 静的ファイル（embed）
│   │   │   ├── index.html
│   │   │   ├── styles.css
│   │   │   └── app.js
│   │   └── dashboard_test.go
│   ├── yaml/                 # YAML 操作
│   │   ├── parser.go
│   │   ├── writer.go
│   │   ├── file_manager.go   # パス検証・セキュリティ（Context対応）
│   │   └── filelock.go       # ファイルロック機構
│   ├── doctor/               # 診断・修復（Context対応）
│   │   └── doctor.go
│   └── generator/            # Claude Code 連携ファイル生成（Context対応）
│       └── generator.go
├── main.go
├── go.mod
└── Makefile
```

## アーキテクチャ

### コアモジュール (internal/core/)

| モジュール | 責務 |
|-----------|------|
| Zeus | メインロジック、プロジェクト初期化、コマンド実行 |
| StateManager | 状態スナップショット管理、履歴追跡 |
| ApprovalManager | 3段階承認フロー (auto/notify/approve)、ファイルロック |
| TaskHandler | タスクエンティティの CRUD 操作 |
| EntityRegistry | エンティティハンドラーの登録・取得 |

### 分析モジュール (internal/analysis/)

| モジュール | 責務 |
|-----------|------|
| GraphBuilder | タスク依存関係グラフの構築 |
| DependencyGraph | グラフ構造、循環検出、統計計算、可視化出力 |
| Predictor | 完了日予測、リスク分析、ベロシティ計算 |

**設計ポイント:**
- `analysis` パッケージは `core` からの import cycle を避けるため独自の型を定義
- `core.Zeus` から `analysis` への変換関数で連携

### レポートモジュール (internal/report/)

| モジュール | 責務 |
|-----------|------|
| Generator | プロジェクトレポートの生成 |
| Templates | TEXT/HTML/Markdown テンプレート |

### ダッシュボードモジュール (internal/dashboard/)

| モジュール | 責務 |
|-----------|------|
| Server | HTTP サーバー管理、静的ファイル配信 |
| Handlers | REST API ハンドラー（/api/status, /api/tasks, /api/graph, /api/predict） |

**設計ポイント:**
- Go 標準ライブラリのみ使用（net/http, embed）
- 静的ファイルは `//go:embed` で埋め込み
- Mermaid.js は CDN から読み込み
- 127.0.0.1 にバインドしてローカルアクセスのみ許可

### DI パターン

Zeus は Option パターンによる依存性注入をサポート:

```go
// 本番環境
z := core.New(projectPath)

// テスト環境（モック注入）
z := core.New(projectPath,
    core.WithFileStore(mockFS),
    core.WithStateStore(mockSS),
    core.WithApprovalStore(mockAS),
)
```

### Context 対応

全ての公開 API が `context.Context` を第一引数として受け取る:
- タイムアウト制御
- キャンセル伝播
- 非同期処理のコントロール

### セキュリティ対策

| 対策 | 実装箇所 | 説明 |
|------|----------|------|
| ディレクトリトラバーサル防止 | file_manager.go | ValidatePath でパス検証 |
| ID 衝突防止 | zeus.go, approval.go | UUID v4 ベースの ID 生成 |
| 承認フロー原子性 | approval.go | flock ベースのファイルロック |
| ダッシュボードローカル専用 | server.go | 127.0.0.1 バインド |

### データ構造 (.zeus/)

zeus init 実行後、ターゲットプロジェクトに生成される構造:

```
.zeus/
├── zeus.yaml              # プロジェクト定義（メイン）
├── tasks/
│   ├── active.yaml        # 進行中タスク
│   └── backlog.yaml       # バックログ
├── state/
│   ├── current.yaml       # 現在の状態
│   └── snapshots/         # 履歴スナップショット
├── approvals/             # 承認管理 (standard/advanced)
│   ├── pending/           # 承認待ち
│   ├── approved/          # 承認済み
│   └── rejected/          # 却下済み
└── backups/               # 自動バックアップ
```

### Claude Code 連携

`zeus init --level=standard` または `--level=advanced` で初期化すると、Claude Code 連携用のファイルが自動生成されます。

**生成される構造:**
```
.claude/
├── agents/                # Zeus 用エージェント
│   ├── zeus-orchestrator.md
│   ├── zeus-planner.md
│   └── zeus-reviewer.md
└── skills/                # Zeus 用スキル
    ├── zeus-project-scan/SKILL.md
    ├── zeus-task-suggest/SKILL.md
    └── zeus-risk-analysis/SKILL.md
```

**設計方針:**
- Zeus CLI はスタンドアロンで動作（外部依存なし）
- Claude Code との連携は生成されたエージェント/スキルを通じて実行
- 提案機能はルールベース + AI ベースのハイブリッド対応

## 実装フェーズ

| Phase | 内容 | 状態 |
|-------|------|------|
| **Phase 1 (MVP)** | init, status, add, list, doctor, fix | 完了 |
| **Phase 2 (Standard)** | pending, approve, reject, snapshot, history | 完了 |
| **Phase 2.5 (Security)** | パス検証、UUID ID、ファイルロック | 完了 |
| **Phase 2.6 (DI/Context)** | DI対応、Context対応、テスト強化 | 完了 |
| **Phase 2.7 (Suggest)** | suggest, apply (ルールベース提案) | 完了 |
| **Phase 3 (AI統合)** | Claude Code 連携、explain、Add+承認フロー連携 | 完了 |
| **Phase 4 (高度な分析)** | graph, predict, report（依存関係グラフ、予測分析、レポート生成） | 完了 |
| **Phase 5 (ダッシュボード)** | Web UI、リアルタイム更新、Mermaid.js グラフ | 完了 |

## ドキュメント

- `docs/SYSTEM_DESIGN.md` - システム設計書（必読）
- `docs/IMPLEMENTATION_GUIDE.md` - Go 実装ガイド
- `docs/OPERATIONS_MANUAL.md` - 運用マニュアル

## 現在の状態

**Phase 5 (ダッシュボード) 完了** - 2026-01-15

### 実装済みコマンド

```bash
# Phase 1 (MVP)
zeus init [--level=simple|standard|advanced]   # プロジェクト初期化
zeus status [--detail]                          # 状態表示
zeus add <entity> <name>                        # エンティティ追加（承認フロー連携）
zeus list [entity]                              # 一覧表示
zeus doctor                                     # 診断
zeus fix [--dry-run]                            # 修復

# Phase 2 (Standard)
zeus pending                                    # 承認待ち一覧
zeus approve <id>                               # 承認
zeus reject <id> [--reason ""]                  # 却下
zeus snapshot create [label]                    # スナップショット作成
zeus snapshot list [-n limit]                   # スナップショット一覧
zeus snapshot restore <timestamp>               # スナップショットから復元
zeus history [-n limit]                         # プロジェクト履歴表示

# Phase 2.7 (Suggest)
zeus suggest [--limit N] [--impact high|medium|low]  # 提案生成
zeus apply <suggestion-id>                      # 提案適用
zeus apply --all [--dry-run]                    # 全提案適用

# Phase 3 (AI統合)
zeus explain <entity-id> [--context]            # エンティティの詳細説明

# Phase 4 (高度な分析)
zeus graph [--format text|dot|mermaid] [-o file]    # 依存関係グラフ表示
zeus predict [completion|risk|velocity|all]         # 予測分析
zeus report [--format text|html|markdown] [-o file] # プロジェクトレポート生成

# Phase 5 (ダッシュボード)
zeus dashboard [--port 8080] [--no-open]            # Web ダッシュボードを起動
```

### ダッシュボード機能

Web ブラウザでプロジェクト状態を可視化:

| 機能 | 説明 |
|------|------|
| プロジェクト概要 | 名前、説明、進捗率、健全性 |
| タスク統計 | 完了/進行中/保留の内訳 |
| タスク一覧 | テーブル形式、ステータス色分け |
| 依存関係グラフ | Mermaid.js でインタラクティブ表示 |
| 予測分析 | 完了日、リスク、ベロシティ |
| 自動更新 | 5秒間隔で Polling |

**API エンドポイント:**
- `GET /api/status` - プロジェクト状態
- `GET /api/tasks` - タスク一覧
- `GET /api/graph` - 依存関係グラフ（Mermaid形式）
- `GET /api/predict` - 予測分析結果

### 承認レベル

3段階の承認レベルをサポート:

| レベル | 説明 | 動作 |
|--------|------|------|
| auto | 自動承認 | 低リスク操作、即時実行 |
| notify | 通知のみ | 中リスク操作、ログ記録して実行 |
| approve | 明示的承認必要 | 高リスク操作、承認待ちキューに追加 |

**実装状態:**
- Simple レベル: 全操作が auto（承認フローなし）
- Standard: 追加操作は notify（通知のみ）
- Advanced: 追加操作は approve（事前承認必要）

`zeus add` 実行時、automation_level に応じて承認フローが自動適用されます。

### テスト

```bash
# 全テスト実行
go test ./...

# 詳細出力
go test -v ./internal/core/...

# カバレッジ計測
go test -cover ./...
```

テストカテゴリ:
- DI テスト: モック注入の検証
- Context タイムアウトテスト: キャンセル処理の検証
- 統合テスト: Init から List までのフロー
- TaskHandler 単体テスト: CRUD 操作の検証
- 分析テスト: グラフ構築、予測計算の検証
- レポートテスト: 各形式の出力検証
- ダッシュボードテスト: API ハンドラー、サーバー起動/停止の検証

### 次のステップ

- テスト強化
  - E2E テストの整備
  - テストカバレッジ 80% 達成
- ドキュメント整備
  - ユーザーガイドの作成
  - API リファレンスの整備
- Phase 6（将来）
  - 外部連携（Slack/Email 通知）
  - Git 統合
