# E2E テスト仕様書

## 概要

Zeus CLI のエンドツーエンド（E2E）テストは、実際の CLI バイナリを実行し、統合的な動作を検証します。

## ディレクトリ構成

```
tests/e2e/
├── e2e_test.go        # TestMain（バイナリビルド管理）
├── helpers.go         # ヘルパー関数
└── scenarios_test.go  # シナリオ別テスト
```

## 実行方法

```bash
# 全 E2E テスト実行
go test -v ./tests/e2e/...

# 特定テストのみ実行
go test -v ./tests/e2e/... -run TestBasicFlow

# 並行実行数を制限
go test -v ./tests/e2e/... -parallel 4
```

## テストカテゴリ

### 基本フロー

| テスト名 | 検証内容 |
|---------|---------|
| TestBasicFlow | init → status → doctor の基本フロー |
| TestInitLevels | simple/standard/advanced 各レベルの初期化 |

### タスク管理

| テスト名 | 検証内容 |
|---------|---------|
| TestTaskManagementFlow | add → list の基本フロー |
| TestAddMultipleTasks | 複数タスクの追加 |

### 承認フロー

| テスト名 | 検証内容 |
|---------|---------|
| TestApprovalFlowAdvanced | advanced レベルの承認フロー |
| TestApprovalFlowStandard | standard レベルの承認フロー |
| TestApproveReject | approve/reject コマンド |

### スナップショット

| テスト名 | 検証内容 |
|---------|---------|
| TestSnapshotFlow | snapshot create → list |
| TestSnapshotRestore | snapshot restore |

### 分析

| テスト名 | 検証内容 |
|---------|---------|
| TestAnalysisFlow | graph → predict → report |
| TestGraphFormats | text/dot/mermaid 形式 |
| TestPredictTypes | completion/risk/velocity/all |
| TestReportFormats | text/html/markdown 形式 |

### エラーケース

| テスト名 | 検証内容 |
|---------|---------|
| TestUninitializedProjectXxx | 未初期化プロジェクトでのエラー |
| TestInvalidArguments | 不正な引数 |
| TestXxxNonexistent | 存在しないエンティティ |
| TestInvalidXxxFormat | 不正なフォーマット指定 |

## ヘルパー関数

### runCommand

コマンドを実行して結果を返す。30秒タイムアウト。

```go
result := runCommand(t, dir, "init", "--level=simple")
```

### setupTempDir / cleanupTempDir

テスト用一時ディレクトリの作成・削除。

```go
dir := setupTempDir(t)
defer cleanupTempDir(t, dir)
```

### assertXxx

アサーション関数群。

```go
assertSuccess(t, result)
assertFailure(t, result)
assertOutputContains(t, result, "expected")
assertDirExists(t, path)
assertFileExists(t, path)
```

## 設計原則

1. **テスト独立性**: 各テストは他のテストに依存しない
2. **並行実行**: `t.Parallel()` で並行実行可能
3. **クリーンアップ保証**: defer でクリーンアップを保証
4. **タイムアウト**: 30秒タイムアウトでハングアップ防止

## 拡張ガイド

### 新規テストの追加

1. `scenarios_test.go` に新規テスト関数を追加
2. `t.Parallel()` を呼び出して並行実行可能に
3. `setupTempDir` / `cleanupTempDir` でディレクトリ管理
4. `assertXxx` 関数でアサーション

```go
func TestNewFeature(t *testing.T) {
    t.Parallel()
    dir := setupTempDir(t)
    defer cleanupTempDir(t, dir)

    // テスト実装
    result := runCommand(t, dir, "init", "--level=simple")
    assertSuccess(t, result)
}
```

### 新規ヘルパーの追加

`helpers.go` に追加。`t.Helper()` を呼び出してスタックトレースを改善。

```go
func assertNewCondition(t *testing.T, ...) {
    t.Helper()
    // 検証ロジック
}
```
