# Zeus 実装ガイド（Go版）

## 1. はじめに

このガイドでは、Zeus システムを Go + Cobra で実装する手順を説明します。

## 2. 技術スタック

| 項目 | 選定 |
|------|------|
| 実装言語 | Go 1.21+ |
| CLIフレームワーク | Cobra |
| データ形式 | YAML（gopkg.in/yaml.v3） |
| AI接続 | Claude Code 経由 |
| 配布形式 | スタンドアロン CLI + Claude Code Plugin |

## 3. プロジェクト構造

### 3.1 Zeus CLI ソースコード構造

```
zeus/
├── cmd/                      # Cobra コマンド
│   ├── root.go               # ルートコマンド
│   ├── init.go               # zeus init
│   ├── status.go             # zeus status
│   ├── add.go                # zeus add
│   ├── list.go               # zeus list
│   ├── suggest.go            # zeus suggest
│   ├── apply.go              # zeus apply
│   ├── approve.go            # zeus approve/reject
│   ├── pending.go            # zeus pending
│   ├── doctor.go             # zeus doctor
│   └── fix.go                # zeus fix
│
├── internal/                 # 内部パッケージ
│   ├── core/                 # コア機能
│   │   ├── zeus.go           # メインロジック
│   │   ├── state.go          # 状態管理
│   │   └── approval.go       # 承認管理
│   ├── yaml/                 # YAML操作
│   │   ├── parser.go
│   │   └── writer.go
│   ├── doctor/               # 診断・修復
│   │   └── doctor.go
│   └── generator/            # Claude Code 連携ファイル生成
│       ├── agents.go         # agent テンプレート生成
│       └── skills.go         # skill テンプレート生成
│
├── templates/                # 埋め込みテンプレート（embed）
│   ├── zeus.yaml.tmpl
│   ├── task.yaml.tmpl
│   ├── state.yaml.tmpl
│   ├── agents/               # agent テンプレート
│   │   ├── zeus-orchestrator.md.tmpl
│   │   ├── zeus-planner.md.tmpl
│   │   └── zeus-reviewer.md.tmpl
│   └── skills/               # skill テンプレート
│       ├── project-scan.md.tmpl
│       ├── task-suggest.md.tmpl
│       └── risk-analysis.md.tmpl
│
├── docs/                     # ドキュメント
│   ├── SYSTEM_DESIGN.md
│   ├── IMPLEMENTATION_GUIDE.md
│   └── OPERATIONS_MANUAL.md
│
├── go.mod
├── go.sum
├── main.go
├── Makefile
└── README.md
```

### 3.2 zeus init 実行後のターゲットリポジトリ構造

```
target-project/               # Zeus を適用するプロジェクト
├── .zeus/                    # Zeus プロジェクト管理
│   ├── zeus.yaml             # メイン設定
│   ├── tasks/                # タスク管理
│   │   ├── active.yaml
│   │   ├── backlog.yaml
│   │   └── _archive/
│   ├── state/                # 状態管理
│   │   ├── current.yaml
│   │   └── snapshots/
│   └── backups/              # 自動バックアップ
│
├── .claude/                  # Claude Code 標準構造
│   ├── agents/               # Zeus 用エージェント
│   │   ├── zeus-orchestrator.md
│   │   ├── zeus-planner.md
│   │   └── zeus-reviewer.md
│   └── skills/               # Zeus 用スキル
│       ├── zeus-project-scan/
│       │   └── SKILL.md
│       ├── zeus-task-suggest/
│       │   └── SKILL.md
│       └── zeus-risk-analysis/
│           └── SKILL.md
│
└── ... (既存のプロジェクトファイル)
```

## 4. 依存パッケージ

### 4.1 go.mod

```go
module github.com/biwakonbu/zeus

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    gopkg.in/yaml.v3 v3.0.1
    github.com/fatih/color v1.16.0
)
```

## 5. コア実装

### 5.1 main.go

