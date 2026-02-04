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
| Predictor | 完了日予測、リスク分析、ベロシティ計算 |
| WBSBuilder | WBS 階層構築、ParentID 循環参照検出 |
| TimelineBuilder | タイムライン構築、クリティカルパス計算（CPM） |

**設計ポイント:**
- `analysis` パッケージは `core` からの import cycle を避けるため独自の型を定義
- `core.Zeus` から `analysis` への変換関数で連携
- WBS と依存関係グラフは独立した循環検出を持つ（ParentID vs Dependencies）

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

## データ構造 (.zeus/)

zeus init 実行後、ターゲットプロジェクトに生成される構造:

```
.zeus/
├── zeus.yaml              # プロジェクト定義（メイン）
├── config/                # 設定ファイル
├── activities/            # Activity 管理
│   └── act-NNN.yaml       # 個別 Activity
├── state/
│   ├── current.yaml       # 現在の状態
│   └── snapshots/         # 履歴スナップショット
├── entities/              # エンティティ定義
├── approvals/             # 承認管理
│   ├── pending/           # 承認待ち
│   ├── approved/          # 承認済み
│   └── rejected/          # 却下済み
├── logs/                  # ログ記録
├── analytics/             # 分析データ
├── graph/                 # 関係性グラフ
├── views/                 # カスタムビュー
├── backups/               # 自動バックアップ
└── .local/                # ローカル設定
```
