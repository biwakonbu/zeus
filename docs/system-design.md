# Zeus システム設計書

## 1. システム概要

### 1.1 ビジョン
Zeusは、AIによるプロジェクトマネジメントを「神の視点」で俯瞰し、上流工程（方針立案からWBS化、タイムライン設計、仕様作成まで）を支援するCLIベースのシステムです。

### 1.2 コアコンセプト
- **神の視点（Zeus View）**: プロジェクト全体を俯瞰し、依存関係、リソース配分、進捗を一元的に把握
- **ファイルベース**: 依存ミドルウェアゼロ、YAMLで可読性とGit親和性を確保
- **人間中心**: AIは助言者、人間が最終決定者
- **シンプルな初期化**: 単一の `zeus init` コマンドで全機能を利用可能

### 1.3 対象ユーザー
1. プロジェクトマネージャー
2. 技術リーダー/アーキテクト
3. プロダクトマネージャー
4. 経営層/ステークホルダー

## 2. アーキテクチャ

### 2.1 システム構成図

```
┌─────────────────────────────────────────────────────────────┐
│                         User Interface                       │
│  ┌─────────────────────┐   ┌─────────────────────┐         │
│  │    Zeus CLI          │   │   Web Dashboard     │         │
│  └──────────┬──────────┘   └──────────┬──────────┘         │
├─────────────┴──────────────────────────┴────────────────────┤
│                         Core Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Command     │  │  Approval   │  │  AI Engine  │        │
│  │  Processor   │  │  Manager    │  │  Interface  │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│         └────────────┬────┴────────────────┘                │
│                      │                                       │
│  ┌─────────────────────────────────────────────────┐       │
│  │               State Manager                       │       │
│  └─────────────────────────────────────────────────┘       │
├───────────────────────────────┼─────────────────────────────┤
│                       Analysis Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Dependency  │  │  Prediction │  │  Report     │        │
│  │  Graph       │  │  Engine     │  │  Generator  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├───────────────────────────────┼─────────────────────────────┤
│                         Data Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  zeus.yaml   │  │  Activities │  │  State      │        │
│  │  (Core)      │  │  Store      │  │  Snapshots  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Entities    │  │  Approvals  │  │  Analytics  │        │
│  │              │  │  Queue      │  │  Tracking   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 ディレクトリ構造

`zeus init` 実行後に生成される統一構造:

```
.zeus/
├── zeus.yaml        # メインプロジェクト定義
├── config/          # 設定ファイル
├── activities/      # アクティビティ管理
├── state/           # 状態管理
├── entities/        # エンティティ定義
├── approvals/       # 承認管理
│   ├── pending/     # 承認待ち
│   ├── approved/    # 承認済み
│   └── rejected/    # 却下済み
├── logs/            # ログ記録
├── analytics/       # 分析データ
├── graph/           # 関係性グラフ
├── views/           # カスタムビュー
├── backups/         # 自動バックアップ
└── .local/          # ローカル設定

.claude/             # Claude Code 連携（常に生成）
├── agents/          # Zeus 用エージェント
└── skills/          # Zeus 用スキル
```

### 2.3 パッケージ構成

#### 2.3.1 コアパッケージ (internal/core/)

| モジュール | 責務 |
|-----------|------|
| Zeus | メインロジック、プロジェクト初期化、コマンド実行 |
| StateManager | 状態スナップショット管理、履歴追跡 |
| ApprovalManager | 3段階承認フロー (auto/notify/approve)、ファイルロック |
| ActivityHandler | Activity エンティティの CRUD 操作 |
| EntityRegistry | エンティティハンドラーの登録・取得 |

#### 2.3.2 分析パッケージ (internal/analysis/)

| モジュール | 責務 |
|-----------|------|
| types.go | 分析用型定義（core への依存を避けるため独立） |
| GraphBuilder | タスク依存関係グラフの構築 |
| DependencyGraph | グラフ構造、循環検出、統計計算、可視化出力、下流/上流タスク取得 |
| Predictor | 完了日予測、リスク分析、ベロシティ計算 |
| WBSBuilder | WBS 階層構築、ParentID 循環参照検出 |
| TimelineBuilder | タイムライン構築、クリティカルパス計算（CPM） |

**設計ポイント:**
- `analysis` パッケージは `core` からの import cycle を避けるため独自の型を定義
- `core.Zeus` から `analysis` への変換関数で連携
- WBS と依存関係グラフは独立した循環検出を持つ（ParentID vs Dependencies）

#### 2.3.3 レポートパッケージ (internal/report/)

| モジュール | 責務 |
|-----------|------|
| Generator | プロジェクトレポートの生成 |
| Templates | TEXT/HTML/Markdown テンプレート |

#### 2.3.4 ダッシュボードパッケージ (internal/dashboard/)

| モジュール | 責務 |
|-----------|------|
| Server | HTTP サーバー管理、静的ファイル配信 |
| Handlers | REST API ハンドラー（/api/status, /api/activities, /api/graph, /api/predict） |

**設計ポイント:**
- Go 標準ライブラリのみ使用（net/http, embed）
- 静的ファイルは `//go:embed` で埋め込み
- Mermaid.js は CDN から読み込み
- 127.0.0.1 にバインドしてローカルアクセスのみ許可

