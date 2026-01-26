---
description: ダッシュボード（SvelteKit + PixiJS）の設計詳細。フロントエンド作業時に参照。
paths:
  - "zeus-dashboard/**"
---

# ダッシュボード

## バックエンドモジュール (internal/dashboard/)

| モジュール | 責務 |
|-----------|------|
| Server | HTTP サーバー管理、静的ファイル配信、CORS ミドルウェア |
| Handlers | REST API ハンドラー |
| SSEBroadcaster | SSE クライアント管理、リアルタイムイベント配信 |

**バックエンド設計ポイント:**
- Go 標準ライブラリのみ使用（net/http, embed）
- SvelteKit ビルド成果物は `//go:embed build/*` で埋め込み
- SSE で リアルタイム更新（ポーリングへのフォールバックあり）
- 開発モード (`--dev`) で CORS 有効化
- 127.0.0.1 にバインドしてローカルアクセスのみ許可
- SPA フォールバックルーティング対応

## フロントエンド (zeus-dashboard/)

**フロントエンド設計ポイント:**
- SvelteKit + TypeScript で型安全な開発
- Factorio 風インダストリアル UI テーマ
- PixiJS (WebGL) で高パフォーマンスなタスクグラフ描画
- Svelte Stores でリアクティブな状態管理
- SSE クライアントで自動再接続ロジック実装

## 統合戦略

| 環境 | 実行方法 | 静的ファイル配信 | API アクセス |
|------|----------|------------------|--------------|
| **開発時** | `make dashboard-dev` + `go run . dashboard --dev` | Vite Dev Server (:5173) | Go Server (:8080) + CORS |
| **本番時** | `make build-all` → `zeus dashboard` | Go embed (:8080) | 同一オリジン |

## API エンドポイント

- `GET /api/status` - プロジェクト状態
- `GET /api/tasks` - タスク一覧
- `GET /api/graph` - 依存関係グラフ（Mermaid形式）
- `GET /api/predict` - 予測分析結果
- `GET /api/wbs` - WBS 階層構造
- `GET /api/timeline` - タイムラインとクリティカルパス
- `GET /api/downstream?task_id=X` - 下流・上流タスク取得
- `GET /api/events` - SSE ストリーム（リアルタイム更新）
- `GET /api/affinity` - 機能間類似度マトリクス（Phase 7 で追加予定）
- `POST /api/metrics` - メトリクス保存（Graph View 計測ログ）
- `GET /api/actors` - Actor 一覧
- `GET /api/usecases` - UseCase 一覧
- `GET /api/subsystems` - Subsystem 一覧
- `GET /api/uml/usecase` - ユースケース図（Mermaid 形式）
- `GET /api/activities` - Activity 一覧
- `GET /api/uml/activity?id=X` - アクティビティ図（指定 ID のノード・遷移）

## ダッシュボード機能

| 機能 | 説明 |
|------|------|
| タスクグラフ | PixiJS によるインタラクティブな依存関係グラフ表示 |
| ミニマップ | 全体像の把握と素早いナビゲーション |
| フィルター | ステータス・優先度・担当者でフィルタリング |
| タスク詳細 | タスク選択時に詳細パネル表示 |
| リアルタイム更新 | SSE + ポーリングフォールバック |

## コマンドオプション

- `--port` - ポート番号（デフォルト: 8080）
- `--no-open` - ブラウザを自動で開かない
- `--dev` - 開発モード（CORS 有効）

## ビュー切り替え

- Graph View: 依存関係グラフ（Factorio 風）
- WBS View: 階層構造ツリー
- Timeline View: ガントチャート風表示
- Affinity Canvas: 機能間関連性可視化（Phase 7 で追加予定、設計書: `docs/design/affinity-canvas.md`）
- UseCaseView: UML ユースケース図（PixiJS ベース）
- ActivityView: UML アクティビティ図（PixiJS ベース）

## 影響範囲可視化

- 選択タスクの下流タスクを黄色でハイライト
- 上流タスクを青色でハイライト
- 選択タスクはオレンジ色で強調

## メトリクス計測（Graph View）

開発時やテスト時に Graph View の操作ログを収集する機能。

**有効化方法:**
- `http://localhost:5173/?metrics` でメトリクス収集を有効化
- `?metricsAutoSave` を付けると `/api/metrics` に自動保存（テストモードは自動で有効）

