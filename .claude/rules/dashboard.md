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
- PixiJS (WebGL) で高パフォーマンスな Activity グラフ描画
- Svelte Stores でリアクティブな状態管理
- SSE クライアントで自動再接続ロジック実装

## 統合戦略

| 環境 | 実行方法 | 静的ファイル配信 | API アクセス |
|------|----------|------------------|--------------|
| **開発時** | `make dashboard-dev` + `go run . dashboard --dev` | Vite Dev Server (:5173) | Go Server (:8080) + CORS |
| **本番時** | `make build-all` → `zeus dashboard` | Go embed (:8080) | 同一オリジン |

## API エンドポイント

- `GET /api/status` - プロジェクト状態
- `GET /api/activities` - Activity 一覧
- `GET /api/graph` - 依存関係グラフ（Mermaid形式）
- `GET /api/predict` - 予測分析結果
- `GET /api/wbs` - WBS 階層構造
- `GET /api/timeline` - タイムラインとクリティカルパス
- `GET /api/downstream?id=X` - 下流・上流 Activity 取得
- `GET /api/events` - SSE ストリーム（リアルタイム更新）
- `GET /api/affinity` - 機能間類似度マトリクス（Phase 7 で追加予定）
- `GET /api/actors` - Actor 一覧
- `GET /api/usecases` - UseCase 一覧
- `GET /api/subsystems` - Subsystem 一覧
- `GET /api/uml/usecase` - ユースケース図（Mermaid 形式）
- `GET /api/activities` - Activity 一覧
- `GET /api/uml/activity?id=X` - アクティビティ図（指定 ID のノード・遷移）

## ダッシュボード機能

| 機能 | 説明 |
|------|------|
| Activity グラフ | PixiJS によるインタラクティブな依存関係グラフ表示 |
| ミニマップ | 全体像の把握と素早いナビゲーション |
| フィルター | ステータス・優先度・担当者でフィルタリング |
| Activity 詳細 | Activity 選択時に詳細パネル表示 |
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

- 選択 Activity の下流 Activity を黄色でハイライト
- 上流 Activity を青色でハイライト
- 選択 Activity はオレンジ色で強調

## UseCaseView

UML ユースケース図を表示するビュー。サブシステムによるグルーピングをサポート。

**サブシステム機能**:
- UseCase をサブシステムごとにグループ化表示
- サブシステム境界を UML 準拠の角丸矩形で描画
- DJB2 ハッシュベースのカラー自動生成（HSL 色空間）
- 未分類 UseCase は「未分類」境界（グレー系）に配置
- サブシステムデータは `/api/subsystems` から並列取得

**フィルタモード**:
- デフォルトで有効（選択するまで図は非表示）
- Actor/UseCase をクリックすると関連エンティティのみ表示
- 関連するサブシステム境界も連動して表示/非表示

**レイアウト**: フルスクリーンキャンバス + オーバーレイパネル
- 左上: 要素一覧パネル（Actor/UseCase リスト、検索・フィルタ機能）
- 右上: 詳細パネル（選択時のみ表示）
- 中央: PixiJS キャンバス（サブシステム境界付き）

**Svelte コンポーネント**:
| コンポーネント | 役割 |
|---------------|------|
| UseCaseView.svelte | メインビュー、エンジン管理 |
| UseCaseListPanel.svelte | Actor/UseCase リスト表示 |
| UseCaseViewPanel.svelte | 選択エンティティの詳細表示 |
| SearchInput.svelte | 検索入力 |
| FilterDropdown.svelte | ステータスフィルタ |
| GroupedList.svelte | グループ化リスト |
| SegmentedTabs.svelte | Actor/UseCase タブ切替 |

**PixiJS レンダリングクラス**:
| クラス | 役割 |
|--------|------|
| UseCaseEngine | エンジン管理、レイアウト計算、インタラクション |
| ActorNode | アクター描画（UML スティックマン） |
| UseCaseNode | ユースケース描画（楕円） |
| SystemBoundary | システム境界描画 |
| SubsystemBoundary | サブシステム境界描画（カラー付き） |
| RelationEdge | include/extend/generalize 関係線 |
| ActorUseCaseEdge | アクター↔ユースケース接続線 |

**インタラクション**:
| 操作 | 結果 |
|------|------|
| クリック（Actor/UseCase） | 詳細パネル表示 + 関連エンティティ表示 |
| ホバー | ツールチップ表示 + 関連エッジハイライト |
| ドラッグ | キャンバスパン |
| ホイール | ズームイン/アウト |
| ESC | パネルを閉じる |

**API 連携**:
- `/api/actors` - Actor 一覧取得
- `/api/usecases` - UseCase 一覧取得
- `/api/subsystems` - Subsystem 一覧取得（並列取得）
- `/api/uml/usecase` - ユースケース図データ取得

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

**レイアウト**: Sugiyama 風階層レイアウト
- トポロジカルソートでノードをレベル分け
- 各レベル内で横方向に配置
- 遷移エッジは直線または折れ線で描画

**Svelte コンポーネント**:
| コンポーネント | 役割 |
|---------------|------|
| ActivityView.svelte | メインビュー、エンジン管理 |
| ActivityListPanel.svelte | アクティビティ一覧表示 |
| ActivityDetailPanel.svelte | 選択ノードの詳細表示 |

**PixiJS レンダリングクラス**:
| クラス | 役割 |
|--------|------|
| ActivityEngine | エンジン管理、Sugiyama レイアウト、インタラクション |
| ActivityNodeBase | ノード基底クラス（共通インターフェース） |
| InitialNode | 開始ノード（黒丸） |
| FinalNode | 終了ノード（二重丸） |
| ActionNode | アクションノード（角丸四角形） |
| DecisionNode | 分岐ノード（ひし形） |
| MergeNode | 合流ノード（ひし形） |
| ForkNode | 並列分岐ノード（太い横線） |
| JoinNode | 並列合流ノード（太い横線） |
| TransitionEdge | 遷移エッジ（矢印付き） |

**オーバーレイパネル構成**:
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
