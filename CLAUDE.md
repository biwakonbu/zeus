# CLAUDE.md

Zeus は「神の視点」でプロジェクト管理を支援する AI 駆動型 CLI システム。Go + Cobra で実装。

## 設計哲学

- ファイルベース: 外部 DB 不要、YAML で人間可読
- 人間中心: AI は提案者、人間が最終決定者
- シンプルな初期化: 単一の `zeus init` コマンドで全機能を利用可能
- Git 親和性: 全データがテキストで差分追跡可能

## 技術スタック

**バックエンド:** Go 1.21+, Cobra, gopkg.in/yaml.v3, fatih/color, github.com/google/uuid

> **Note:** Go 1.21+ は `min()`, `max()` 組み込み関数のため必須。`slices` パッケージも使用。

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
| Phase 7 (Affinity Canvas) | 機能間関連性可視化、フォースダイレクテッド | 完了 |
| 10概念モデル Phase 1 | Vision, Objective, Deliverable, 参照整合性 | 完了 |
| 10概念モデル Phase 2+3 | Consideration, Decision, Problem, Risk, Assumption, Constraint, Quality | 完了 (M1-M3対応推奨) |
| UML UseCase | Actor, UseCase, シナリオ、PixiJS ビューワー | 完了 |
| UML Activity | アクティビティ図、ノード/遷移、PixiJS ビューワー | 完了 |
| UML Subsystem | サブシステム分類、UseCase グルーピング、境界描画 | 完了 |

## 実装済みコマンド

```bash
# コア操作
zeus init                                       # プロジェクト初期化
zeus status                                     # 状態表示
zeus add <entity> <name> [options]              # エンティティ追加
  # entity: task, vision, objective, deliverable, consideration, decision,
  #         problem, risk, assumption, constraint, quality, actor, usecase, subsystem
  # --parent <id>  --start <date>  --due <date>  --progress <0-100>  --wbs <code>
  # --statement <text>  --objective <id>  --format <type>  --subsystem <id>
zeus list [entity]                              # 一覧表示
  # entity: tasks, vision, objectives, deliverables, considerations, decisions,
  #         problems, risks, assumptions, constraints, quality, actors, usecases, subsystems
zeus doctor                                     # 診断（参照整合性・循環参照チェック含む）
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

## 10概念モデル

Task ベースのシステムを拡張し、プロジェクト管理の本質的な概念を表現する 10 概念モデルを導入。

### Phase 1 実装済み（3概念）

| 概念 | 説明 | ファイル |
|------|------|----------|
| Vision | プロジェクトの目指す姿（単一） | `.zeus/vision.yaml` |
| Objective | 達成目標（階層構造可） | `.zeus/objectives/obj-NNN.yaml` |
| Deliverable | 成果物定義 | `.zeus/deliverables/del-NNN.yaml` |

### Phase 2 実装済み（5概念）

| 概念 | 説明 | ファイル | 特性 |
|------|------|----------|------|
| Consideration | 検討事項（複数オプション） | `.zeus/considerations/con-NNN.yaml` | 検討プロセス記録 |
| Decision | 意思決定（イミュータブル） | `.zeus/decisions/dec-NNN.yaml` | 一度決定後は変更不可 |
| Problem | 問題報告 | `.zeus/problems/prob-NNN.yaml` | 重大度レベル記録 |
| Risk | リスク管理 | `.zeus/risks/risk-NNN.yaml` | スコア自動計算 |
| Assumption | 前提条件 | `.zeus/assumptions/assum-NNN.yaml` | 検証ステータス記録 |

### Phase 3 実装済み（2概念）

| 概念 | 説明 | ファイル | 特性 |
|------|------|----------|------|
| Constraint | 制約条件 | `.zeus/constraints.yaml` | グローバル単一ファイル |
| Quality | 品質基準 | `.zeus/quality/qual-NNN.yaml` | メトリクス・ゲート管理 |

### UML 拡張

| 概念 | 説明 | ファイル | 特性 |
|------|------|----------|------|
| Actor | アクター定義 | `.zeus/actors.yaml` | 単一ファイル |
| UseCase | ユースケース定義 | `.zeus/usecases/uc-NNN.yaml` | Objective 参照必須 |
| Subsystem | サブシステム定義 | `.zeus/subsystems.yaml` | 単一ファイル、UseCase グルーピング |
| Activity | アクティビティ図 | `.zeus/activities/act-NNN.yaml` | UseCase 参照任意 |

### 参照整合性

- `zeus doctor` で全参照をチェック：
  - Deliverable → Objective（必須）
  - Objective → Objective (親)（任意、循環参照チェック）
  - Decision → Consideration（必須）
  - Quality → Deliverable（必須）
  - Problem/Risk/Assumption → Objective/Deliverable（任意）
  - UseCase → Subsystem（任意、警告レベル）
- 循環参照検出実装済み
- セキュリティ: ValidatePath, ValidateID, Sanitizer

### コードレビュー結果（Phase 2+3）

**実装完了度:** 95% | **コード品質:** 85-90%

**指摘事項 (優先度順):**
1. M1: Decision の Delete も禁止化すべき（イミュータブル制約）- 1時間
2. M3: Decision/Consideration の逆参照整合性チェック追加 - 2時間
3. M2: Quality メトリクス CLI 実装完了 - 2時間（中期）
4. M5: ID 生成パフォーマンス改善（O(N)→O(1)）- 3時間（中期）

**強み:**
- EntityHandler パターンの一貫性が高い
- セキュリティ検証（パストラバーサル、インジェクション対策）堅牢
- 参照整合性チェックが網羅的
- テスト成功率 100%

**推奨対応:** Priority 1 の 2 タスク（計 5時間）対応後、本番展開可能

詳細: REVIEW_PHASE_2_3.md / REVIEW_SUMMARY.json 参照

## ドキュメント

- `docs/system-design.md` - システム設計書（必読）
- `docs/implementation-guide.md` - Go 実装ガイド
- `docs/operations-manual.md` - 運用マニュアル
- `docs/detailed-design.md` - 10概念モデル詳細設計
- `docs/design/affinity-canvas.md` - Affinity Canvas 設計書（Phase 7）
- `docs/api-spec.md` - API 仕様（将来実装）
- `docs/security.md` - セキュリティ実装ガイド

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