**出力:**
- 画面右上の `DL` ボタンで `zeus-viewer-metrics-*.json` をダウンロード
- 自動保存先: `.zeus/metrics/dashboard-metrics-<session>.jsonl`
- 収集ログは `window.__VIEWER_METRICS__` にも格納され、ステータスバーに件数が表示される

## UseCaseView

UML ユースケース図を表示するビュー。サブシステムによるグルーピングをサポート。

**サブシステム機能**:
- UseCase をサブシステムごとにグループ化表示
- サブシステム境界を UML 準拠の角丸矩形で描画
- ハッシュベースのカラー自動生成（HSL 色空間）
- 未分類 UseCase は「未分類」境界（グレー系）に配置

**レイアウト**: 3 カラム構成
- 左: Actor 一覧パネル + サブシステムフィルタ
- 中央: PixiJS ユースケース図（サブシステム境界付き）
- 右: 選択エンティティの詳細パネル

**インタラクション**:
| 操作 | 結果 |
|------|------|
| クリック（Actor/UseCase） | 詳細パネル表示 |
| ホバー | 関連エンティティハイライト |
| リフレッシュボタン | データ再取得 |

**API 連携**:
- `/api/actors` - Actor 一覧取得
- `/api/usecases` - UseCase 一覧取得
- `/api/subsystems` - Subsystem 一覧取得
- `/api/uml/usecase` - Mermaid 図取得

## ActivityView

UML アクティビティ図を表示するビュー。フルスクリーンキャンバスにオーバーレイパネル方式。

**ノードタイプ**:
| タイプ | UML 記号 | 説明 |
|--------|---------|------|
| `initial` | 黒丸 | 開始ノード |
| `final` | 二重丸 | 終了ノード |
| `action` | 角丸四角形 | アクション |
| `decision` | ひし形 | 分岐 |
| `merge` | ひし形 | 合流 |
| `fork` | 太い横線 | 並列分岐 |
| `join` | 太い横線 | 並列合流 |

**レイアウト**: オーバーレイパネル構成
- 左上: アクティビティ一覧パネル（デフォルト表示）
- 右上: 選択ノードの詳細パネル（選択時のみ表示）
- 中央: PixiJS キャンバス（フルスクリーン）

**インタラクション**:
| 操作 | 結果 |
|------|------|
| クリック（ノード） | 詳細パネル表示 + 選択状態 |
| ホバー | ツールチップ表示 |
| ドラッグ | キャンバスパン |
| ホイール | ズームイン/アウト |
| ESC | パネルを閉じる |

**API 連携**:
- `/api/activities` - アクティビティ一覧取得
- `/api/uml/activity?id=X` - 指定アクティビティのノード・遷移取得

**PixiJS レンダリングクラス**:
| クラス | 役割 |
|--------|------|
| ActivityEngine | エンジン管理、レイアウト計算、インタラクション |
| ActivityNodeBase | ノード基底クラス |
| InitialNode, FinalNode | 開始/終了ノード |
| ActionNode | アクションノード |
| DecisionNode, MergeNode | 分岐/合流ノード |
| ForkNode, JoinNode | 並列分岐/合流ノード |
| TransitionEdge | 遷移エッジ（矢印） |

## デザインガイドライン

### 禁止事項

- **Unicode Emoji の使用禁止** - Lucide Icons を使用すること
- **派手な色使い禁止** - Factorio 風の抑えた工業的配色を維持
- **過度なアニメーション禁止** - 200ms 以下に統一

### アイコンシステム

| 項目 | 仕様 |
|------|------|
| ライブラリ | Lucide Icons (lucide-svelte) |
| コンポーネント | `$lib/components/ui/Icon.svelte` |
| stroke-width | 2.5（デフォルト、太線化） |
| stroke-linecap | square（角ばったエッジ） |
| 効果 | `glow` prop で `filter: drop-shadow()` グロー効果 |

**Icon コンポーネント使用例:**
```svelte
<script>
  import { Icon } from '$lib/components/ui';
</script>

<Icon name="Heart" size={16} />
<Icon name="AlertTriangle" size={24} glow />
<Icon name="Settings" size={20} label="設定" />
```

**利用可能なアイコン一覧:**
- ナビゲーション: Heart, Calendar, Flame, RefreshCw, X
- 状態: AlertTriangle, CheckCircle, Info, XCircle
- アクション: ClipboardList, Target, BarChart, Ruler, ZoomIn, ZoomOut
- UI: Keyboard, Inbox, Settings