### 2.4 データフォーマット

#### 2.4.1 メイン設定（zeus.yaml）
```yaml
# .zeus/zeus.yaml
version: "1.0"
project:
  id: "project-zeus-001"
  name: "Zeus Development"
  description: "AI-driven project management system"
  start_date: "2024-01-14"

objectives:
  - id: "obj-001"
    title: "MVP Release"
    deadline: "2024-03-31"
    priority: "high"

settings:
  automation_level: "auto"        # auto|notify|approve（デフォルト: auto）
  approval_mode: "default"        # default|strict|loose
  ai_provider: "claude-code"      # claude-code|gemini|codex
```

#### 2.4.2 Activity 定義
```yaml
# .zeus/activities/act-001.yaml
id: "act-001"
title: "Design core data structure"
status: "in_progress"
assignee: "ai"
estimate_hours: 8
actual_hours: null
dependencies: []
approval_level: "auto"
```

#### 2.4.3 状態スナップショット
```yaml
# .zeus/state/current.yaml
snapshot:
  timestamp: "2024-01-14T16:00:00Z"
  summary:
    total_activities: 42
    completed: 15
    in_progress: 10
    pending: 17
  health: "good"
  risks: []
```

## 3. 機能仕様

### 3.1 コアコマンド

#### 3.1.1 初期化
```bash
zeus init
```
- プロジェクトディレクトリを初期化
- 統一された構造を生成（.zeus/ と .claude/）
- デフォルトの `automation_level` は `auto`（即時実行、承認不要）

**Claude Code 連携ファイルの更新:**
```bash
zeus update-claude
```
- 既存プロジェクトの `.claude/` ディレクトリ内ファイルを最新テンプレートで再生成
- `zeus init` を再実行せずに、Claude Code 連携ファイルのみを更新可能
- Zeus のバージョンアップ後に新機能をエージェント・スキルに反映する際に使用

#### 3.1.2 状態確認
```bash
zeus status
```
- プロジェクト全体の状態を表示
- 3層ビュー（Quick/Detailed/Rich）に対応

#### 3.1.3 エンティティ操作
```bash
zeus add <entity> <name>
zeus update <entity-id> <field> <value>
zeus list [entity] [--filter]
```

### 3.2 AI機能

#### 3.2.1 提案システム
```bash
zeus suggest                  # AI提案一覧
zeus suggest --detail <id>    # 提案詳細
zeus apply <id> [--dry-run]   # 提案適用
```

#### 3.2.2 説明機能
```bash
zeus explain <entity-id>      # AIによる解説
```

### 3.3 承認フロー

#### 3.3.1 承認レベル
1. **auto**: 自動実行（読み取り、計算、レポート生成）- デフォルト
2. **notify**: 通知付き実行（ステータス更新、軽微な変更）
3. **approve**: 事前承認必須（重要な変更、外部連携）

承認レベルは `zeus.yaml` の `automation_level` で設定:
- `auto`: 全操作が即時実行（承認フローなし）
- `notify`: 追加操作は通知のみ（ログ記録して実行）
- `approve`: 追加操作は事前承認必要（承認待ちキューに追加）

#### 3.3.2 承認コマンド
```bash
zeus pending              # 承認待ち一覧
zeus approve <id>         # 承認
zeus reject <id> [reason] # 却下
```

### 3.4 分析機能（Phase 4）

#### 3.4.1 依存関係グラフ
```bash
zeus graph                            # テキスト形式で表示
zeus graph --format dot               # DOT形式（Graphviz互換）
zeus graph --format mermaid           # Mermaid形式
zeus graph --format mermaid -o graph.md  # ファイル出力
```

**出力形式:**

