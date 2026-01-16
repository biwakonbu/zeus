# zeus init 簡素化 - 設計書

## アーキテクチャ変更

### コンポーネント図

```
┌─────────────────────────────────────────────────────────┐
│                     cmd/init.go                         │
│  - --level フラグ削除                                   │
│  - runInit() 簡素化                                     │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                internal/core/zeus.go                    │
│  - Init(ctx) シグネチャ変更                             │
│  - getDirectoryStructure() 統一                         │
│  - generateInitialConfig() デフォルト auto              │
└──────────────────────┬──────────────────────────────────┘
                       │
           ┌───────────┴───────────┐
           ▼                       ▼
┌──────────────────────┐  ┌──────────────────────────────┐
│ internal/core/       │  │ internal/generator/          │
│ types.go             │  │ generator.go                 │
│ - InitResult 更新    │  │ - GenerateAll() level削除    │
└──────────────────────┘  └──────────────────────────────┘
```

## 詳細設計

### cmd/init.go

```go
// 変更後
func init() {
    rootCmd.AddCommand(initCmd)
    // フラグなし
}

func runInit(cmd *cobra.Command, args []string) error {
    ctx := getContext(cmd)

    zeus := getZeus(cmd)
    result, err := zeus.Init(ctx)
    // ...
    // Level 出力を削除
}
```

### internal/core/zeus.go

```go
// 変更後
func (z *Zeus) Init(ctx context.Context) (*InitResult, error) {
    // ...
    dirs := z.getDirectoryStructure()
    // ...
    // 常に Claude Code 連携ファイルを生成
    gen := generator.NewGenerator(z.ProjectPath)
    gen.GenerateAll(ctx, config.Project.Name)

    return &InitResult{
        // Level フィールドなし
        // ...
    }
}

func (z *Zeus) getDirectoryStructure() []string {
    // standard 相当の統一構造
    return []string{
        "config", "tasks", "tasks/_archive", "state", "state/snapshots",
        "entities", "approvals/pending", "approvals/approved", "approvals/rejected",
        "logs", "analytics", "backups",
    }
}

func (z *Zeus) generateInitialConfig() *ZeusConfig {
    return &ZeusConfig{
        Settings: Settings{
            AutomationLevel: "auto",  // デフォルトを auto に変更
            // ...
        },
    }
}
```

### internal/core/types.go

```go
// 変更後
type InitResult struct {
    Success    bool
    ZeusPath   string
    ClaudePath string
}
```

### internal/generator/generator.go

```go
// 変更後
func (g *Generator) GenerateAll(ctx context.Context, projectName string) error
```

## データモデル変更

### zeus.yaml デフォルト設定

```yaml
version: "1.0"
project:
  id: "zeus-{timestamp}"
  name: "New Zeus Project"
  description: "Project managed by Zeus"
  start_date: "{today}"
objectives: []
settings:
  automation_level: "auto"    # 変更: standard -> auto
  approval_mode: "default"
  ai_provider: "claude-code"
```

## 影響を受けるテスト

### 修正対象テスト一覧

| テスト関数 | 修正内容 |
|-----------|----------|
| TestZeusIntegration | Init(ctx) に変更 |
| TestZeusSnapshot | Init(ctx) に変更 |
| TestPending | Init(ctx) に変更 |
| TestApprove | Init(ctx) に変更 |
| TestReject | Init(ctx) に変更 |
| TestRestoreSnapshotIntegration | Init(ctx) に変更 |
| TestInit_AllLevels | 削除（単一レベルのため不要） |
| TestGetDirectoryStructure | 単一ケースに変更 |
| TestInitContextTimeout | Init(ctx) に変更 |

### 修正パターン

```go
// 変更前
_, err = z.Init(ctx, "simple")
_, err = z.Init(ctx, "standard")
_, err = z.Init(ctx, "advanced")

// 変更後
_, err = z.Init(ctx)
```

## 移行計画

| Step | 対象 | 内容 |
|------|------|------|
| 1 | internal/core/types.go | InitResult から Level フィールドを削除 |
| 2 | internal/generator/generator.go | GenerateAll() から level パラメータを削除 |
| 3 | internal/core/zeus.go | Init() シグネチャ変更、getDirectoryStructure() 統一、generateInitialConfig() デフォルト変更、Claude Code 生成の条件分岐削除 |
| 4 | cmd/init.go | --level フラグ削除、runInit() 修正 |
| 5 | テスト | 全 Init 呼び出しを修正 |
| 6 | ドキュメント | USER_GUIDE.md, API_REFERENCE.md, SYSTEM_DESIGN.md, CLAUDE.md |

## ロールバック計画

変更はすべて Go コードとドキュメントのみ。Git revert で即座にロールバック可能。