```go
package main

import (
    "os"

    "github.com/biwakonbu/zeus/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### 5.2 cmd/root.go

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "zeus",
    Short: "AI-driven project management with god's eye view",
    Long: `Zeus は AI によるプロジェクトマネジメントを「神の視点」で
俯瞰するシステムです。上流工程（方針立案からWBS化、タイムライン設計、
仕様作成まで）を支援します。`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // グローバルフラグ
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "詳細出力")
}
```

### 5.3 cmd/init.go

```go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/biwakonbu/zeus/internal/core"
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Zeus プロジェクトを初期化",
    Long:  `プロジェクトディレクトリに .zeus/ と .claude/ を生成します。`,
    RunE: func(cmd *cobra.Command, args []string) error {
        level, _ := cmd.Flags().GetString("level")

        zeus := core.New(".")
        result, err := zeus.Init(level)
        if err != nil {
            return err
        }

        fmt.Printf("Zeus initialized successfully!\n")
        fmt.Printf("  Level: %s\n", result.Level)
        fmt.Printf("  Path: %s\n", result.ZeusPath)
        fmt.Printf("  Claude integration: %s\n", result.ClaudePath)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
    initCmd.Flags().StringP("level", "l", "simple", "初期化レベル (simple|standard|advanced)")
}
```

### 5.4 internal/core/zeus.go

```go
package core

import (
    "embed"
    "os"
    "path/filepath"
    "time"

    "github.com/biwakonbu/zeus/internal/generator"
    "github.com/biwakonbu/zeus/internal/yaml"
)

//go:embed templates/*
var templates embed.FS

// Zeus はメインのアプリケーション構造体
type Zeus struct {
    ProjectPath string
    ZeusPath    string
    ClaudePath  string
    State       *StateManager
    Approval    *ApprovalManager
}

// InitResult は初期化結果
type InitResult struct {
    Success    bool
    Level      string
    ZeusPath   string
    ClaudePath string
}

// New は新しい Zeus インスタンスを作成
func New(projectPath string) *Zeus {
    return &Zeus{
        ProjectPath: projectPath,
        ZeusPath:    filepath.Join(projectPath, ".zeus"),
        ClaudePath:  filepath.Join(projectPath, ".claude"),
    }
}

// Init はプロジェクトを初期化
func (z *Zeus) Init(level string) (*InitResult, error) {
    // .zeus ディレクトリ構造を作成
    if err := z.createZeusStructure(level); err != nil {
        return nil, err
    }

    // zeus.yaml を生成
    config := z.generateInitialConfig()
    if err := yaml.WriteFile(filepath.Join(z.ZeusPath, "zeus.yaml"), config); err != nil {
        return nil, err
    }

    // .claude ディレクトリを作成（agents, skills）
    if err := generator.GenerateClaudeIntegration(z.ClaudePath); err != nil {
        return nil, err
    }

    return &InitResult{
        Success:    true,
        Level:      level,
        ZeusPath:   z.ZeusPath,
        ClaudePath: z.ClaudePath,
    }, nil
}

func (z *Zeus) createZeusStructure(level string) error {
    dirs := z.getDirectoryStructure(level)
    for _, dir := range dirs {
        path := filepath.Join(z.ZeusPath, dir)
        if err := os.MkdirAll(path, 0755); err != nil {
            return err
        }
    }
    return nil
}

func (z *Zeus) getDirectoryStructure(level string) []string {
    switch level {
    case "simple":
        return []string{".", "tasks", "state", "backups"}
    case "standard":
        return []string{
            ".", "config", "tasks", "tasks/_archive",
            "state", "state/snapshots",
            "entities", "approvals/pending", "approvals/approved", "approvals/rejected",
            "logs", "analytics", "backups",
        }
    case "advanced":
        return []string{
            ".", "config", "tasks", "tasks/_archive",
            "state", "state/snapshots",
            "entities", "approvals/pending", "approvals/approved", "approvals/rejected",
            "logs", "logs/ai-actions", "logs/decisions",
            "analytics", "graph", "graph/computed",
            "views", "views/templates", "views/generated",
            "backups", ".local",
        }
    default:
        return []string{".", "tasks", "state", "backups"}
    }
}

func (z *Zeus) generateInitialConfig() map[string]interface{} {
    return map[string]interface{}{
        "version": "1.0",
        "project": map[string]interface{}{
            "id":          fmt.Sprintf("zeus-%d", time.Now().Unix()),
            "name":        "New Zeus Project",
            "description": "Project managed by Zeus",
            "start_date":  time.Now().Format("2006-01-02"),
        },
        "objectives": []interface{}{},
        "settings": map[string]interface{}{
            "automation_level": "standard",
            "approval_mode":    "default",
            "ai_provider":      "claude-code",
        },
    }
}
```

