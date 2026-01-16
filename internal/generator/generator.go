package generator

import (
	"context"
	"os"
	"path/filepath"
	"text/template"

	"github.com/biwakonbu/zeus/internal/yaml"
)

// Generator は Claude Code 連携ファイルを生成
type Generator struct {
	projectPath string
	claudePath  string
	fileManager *yaml.FileManager
}

// NewGenerator は新しい Generator を作成
func NewGenerator(projectPath string) *Generator {
	claudePath := filepath.Join(projectPath, ".claude")
	return &Generator{
		projectPath: projectPath,
		claudePath:  claudePath,
		fileManager: yaml.NewFileManager(claudePath),
	}
}

// GenerateAll は全ての Claude Code 連携ファイルを生成（Context対応）
func (g *Generator) GenerateAll(ctx context.Context, projectName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := g.GenerateAgents(ctx, projectName); err != nil {
		return err
	}
	if err := g.GenerateSkills(ctx, projectName); err != nil {
		return err
	}
	return nil
}

// GenerateAgents はエージェントファイルを生成（Context対応）
func (g *Generator) GenerateAgents(ctx context.Context, projectName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := g.fileManager.EnsureDir(ctx, "agents"); err != nil {
		return err
	}

	agents := []struct {
		name     string
		template string
	}{
		{"zeus-orchestrator.md", orchestratorTemplate},
		{"zeus-planner.md", plannerTemplate},
		{"zeus-reviewer.md", reviewerTemplate},
	}

	data := map[string]string{
		"ProjectName": projectName,
	}

	for _, agent := range agents {
		if err := ctx.Err(); err != nil {
			return err
		}

		content, err := g.executeTemplate(agent.template, data)
		if err != nil {
			return err
		}
		if err := g.fileManager.WriteFile(ctx, filepath.Join("agents", agent.name), []byte(content)); err != nil {
			return err
		}
	}

	return nil
}

// GenerateSkills はスキルファイルを生成（Context対応）
func (g *Generator) GenerateSkills(ctx context.Context, projectName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	skills := []struct {
		dir      string
		template string
	}{
		{"zeus-project-scan", projectScanTemplate},
		{"zeus-task-suggest", taskSuggestTemplate},
		{"zeus-risk-analysis", riskAnalysisTemplate},
	}

	data := map[string]string{
		"ProjectName": projectName,
	}

	for _, skill := range skills {
		if err := ctx.Err(); err != nil {
			return err
		}

		skillDir := filepath.Join("skills", skill.dir)
		if err := g.fileManager.EnsureDir(ctx, skillDir); err != nil {
			return err
		}

		content, err := g.executeTemplate(skill.template, data)
		if err != nil {
			return err
		}
		if err := g.fileManager.WriteFile(ctx, filepath.Join(skillDir, "SKILL.md"), []byte(content)); err != nil {
			return err
		}
	}

	return nil
}

// executeTemplate はテンプレートを実行
func (g *Generator) executeTemplate(tmplContent string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf []byte
	writer := &byteWriter{buf: &buf}
	if err := tmpl.Execute(writer, data); err != nil {
		return "", err
	}

	return string(buf), nil
}

// byteWriter は []byte への Writer
type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// EnsureClaudeDir は .claude ディレクトリを確認・作成
func (g *Generator) EnsureClaudeDir() error {
	return os.MkdirAll(g.claudePath, 0755)
}

// Agent Templates

const orchestratorTemplate = `---
description: Zeus プロジェクト管理を統括するオーケストレーター
tools: [Bash, Read, Write, Glob, Grep]
model: sonnet
---

# Zeus Orchestrator Agent

このエージェントは Zeus プロジェクト（{{.ProjectName}}）のオーケストレーターとして機能します。

## 役割

1. **プロジェクト全体の把握**: タスク、目標、リソースを俯瞰
2. **優先順位付け**: 重要度と緊急度に基づいた判断
3. **リスク検知**: 潜在的な問題を早期発見
4. **進捗管理**: 全体の進捗状況を追跡

## コマンド

- ` + "`zeus status`" + ` - 現在の状態を確認
- ` + "`zeus list tasks`" + ` - タスク一覧を表示
- ` + "`zeus doctor`" + ` - システム診断

## 判断基準

1. **迷ったら人間に聞く**: 確信がない判断は保留
2. **安全第一**: リスクのある変更は承認を求める
3. **透明性**: 全ての判断理由を記録

## 使用スキル

- @zeus-project-scan - プロジェクトスキャン
- @zeus-task-suggest - タスク提案
- @zeus-risk-analysis - リスク分析
`