| 形式 | 説明 | 用途 |
|------|------|------|
| text | ASCII アート風テキスト | CLI での確認 |
| dot | Graphviz DOT 言語 | 画像生成（png, svg） |
| mermaid | Mermaid 記法 | Markdown ドキュメント |

**グラフ機能:**
- タスク間の依存関係可視化
- 循環依存の検出と警告
- クリティカルパスの強調表示
- 統計情報（ノード数、エッジ数、平均依存数）

#### 3.4.2 予測分析
```bash
zeus predict                  # 全予測を表示
zeus predict completion       # 完了日予測のみ
zeus predict risk             # リスク分析のみ
zeus predict velocity         # ベロシティ分析のみ
```

**予測モデル:**

| 予測種別 | アルゴリズム | 出力 |
|---------|-------------|------|
| 完了日予測 | 残タスク / ベロシティ | 予測完了日、信頼区間 |
| リスク分析 | 見積精度、依存関係複雑度 | リスクレベル（high/medium/low）、要因 |
| ベロシティ | 過去実績の移動平均 | タスク/日、トレンド |

#### 3.4.3 レポート生成
```bash
zeus report                           # テキスト形式で標準出力
zeus report --format html             # HTML形式
zeus report --format markdown         # Markdown形式
zeus report --format html -o report.html  # ファイル出力
```

**レポート内容:**
1. プロジェクト概要
2. 進捗サマリー（完了率、残タスク数）
3. タスク一覧（ステータス別）
4. 依存関係グラフ（Mermaid形式）
5. 予測分析結果
6. リスク・課題

### 3.5 Web ダッシュボード（Phase 5）

#### 3.5.1 ダッシュボード起動
```bash
zeus dashboard                # デフォルトポート 8080 で起動
zeus dashboard --port 3000    # カスタムポート
zeus dashboard --no-open      # ブラウザ自動起動を無効化
```

#### 3.5.2 REST API

| エンドポイント | メソッド | 説明 |
|---------------|---------|------|
| `/api/status` | GET | プロジェクト状態（名前、進捗率、健全性） |
| `/api/activities` | GET | Activity 一覧（JSON配列） |
| `/api/graph` | GET | 依存関係グラフ（Mermaid形式） |
| `/api/predict` | GET | 予測分析結果 |
| `/api/wbs` | GET | WBS 階層構造（Phase 6） |
| `/api/timeline` | GET | タイムラインとクリティカルパス（Phase 6） |
| `/api/downstream` | GET | 下流・上流 Activity 取得（Phase 6） |

#### 3.5.3 Factorio 風ビューワー

ダッシュボードは「神の視点」で1000以上のタスクを俯瞰できる Factorio 風ビューワーを採用。SSE（Server-Sent Events）によるリアルタイム更新をサポート。

**アーキテクチャ:**

```
┌─────────────────────────────────────────────────────────┐
│                  FactorioViewer.svelte                    │
│  ┌─────────────────────────────────────────────────┐    │
│  │            PixiJS Canvas Layer                    │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐      │    │
│  │  │ TaskNode │  │ TaskEdge │  │ Minimap  │      │    │
│  │  └──────────┘  └──────────┘  └──────────┘      │    │
│  └─────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Svelte UI Layer                      │    │
│  │  ┌──────────────┐  ┌──────────────┐            │    │
│  │  │ FilterPanel  │  │ Tooltip      │            │    │
│  │  └──────────────┘  └──────────────┘            │    │
│  └─────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────┤
│                      Engine Layer                         │
│  ┌──────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Viewer   │  │ Layout       │  │ Spatial      │      │
│  │ Engine   │  │ Engine       │  │ Index        │      │
│  └──────────┘  └──────────────┘  └──────────────┘      │
├─────────────────────────────────────────────────────────┤
│                   Interaction Layer                       │
│  ┌──────────────────┐  ┌──────────────────┐            │
│  │ SelectionManager │  │ FilterManager    │            │
│  └──────────────────┘  └──────────────────┘            │
└─────────────────────────────────────────────────────────┘
```

**技術選定:**

| コンポーネント | 技術 | 理由 |
|--------------|------|------|
| 描画エンジン | PixiJS 8.x (WebGL) | 1000以上のノードで60fps維持 |
| 空間インデックス | Quadtree | O(log n) での可視範囲クエリ |
| レイアウト | トポロジカルソート + 力学シミュレーション | 依存関係を反映した自動配置 |
| UI | Svelte 5 Runes | リアクティブな状態管理 |

**LOD（Level of Detail）システム:**

