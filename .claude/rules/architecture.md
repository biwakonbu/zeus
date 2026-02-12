---
description: Zeus のアーキテクチャ詳細。Go コード編集時に参照。
paths:
  - "internal/**"
  - "cmd/**"
---

# アーキテクチャ

## コアモジュール (internal/core/)

| モジュール | 責務 |
|-----------|------|
| Zeus | メインロジック、プロジェクト初期化、コマンド実行 |
| StateManager | 状態スナップショット管理、履歴追跡 |
| ApprovalManager | 3段階承認フロー (auto/notify/approve)、ファイルロック |
| ActivityHandler | Activity エンティティの CRUD 操作 |
| EntityRegistry | エンティティハンドラーの登録・取得 |

## 分析モジュール (internal/analysis/)

| モジュール | 責務 |
|-----------|------|
| GraphBuilder | Activity 依存関係グラフの構築 |
| DependencyGraph | グラフ構造、循環検出、統計計算、可視化出力、下流/上流 Activity 取得 |
| UnifiedGraphBuilder | Activity/UseCase/Objective 統合グラフ構築（2層モデル） |
| AffinityCalculator | Objective ベースの親和性計算 |
| StaleAnalyzer | エンティティの陳腐化検出 |
| CoverageAnalyzer | UseCase/Activity カバレッジ分析 |
| GraphFilter | UnifiedGraph フィルタリング |

**設計ポイント:**
- `analysis` パッケージは `core` からの import cycle を避けるため独自の型を定義
- `core.Zeus` から `analysis` への変換関数で連携

## レポートモジュール (internal/report/)

| モジュール | 責務 |
|-----------|------|
| Generator | プロジェクトレポートの生成 |
| Templates | TEXT/HTML/Markdown テンプレート |

## DI パターン

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

## Context 対応

全ての公開 API が `context.Context` を第一引数として受け取る:
- タイムアウト制御
- キャンセル伝播
- 非同期処理のコントロール

## セキュリティ対策

| 対策 | 実装箇所 | 説明 |
|------|----------|------|
| ディレクトリトラバーサル防止 | file_manager.go | ValidatePath でパス検証 |
| ID 衝突防止 | zeus.go, approval.go | UUID v4 ベースの ID 生成 |
| 承認フロー原子性 | approval.go | flock ベースのファイルロック |
| ダッシュボードローカル専用 | server.go | 127.0.0.1 バインド |

## 4層階層モデル

Zeus は以下の4層でプロジェクトを構造化する:

| 層 | エンティティ | 役割 | 参照関係 |
|---|---|---|---|
| ゴール | Vision（単一） | 実現するべきゴール | `vision.yaml` |
| 目標 | Objective（フラット） | 測定可能な成果目標 | `objectives/obj-*.yaml` |
| 抽象 | UseCase | 本質的な求め | `usecases/uc-*.yaml`（`objective_id` 必須） |
| 具体 | Activity | 実現手段 | `activities/act-*.yaml`（`usecase_id` 任意） |

## データ構造 (.zeus/)

zeus init 実行後、ターゲットプロジェクトに生成される構造:

```
.zeus/
├── zeus.yaml              # プロジェクト設定
├── vision.yaml            # Vision（単一）
├── objectives/            # Objective（個別ファイル）
├── usecases/              # UseCase（個別ファイル）
├── activities/            # Activity（個別ファイル）
├── actors.yaml            # Actor 一覧
├── subsystems.yaml        # Subsystem 一覧
├── constraints.yaml       # Constraint 一覧
├── considerations/        # Consideration（個別ファイル）
├── decisions/             # Decision（個別ファイル）
├── problems/              # Problem（個別ファイル）
├── risks/                 # Risk（個別ファイル）
├── assumptions/           # Assumption（個別ファイル）
├── quality/               # Quality（個別ファイル）
└── state/                 # 状態管理
    ├── current.yaml
    └── snapshots/
```