### UI コンポーネント

| コンポーネント | 用途 | パス |
|---------------|------|------|
| Icon | Lucide Icons ラッパー | `$lib/components/ui/Icon.svelte` |
| Panel | パネルコンテナ | `$lib/components/ui/Panel.svelte` |
| Badge | ステータスバッジ | `$lib/components/ui/Badge.svelte` |
| ProgressBar | 進捗バー | `$lib/components/ui/ProgressBar.svelte` |
| EmptyState | データなし状態 | `$lib/components/ui/EmptyState.svelte` |
| Toast | トースト通知 | `$lib/components/ui/Toast.svelte` |
| ToastContainer | トーストコンテナ | `$lib/components/ui/ToastContainer.svelte` |
| ContextMenu | コンテキストメニュー | `$lib/components/ui/ContextMenu.svelte` |
| KeyboardHelp | ショートカットヘルプ | `$lib/components/ui/KeyboardHelp.svelte` |

### Store システム

| Store | 用途 | パス |
|-------|------|------|
| toastStore | トースト通知管理 | `$lib/stores/toast.ts` |
| keyboardStore | キーボードショートカット | `$lib/stores/keyboard.ts` |
| connectionState | SSE 接続状態 | `$lib/stores/connection.ts` |

**Toast 使用例:**
```typescript
import { toastStore } from '$lib/stores/toast';

toastStore.success('保存しました');
toastStore.error('エラーが発生しました', { duration: 8000 });
toastStore.warning('注意してください');
toastStore.info('情報メッセージ');
```

**Keyboard Shortcut 使用例:**
```typescript
import { keyboardStore } from '$lib/stores/keyboard';

const unregister = keyboardStore.register({
  key: 'k',
  modifiers: ['cmd'],
  description: 'コマンドパレットを開く',
  category: 'ナビゲーション',
  action: () => openCommandPalette()
});

// 登録解除
unregister();
```

### アニメーション・トランジション

| 操作 | duration | easing |
|------|----------|--------|
| hover（即時反応） | 0ms | - |
| tooltip 表示 | 300ms | ease-out |
| select アニメーション | 150ms | ease-out |
| panel 開閉 | 200ms | ease-out |

**実装規則:**
- GPU アクセラレート必須（transform, opacity のみ使用）
- `prefers-reduced-motion` 対応必須
- CSS transform ベースで実装

### レイアウト

| 項目 | 仕様 |
|------|------|
| サイドペイン幅 | 固定 360px |
| レイアウト方式 | CSS Grid |
| リフロー制御 | `contain: layout` |

### サイドペイン

- **閉じる操作**: Escape キー、外部クリック、×ボタン
- **アニメーション**: 200ms ease-out、GPU アクセラレート
- **スクロール**: ヘッダー/フッター固定、位置記憶

### インタラクション

| 操作 | 挙動 |
|------|------|
| Click | 選択を置換 |
| Cmd+Click | 選択に追加 |
| Shift+Click | チェーン選択 |
| ダブルクリック | 詳細パネルを開く |
| 右クリック | コンテキストメニュー（最大8項目） |

### キーボードナビゲーション

- **Tab/Shift+Tab**: フォーカス移動
- **Enter**: 選択/実行
- **Escape**: パネル閉じる/選択解除
- **/ キー**: 検索フォーカス
- **? キー**: ショートカットヘルプ

**ビュー別移動ロジック:**
- Graph View: 依存関係順
- WBS View: リスト順
- Timeline View: 時系列順

### モバイル対応

| 項目 | 仕様 |
|------|------|
| breakpoint | 768px |
| サイドペイン | ボトムシート方式 |
| タッチターゲット | 最小 44px |

### パフォーマンス目標

| メトリクス | 必須目標 | 理想目標 |
|-----------|---------|---------|
| LCP | < 2.5s | < 1.5s |
| FID | < 100ms | < 50ms |
| CLS | < 0.1 | < 0.05 |
| グラフレンダリング (100 nodes) | < 1000ms | < 400ms |
| サイドペイン開閉 | < 200ms | < 100ms |
| アニメーション | 60fps | 60fps |

### Factorio 風デザイン要素（ミディアム）

| 要素 | 仕様 |
|------|------|
| ボーダー | 2px（上部明るく、下部暗く） |
| テクスチャ | 微細グラデーション（ノイズなし） |
| 効果 | インナーシャドウで凹み感 |
| 角丸 | 最小限（2-4px） |