| ズームレベル | 表示内容 | 対象ノード数 |
|-------------|---------|------------|
| Macro（< 0.3） | ステータス色のみ | 1000+ |
| Meso（0.3-0.7） | ステータス + ID | 100-200 |
| Micro（> 0.7） | 全情報（タイトル、担当者、進捗） | 10-50 |

**パフォーマンス最適化:**

1. **Quadtree 空間インデックス**: 画面内のノードのみを O(log n) で取得
2. **仮想化レンダリング**: 画面外のノードは描画スキップ
3. **バッチレンダリング**: PixiJS の自動バッチ処理を活用
4. **LOD システム**: ズームレベルに応じて描画詳細度を調整

**インタラクション:**

| 操作 | 動作 |
|------|------|
| マウスホイール | ズームイン/アウト（ポインタ位置を中心に） |
| ドラッグ | パン（ビューポート移動） |
| クリック | タスク選択 |
| Ctrl+クリック | 複数選択に追加/削除 |
| Shift+クリック | 範囲選択 |
| ダブルクリック | タスクにフォーカス |

**ファイル構成:**

```
zeus-dashboard/src/lib/viewer/
├── FactorioViewer.svelte      # メインコンポーネント
├── index.ts                   # エクスポート
├── engine/
│   ├── ViewerEngine.ts        # PixiJS 初期化・管理
│   ├── LayoutEngine.ts        # 自動レイアウト（トポロジカル + 力学）
│   └── SpatialIndex.ts        # Quadtree 空間インデックス
├── rendering/
│   ├── TaskNode.ts            # ノード描画（LOD対応）
│   └── TaskEdge.ts            # エッジ描画（通常/クリティカル/ブロック）
├── interaction/
│   ├── SelectionManager.ts    # 選択状態管理
│   └── FilterManager.ts       # フィルター管理
└── ui/
    ├── Minimap.svelte         # ミニマップ
    └── FilterPanel.svelte     # フィルターパネル
```

### 3.6 WBS・タイムライン機能（Phase 6）

#### 3.6.1 データモデル拡張

Activity 型に以下のフィールドを追加（全て optional、後方互換性維持）:

```yaml
# Activity 定義の拡張フィールド
activities:
  - id: "act-001"
    title: "Design core data structure"
    # 既存フィールド...

    # Phase 6 拡張フィールド
    parent_id: "act-000"       # 親 Activity ID（WBS階層）
    start_date: "2026-01-01"   # 開始日（ISO8601）
    due_date: "2026-01-15"     # 期限日（ISO8601）
    progress: 75               # 進捗率（0-100）
    wbs_code: "1.2.3"          # WBS番号
```

#### 3.6.2 WBS 機能

**コマンド:**
```bash
zeus add activity "子 Activity" --parent <parent-activity-id>
```

**API レスポンス（/api/wbs）:**
```json
{
  "roots": [
    {
      "id": "act-001",
      "title": "プロジェクト設計",
      "wbs_code": "1",
      "status": "in_progress",
      "progress": 80,
      "children": [
        {
          "id": "act-002",
          "title": "要件定義",
          "wbs_code": "1.1",
          "status": "completed",
          "progress": 100,
          "children": []
        }
      ]
    }
  ],
  "max_depth": 3,
  "stats": {
    "total_nodes": 15,
    "root_count": 3,
    "leaf_count": 10,
    "avg_progress": 65.5,
    "completed_pct": 40.0
  }
}
```

**循環参照検出:**
- ParentID の循環参照を DFS アルゴリズムで検出
- 検出時はエラーを返却（WBS 構築を中止）
- Dependencies の循環検出とは独立して動作

#### 3.6.3 タイムライン機能

**API レスポンス（/api/timeline）:**
```json
{
  "items": [
    {
      "activity_id": "act-001",
      "title": "要件定義",
      "start_date": "2026-01-01",
      "end_date": "2026-01-15",
      "progress": 100,
      "status": "completed",
      "is_on_critical_path": true,
      "slack": 0,
      "dependencies": []
    }
  ],
  "critical_path": ["act-001", "act-003", "act-005"],
  "project_start": "2026-01-01",
  "project_end": "2026-03-31",
  "total_duration": 90,
  "stats": {
    "total_activities": 20,
    "activities_with_dates": 15,
    "on_critical_path": 5,
    "average_slack": 3.5,
    "overdue_activities": 2
  }
}
```