### 5.5 internal/generator/agents.go

```go
package generator

import (
    "os"
    "path/filepath"
)

// GenerateClaudeIntegration は .claude/ ディレクトリを生成
func GenerateClaudeIntegration(claudePath string) error {
    // agents ディレクトリ
    agentsPath := filepath.Join(claudePath, "agents")
    if err := os.MkdirAll(agentsPath, 0755); err != nil {
        return err
    }

    // skills ディレクトリ
    skillsPath := filepath.Join(claudePath, "skills")
    if err := os.MkdirAll(skillsPath, 0755); err != nil {
        return err
    }

    // エージェントファイルを生成
    agents := map[string]string{
        "zeus-orchestrator.md": zeusOrchestratorTemplate,
        "zeus-planner.md":      zeusPlannerTemplate,
        "zeus-reviewer.md":     zeusReviewerTemplate,
    }
    for name, content := range agents {
        path := filepath.Join(agentsPath, name)
        if err := os.WriteFile(path, []byte(content), 0644); err != nil {
            return err
        }
    }

    // スキルディレクトリを生成
    skills := []struct {
        name    string
        content string
    }{
        {"zeus-project-scan", zeusProjectScanSkill},
        {"zeus-task-suggest", zeusTaskSuggestSkill},
        {"zeus-risk-analysis", zeusRiskAnalysisSkill},
    }
    for _, skill := range skills {
        skillDir := filepath.Join(skillsPath, skill.name)
        if err := os.MkdirAll(skillDir, 0755); err != nil {
            return err
        }
        skillFile := filepath.Join(skillDir, "SKILL.md")
        if err := os.WriteFile(skillFile, []byte(skill.content), 0644); err != nil {
            return err
        }
    }

    return nil
}

const zeusOrchestratorTemplate = `---
description: Zeus プロジェクト管理を統括するオーケストレーター
tools: [Bash, Read, Write, Glob, Grep]
model: sonnet
---

# Zeus Orchestrator

Zeus CLI を使用してプロジェクト管理を行うエージェントです。

## 役割

1. プロジェクト全体の状況把握
2. AI 提案の生成と管理
3. 承認フローの調整
4. レポート生成

## 使用するコマンド

- zeus status: プロジェクト状況確認
- zeus suggest: AI 提案生成
- zeus pending: 承認待ち確認
- zeus report: レポート生成
`

const zeusPlannerTemplate = `---
description: タスク計画とスケジューリングを担当
tools: [Bash, Read, Write, Glob, Grep]
model: sonnet
---

# Zeus Planner

プロジェクトのタスク計画を担当するエージェントです。

## 役割

1. タスクブレークダウン
2. 優先順位付け
3. 依存関係の特定
4. タイムライン最適化
`

const zeusReviewerTemplate = `---
description: 進捗レビューと品質確認を担当
tools: [Bash, Read, Glob, Grep]
model: sonnet
---

# Zeus Reviewer

プロジェクトの進捗レビューを担当するエージェントです。

## 役割