const plannerTemplate = `---
description: Zeus プロジェクトの計画立案エージェント
tools: [Bash, Read, Write, Glob]
model: sonnet
---

# Zeus Planner Agent

このエージェントは Zeus プロジェクト（{{.ProjectName}}）の計画立案を担当します。

## 役割

1. **WBS 作成**: タスクの分解と構造化
2. **見積もり**: 工数の見積もり
3. **依存関係分析**: タスク間の依存関係を特定
4. **タイムライン設計**: スケジュールの策定

## コマンド

- ` + "`zeus add task <name>`" + ` - タスクを追加
- ` + "`zeus list tasks --status pending`" + ` - 未着手タスクを表示

## 計画の原則

1. **保守的な見積もり**: バッファを確保
2. **段階的計画**: 大きなタスクは分割
3. **柔軟性**: 変更に対応できる余地を残す

## 出力形式

` + "```yaml" + `
tasks:
  - id: task-1
    title: "タスク名"
    estimate_hours: 8
    dependencies: []
` + "```" + `
`

const reviewerTemplate = `---
description: Zeus プロジェクトのレビュー・品質管理エージェント
tools: [Bash, Read, Glob, Grep]
model: sonnet
---

# Zeus Reviewer Agent

このエージェントは Zeus プロジェクト（{{.ProjectName}}）のレビューを担当します。

## 役割

1. **進捗レビュー**: タスクの進捗を評価
2. **品質チェック**: 成果物の品質を確認
3. **リスク評価**: 潜在的な問題を評価
4. **改善提案**: 改善点を提案

## コマンド

- ` + "`zeus status`" + ` - 状態を確認
- ` + "`zeus pending`" + ` - 承認待ちアイテムを確認
- ` + "`zeus approve <id>`" + ` - アイテムを承認

## レビュー基準

1. **完了の定義**: 明確な完了条件を確認
2. **品質基準**: 定義された品質基準を満たしているか
3. **依存関係**: 後続タスクへの影響

## 承認レベル

- **auto**: 自動承認（低リスク操作）
- **notify**: 通知のみ（中リスク操作）
- **approve**: 明示的承認が必要（高リスク操作）
`

// Skill Templates

const projectScanTemplate = `---
description: プロジェクト全体をスキャンし、現在の状態を分析するスキル
---

# zeus-project-scan

プロジェクト全体をスキャンし、現在の状態を分析するスキル。

## 概要

Zeus プロジェクト（{{.ProjectName}}）の構造、タスク、進捗を分析します。

## 入力

なし（カレントディレクトリの .zeus/ を参照）

## 出力

` + "```yaml" + `
project:
  name: "プロジェクト名"
  health: "good|fair|poor"
  tasks:
    total: 10
    completed: 3
    in_progress: 2
    pending: 5
  risks: []
` + "```" + `

## 使用方法

1. ` + "`zeus status`" + ` コマンドを実行
2. 出力を解析
3. 改善提案を生成

## 関連

- zeus-task-suggest
- zeus-risk-analysis
`

const taskSuggestTemplate = `---
description: 現在の状態に基づいてタスクを提案するスキル
---

# zeus-task-suggest

現在の状態に基づいてタスクを提案するスキル。

## 概要

Zeus プロジェクト（{{.ProjectName}}）の状態を分析し、次に取り組むべきタスクを提案します。

## 入力

` + "```yaml" + `
context:
  current_tasks: []
  objectives: []
  blockers: []
` + "```" + `

## 出力

` + "```yaml" + `
suggestions:
  - id: suggestion-1
    type: "new_task|priority_change|dependency"
    description: "提案の説明"
    rationale: "理由"
    impact: "high|medium|low"
` + "```" + `

## アルゴリズム

1. 現在のタスク状態を分析
2. 目標との差分を計算
3. 優先度に基づいて提案を生成

## 承認

提案は ` + "`zeus pending`" + ` で確認し、` + "`zeus approve <id>`" + ` で適用します。
`

const riskAnalysisTemplate = `---
description: プロジェクトのリスクを分析し、対策を提案するスキル
---

# zeus-risk-analysis

プロジェクトのリスクを分析するスキル。

## 概要

Zeus プロジェクト（{{.ProjectName}}）のリスク要因を特定し、対策を提案します。

## 入力

` + "```yaml" + `
project_state:
  tasks: []
  timeline: {}
  resources: []
` + "```" + `

## 出力

` + "```yaml" + `
risks:
  - id: risk-1
    category: "schedule|resource|technical|external"
    description: "リスクの説明"
    probability: "high|medium|low"
    impact: "high|medium|low"
    mitigation: "対策"
` + "```" + `

## リスクカテゴリ

1. **スケジュールリスク**: 遅延、見積もり誤差
2. **リソースリスク**: 人員不足、スキル不足
3. **技術リスク**: 技術的困難、依存関係
4. **外部リスク**: 外部要因による影響

## 対策優先度

- 高確率 x 高影響 = 最優先
- 低確率 x 低影響 = 監視のみ
`