**クリティカルパス計算:**
- CPM（Critical Path Method）アルゴリズムを使用
- 各 Activity の slack（余裕時間）を計算
- クリティカルパス上の Activity は slack = 0

#### 3.6.4 影響範囲可視化

**API レスポンス（/api/downstream?id=X）:**
```json
{
  "activity_id": "act-003",
  "downstream": ["act-005", "act-007", "act-008"],
  "upstream": ["act-001", "act-002"],
  "count": 5
}
```

**フロントエンド機能:**
- ノードホバー時に下流 Activity を黄色でハイライト
- 上流 Activity を青色でハイライト
- 選択 Activity はオレンジ色で強調

#### 3.6.5 ビュー切り替え

ダッシュボードで 3 つのビューを切り替え可能:

| ビュー | 説明 | 主な用途 |
|-------|------|---------|
| Graph View | 依存関係グラフ（Factorio 風） | 依存関係の確認 |
| WBS View | 階層構造ツリー | 作業分解の確認 |
| Timeline View | ガントチャート風表示 | スケジュール確認 |

### 3.7 フィードバックシステム

#### 3.7.1 自動追跡
- Activity 完了時の見積もり精度
- オーバーライド時の差分
- 承認/却下の結果

#### 3.7.2 明示的フィードバック
```bash
zeus ok <id>              # 成功報告
zeus ng <id> [reason]     # 失敗報告
zeus review               # 週次レビュー
zeus stats                # 精度統計
```

### 3.8 エラー処理と復旧

#### 3.8.1 診断と修復
```bash
zeus doctor               # システム診断
zeus fix [--dry-run]      # 自動修復
```

#### 3.8.2 バックアップと復元
```bash
zeus restore [point]      # バックアップ復元
zeus restore --latest     # 最新から復元
```

## 4. CLIインターフェース設計

### 4.1 コマンド体系
- 動詞＋名詞の一貫した構造
- 短縮形とエイリアスのサポート
- --dry-runオプションの統一的提供

### 4.2 出力形式
- デフォルト: 人間可読なテキスト
- --format=json: プログラム連携用
- --format=html: レポート出力用

### 4.3 対話的操作
- 危険な操作には確認プロンプト
- zeus reviewでの対話的フィードバック
- $EDITORを使用した直接編集

## 5. Agent Skills設計

### 5.1 Zeus Agent
Claude Code Pluginとして実装し、以下のスキルを提供：

1. **project-scan**: プロジェクト構造の自動認識
2. **task-suggest**: タスク分割と優先順位提案
3. **risk-analysis**: リスク分析と対策提案
4. **timeline-optimize**: タイムライン最適化

### 5.2 実装パターン
```javascript
// skills/project-scan/skill.js
export async function projectScan(context) {
  // プロジェクトファイルを探索
  // 依存関係を分析
  // zeus.yamlを生成
}
```

## 6. 拡張性設計

### 6.1 プラグインシステム
- .zeus/plugins/ディレクトリでカスタムスキル定義
- JavaScript/TypeScriptでの実装
- 標準APIの提供

### 6.2 外部連携（将来）
- Git統合（Phase 6）
- Slack/Email通知（Phase 6）
- 他のCLIツールとの連携（Phase 6）

## 7. セキュリティとプライバシー

### 7.1 データ保護
- ローカルファイルのみ使用
- 外部送信は明示的承認が必要
- .localディレクトリはgitignore推奨

### 7.2 アクセス制御
- ファイルシステムの権限に依存
- 承認履歴の完全記録
- オーバーライドの監査証跡

### 7.3 ダッシュボードセキュリティ
- 127.0.0.1 バインドでローカルアクセスのみ許可
- 外部ネットワークからのアクセス不可
- 認証機能は将来実装予定（Phase 6）

## 8. パフォーマンス目標

### 8.1 レスポンスタイム
- zeus status: < 100ms
- zeus suggest: < 3s
- zeus doctor: < 5s
- zeus graph: < 1s（タスク1000件まで）
- zeus predict: < 500ms
- zeus report: < 2s
- zeus dashboard起動: < 1s

### 8.2 スケーラビリティ
- タスク数: ~10,000まで対応
- 履歴保持: 90日分
- ファイルサイズ: 各YAMLは10MB以下

## 9. 実装優先順位

### 9.1 Phase 1（MVP）- 完了
1. zeus.yaml構造定義
2. 基本CLIコマンド（init, status, add, list）
3. doctor/fix機能