1. 進捗状況のレビュー
2. リスク評価
3. 品質チェック
4. 改善提案
`

const zeusProjectScanSkill = `---
description: プロジェクト構造をスキャンして分析
---

# Project Scan スキル

プロジェクトの構造を分析し、Zeus 管理に必要な情報を収集します。

## 実行内容

1. プロジェクトファイルの探索
2. 技術スタックの特定
3. 既存タスク管理ツールの検出
4. zeus.yaml の初期構造を提案
`

const zeusTaskSuggestSkill = `---
description: タスク分割と優先順位を提案
---

# Task Suggest スキル

プロジェクトの目標からタスクを分割し、優先順位を提案します。

## 実行内容

1. 目標の分析
2. タスクブレークダウン
3. 優先順位付け
4. 依存関係の特定
`

const zeusRiskAnalysisSkill = `---
description: プロジェクトリスクを分析して対策を提案
---

# Risk Analysis スキル

プロジェクトのリスクを分析し、対策を提案します。

## 実行内容

1. リスク要因の特定
2. 影響度の評価
3. 発生確率の推定
4. 対策の提案
`
```

### 5.6 internal/yaml/parser.go

```go
package yaml

import (
    "os"

    "gopkg.in/yaml.v3"
)

// ReadFile は YAML ファイルを読み込む
func ReadFile(path string, v interface{}) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    return yaml.Unmarshal(data, v)
}

// WriteFile は YAML ファイルを書き込む
func WriteFile(path string, v interface{}) error {
    data, err := yaml.Marshal(v)
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}
```

## 6. ビルドと実行

### 6.1 Makefile

```makefile
.PHONY: build clean test install

BINARY_NAME=zeus
VERSION=1.0.0

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

# 開発用
dev:
	go run . $(ARGS)
```

### 6.2 ビルド確認

```bash
# ビルド
make build

# 実行確認
./zeus --help
./zeus init
./zeus status
```

## 7. 実装優先順位

### Phase 1: MVP（最小実行可能プロダクト）

| コマンド | 優先度 | 説明 |
|---------|--------|------|
| `zeus init` | 高 | プロジェクト初期化（.zeus/ + .claude/ 生成） |
| `zeus status` | 高 | 状態表示 |
| `zeus add task` | 高 | タスク追加 |
| `zeus list` | 高 | 一覧表示 |
| `zeus doctor` | 中 | 診断 |
| `zeus fix` | 中 | 修復 |

### Phase 2: AI 統合

| コマンド | 説明 |
|---------|------|
| `zeus suggest` | AI 提案 |
| `zeus apply` | 提案適用 |
| `zeus explain` | AI 解説 |

### Phase 3: 承認フロー

| コマンド | 説明 |
|---------|------|
| `zeus pending` | 承認待ち一覧 |
| `zeus approve` | 承認 |
| `zeus reject` | 却下 |

## 8. テスト

### 8.1 ユニットテスト例

```go
// internal/core/zeus_test.go
package core

import (
    "os"
    "path/filepath"
    "testing"
)

func TestZeusInit(t *testing.T) {
    // 一時ディレクトリを作成
    tmpDir, err := os.MkdirTemp("", "zeus-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // Zeus インスタンスを作成
    zeus := New(tmpDir)

    // 初期化を実行
    result, err := zeus.Init("simple")
    if err != nil {
        t.Fatalf("Init failed: %v", err)
    }

    // 結果を検証
    if !result.Success {
        t.Error("Expected success to be true")
    }
    if result.Level != "simple" {
        t.Errorf("Expected level 'simple', got '%s'", result.Level)
    }

    // ファイルが作成されたか確認
    zeusYaml := filepath.Join(tmpDir, ".zeus", "zeus.yaml")
    if _, err := os.Stat(zeusYaml); os.IsNotExist(err) {
        t.Error("zeus.yaml was not created")
    }
}
```

---

*Zeus Implementation Guide (Go版) v1.0*
*作成日: 2026-01-14*
