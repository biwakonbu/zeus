---
description: Zeus プロジェクトの統合 E2E テストスキル（CLI + Web）。State-First アプローチで PixiJS Canvas 内部状態を検証。
use_when: |
  Use when user wants to run E2E tests for Zeus project.
  Also use when user says "E2E テスト", "統合テスト", "e2e test", "integration test".
skills:
  - zeus-e2e-tester
model: sonnet
---

# zeus-e2e-tester

Zeus プロジェクトの E2E テストを実行するスキル。CLI コマンドと Web ダッシュボードを複合的にテストし、ゴールデンファイルと比較する。

## 設計哲学

### State-First & Deterministic Export

従来の Playwright ピクセル比較ではなく、**アプリケーション内部状態の直接検証**を採用。

- **理由**: Canvas（PixiJS）はアクセシビリティツリーに現れず、agent-browser の snapshot では検証不可
- **解決策**: `window.__ZEUS__` グローバル API でシーングラフにアクセスし、JSON 形式で状態をエクスポート
- **利点**: OS/ブラウザ依存性排除、AI API コスト 90% 削減、Git フレンドリー

### 検証の 2 層構造

1. **状態検証（必須）**: JSON ハッシュ比較
2. **視覚確認（補完）**: 低解像度スナップショット（致命的エラー検知用）

## 実行モード

| オプション | 説明 |
|-----------|------|
| `--mode all` | 全テスト実行（デフォルト） |
| `--mode cli` | CLI テストのみ |
| `--mode web` | Web テストのみ |
| `--update-golden` | ゴールデンファイルを更新 |

## テストフロー

### Step 1: 環境準備

1. テスト用プロジェクトディレクトリを作成（`/tmp/zeus-e2e-test-XXXXXX`）
2. `zeus init` で初期化
3. サンプル Activity を追加

### Step 2: CLI テスト

1. 各コマンドを実行
2. 出力をキャプチャ
3. ゴールデンファイルと比較（正規表現マッチ or JSON 構造比較）

```bash
# テスト対象コマンド例
zeus init
zeus status
zeus add activity "Activity A"
zeus add activity "Activity B" --parent <act-a-id>
zeus list activities
zeus graph --format mermaid
```

### Step 3: Web テスト

1. `zeus dashboard --dev` を起動（ポート 8080）
2. agent-browser（chrome-automation）でページを開く
3. `window.__ZEUS__.isReady()` で描画完了を待機
4. `window.__ZEUS__.getGraphState()` で状態取得
5. ゴールデンファイルと比較
6. スナップショット取得（視覚確認用）

### Step 4: 結果レポート

1. PASS/FAIL をまとめ
2. 差分があれば詳細表示

## window.__ZEUS__ API

開発/テスト環境でのみ公開されるグローバル API。

```typescript
window.__ZEUS__ = {
  // グラフの論理構造を返す
  getGraphState: () => ({
    nodes: [{ id, name, x, y, status, progress }],
    edges: [{ from, to }],
    viewport: { zoom, panX, panY },
    activityCount: number,
    edgeCount: number
  }),

  // 選択状態を返す
  getSelectionState: () => ({
    selectedIds: string[],
    count: number,
    multiSelect: boolean
  }),

  // フィルター状態を返す
  getFilterState: () => ({
    criteria: object,
    visibleCount: number,
    totalCount: number
  }),

  // 描画完了を待機
  isReady: () => boolean,

  // バージョン情報
  getVersion: () => string
};
```

## ゴールデンファイル形式

```json
{
  "metadata": {
    "test_id": "graph-state-001",
    "generated_at": "2026-01-17T04:00:00Z",
    "zeus_version": "0.1.0",
    "hash": "sha256:abc123..."
  },
  "state": {
    "nodes": [
      {
        "id": "act-001",
        "name": "Activity A",
        "x": 100,
        "y": 200,
        "status": "draft",
        "progress": 0
      }
    ],
    "edges": [
      { "from": "act-001", "to": "act-002" }
    ],
    "viewport": {
      "zoom": 1.0,
      "panX": 0,
      "panY": 0
    }
  },
  "tolerance": {
    "position": 5,
    "zoom": 0.01
  }
}
```

## テスト実行手順

### 1. 環境セットアップ

```bash
# プロジェクトルートで実行
cd /Users/biwakonbu/github/zeus

# ビルド
make build

# ダッシュボード依存関係（初回のみ）
make dashboard-deps
```

### 2. スキル実行

```bash
# Claude Code でスキルを呼び出す
/zeus-e2e-tester

# または引数付き
/zeus-e2e-tester --mode cli
/zeus-e2e-tester --update-golden
```

### 3. 期待される出力

```
Zeus E2E Test Suite
===================

CLI Tests:
  ✓ zeus init                    [PASS]
  ✓ zeus add activity            [PASS]
  ✓ zeus status                  [PASS]

Web Tests:
  ✓ Dashboard connection         [PASS]
  ✓ Graph state verification     [PASS]
  ✓ Activity selection           [PASS]

Integration Tests:
  ✓ CLI → Web state sync         [PASS]

Summary: 7/7 tests passed
```

## シナリオファイル

シナリオは `resources/scenarios/` に YAML 形式で定義。

### ステップタイプ

| type | 説明 |
|------|------|
| `cli` | CLI コマンド実行 |
| `web` | Web ブラウザ操作 |
| `wait` | 条件待機 |
| `assert` | アサーション |

### CLI ステップ

```yaml
- type: cli
  command: zeus init
  capture: init-output       # 変数にキャプチャ
  expect_exit_code: 0        # 期待する終了コード
  golden: golden/cli-init.golden.json  # ゴールデン比較
```

### Web ステップ

```yaml
- type: web
  action: navigate
  url: http://localhost:8080

- type: web
  action: wait
  condition: window.__ZEUS__.isReady()
  timeout: 10000

- type: web
  action: capture_state
  method: window.__ZEUS__.getGraphState()
  golden: golden/graph-state.golden.json
  tolerance:
    position: 5

- type: web
  action: click
  selector: "[data-testid='activity-node-001']"

- type: web
  action: snapshot
  path: golden/snapshots/dashboard-loaded.png
  mode: reference  # 参照用、厳密比較しない
```

## 注意事項

### agent-browser の制限

- Canvas 内部要素は DOM セレクタで選択不可
- スナップショットはピクセル比較ではなく参照用途
- `window.__ZEUS__` API 経由で状態を取得

### 環境依存性の排除

- 座標値は tolerance 付きで比較
- タイムスタンプは除外
- UUID は形式のみ検証

### テスト環境の分離

- テスト用ディレクトリは `/tmp` に作成
- 終了時にクリーンアップ
- 既存のプロジェクトには影響しない

## リソースファイル

| ファイル | 説明 |
|---------|------|
| `resources/scenarios/cli-basic.yaml` | CLI 基本操作シナリオ |
| `resources/scenarios/web-interaction.yaml` | Web インタラクションシナリオ |
| `resources/scenarios/integration.yaml` | 統合シナリオ |
| `resources/golden/README.md` | ゴールデンファイル管理ガイド |
| `resources/golden/*.golden.json` | ゴールデンファイル |
| `resources/golden/snapshots/` | 参照スクリーンショット |