### 9.2 Phase 2（標準機能）- 完了
1. 3段階承認レベル（pending, approve, reject）
2. スナップショット・履歴管理
3. セキュリティ強化（パス検証、UUID、ファイルロック）
4. DI/Context対応（テスタビリティ向上）

### 9.3 Phase 2.7（提案機能）- 完了
1. suggest コマンド（ルールベース提案）
2. apply コマンド（提案適用）
3. 影響度フィルタリング

### 9.4 Phase 3（AI統合）- 完了
1. Claude Code 連携（.claude/ 自動生成）
2. explain コマンド（エンティティ詳細説明）
3. Add コマンドと承認フローの連携
4. priority_change / dependency 提案タイプ対応
5. update-claude コマンド（Claude Code ファイルの再生成）

**Claude Code 連携テンプレート（Phase 6 対応済み）:**
- `zeus-orchestrator.md` - 全コマンド一覧、ダッシュボード、Phase 6 機能
- `zeus-planner.md` - WBS階層作成、タイムライン設計、依存関係
- `zeus-reviewer.md` - 分析ツール、Phase 6 レビュー項目
- `zeus-suggest/SKILL.md` - Activity、リスク軽減、優先度変更などを提案
- `zeus-risk-analysis/SKILL.md` - Phase 6 固有のリスク分析
- `zeus-wbs-design/SKILL.md` - WBS 階層設計のガイド
- `zeus-e2e-tester/SKILL.md` - E2E テストスキル

### 9.5 Phase 4（高度な分析）- 完了
1. 依存関係グラフの可視化（text/dot/mermaid）
2. 予測分析（完了日予測、リスク予測、ベロシティ）
3. レポート生成機能（text/html/markdown）

### 9.6 Phase 5（Web ダッシュボード）- 完了
1. HTTP サーバー（Go 標準 net/http）
2. REST API（/api/status, /api/activities, /api/graph, /api/predict, /api/events）
3. 静的ファイル埋め込み（//go:embed）
4. Mermaid.js によるグラフ表示
5. SSE（Server-Sent Events）によるリアルタイム更新

### 9.7 Phase 5.5（Factorio 風ビューワー）- 完了
1. PixiJS 8.x による WebGL 描画
2. Quadtree 空間インデックス（O(log n) クエリ）
3. LOD システム（Macro/Meso/Micro）
4. 自動レイアウト（トポロジカルソート + 力学シミュレーション）
5. 選択・フィルター機能
6. ミニマップ

### 9.8 Phase 6（WBS・タイムライン・依存関係強化）- 完了
1. データモデル拡張（ParentID, StartDate, DueDate, Progress, WBSCode）
2. WBS 階層構築・表示
3. タイムライン・CPM 計算
4. クリティカルパス表示
5. ParentID 循環参照検出
6. 影響範囲可視化（下流/上流タスクのハイライト）
7. ビュー切り替え UI（Graph/WBS/Timeline）
8. API エンドポイント（/api/wbs, /api/timeline, /api/downstream）

### 9.9 Phase 7（外部連携）- 未実装
1. Git 統合（コミット履歴との連携）
2. Slack/Email 通知
3. 認証機能（ダッシュボード）
4. 他のCLIツールとの連携

## 10. 設計決定の根拠

### 10.1 ファイルベースを選択した理由
- 依存関係の最小化
- Git親和性の確保
- 可読性とデバッグの容易さ
- バックアップとリカバリの簡単さ

### 10.2 YAMLを選択した理由
- 人間可読性の高さ
- 構造化データの表現力
- コメント機能
- 広範なツールサポート

### 10.3 CLIを選択した理由
- 開発者フレンドリー
- 自動化との親和性
- リモート操作の容易さ
- Claude Code等との統合

### 10.4 Go標準ライブラリを優先した理由（ダッシュボード）
- 依存関係の最小化
- 単一バイナリでの配布
- セキュリティリスクの低減
- 長期的なメンテナンス性

### 10.5 init コマンドの簡略化（2026-01-16）
- **変更前**: `zeus init --level=simple|standard|advanced`
- **変更後**: `zeus init`（オプションなし）
- **理由**:
  - ユーザーの認知負荷を軽減
  - 「Cursor yolo mode」スタイル: 即時実行、承認不要
  - 承認フローは `zeus.yaml` の `automation_level` で個別設定可能
  - 全機能を統一的に提供（必要に応じて設定で調整）

---

*Zeus System Design Document v1.4*
*作成日: 2026-01-14*
*更新日: 2026-01-17（Phase 6: WBS・タイムライン・依存関係強化）*
