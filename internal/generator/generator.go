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

1. **プロジェクト全体の把握**: タスク、目標、リソース、WBS階層を俯瞰
2. **優先順位付け**: 重要度・緊急度・クリティカルパスに基づいた判断
3. **リスク検知**: 潜在的な問題を早期発見、予測分析の活用
4. **進捗管理**: 全体の進捗状況をダッシュボードで追跡

## コマンド一覧

### 基本操作
- ` + "`zeus init`" + ` - プロジェクト初期化
- ` + "`zeus status`" + ` - 現在の状態を確認
- ` + "`zeus add <entity> <name> [options]`" + ` - エンティティ追加
- ` + "`zeus list [entity]`" + ` - 一覧表示
- ` + "`zeus doctor`" + ` - システム診断
- ` + "`zeus fix [--dry-run]`" + ` - 修復

### 承認管理
- ` + "`zeus pending`" + ` - 承認待ち一覧
- ` + "`zeus approve <id>`" + ` - 承認
- ` + "`zeus reject <id> [--reason \"\"]`" + ` - 却下

### 状態管理
- ` + "`zeus snapshot create [label]`" + ` - スナップショット作成
- ` + "`zeus snapshot list [-n limit]`" + ` - スナップショット一覧
- ` + "`zeus snapshot restore <timestamp>`" + ` - 復元
- ` + "`zeus history [-n limit]`" + ` - 履歴表示

### AI機能
- ` + "`zeus suggest [--limit N] [--impact level]`" + ` - AI提案生成
- ` + "`zeus apply <suggestion-id>`" + ` - 提案を個別適用
- ` + "`zeus apply --all [--dry-run]`" + ` - 全提案適用
- ` + "`zeus explain <entity-id> [--context]`" + ` - 詳細説明

### 分析機能（Phase 4-6）
- ` + "`zeus graph [--format text|dot|mermaid] [-o file]`" + ` - 依存関係グラフ
- ` + "`zeus predict [completion|risk|velocity|all]`" + ` - 予測分析
- ` + "`zeus report [--format text|html|markdown] [-o file]`" + ` - レポート生成
- ` + "`zeus dashboard [--port 8080] [--no-open] [--dev]`" + ` - Webダッシュボード起動

## Phase 6 機能（WBS・タイムライン）

### タスク追加時のオプション
` + "```bash" + `
zeus add task "タスク名" \
  --parent <id>      # 親タスクID（WBS階層構造）
  --start <date>     # 開始日（ISO8601: 2026-01-17）
  --due <date>       # 期限日（ISO8601: 2026-01-31）
  --progress <0-100> # 進捗率
  --wbs <code>       # WBSコード（例: 1.2.3）
` + "```" + `

### ダッシュボード機能
- **WBS階層ビュー** - 親子関係のツリー表示
- **タイムラインビュー** - ガントチャート、クリティカルパス
- **グラフビュー** - 依存関係の可視化、影響範囲ハイライト
- **リアルタイム更新** - SSE による自動更新

### API エンドポイント
- ` + "`GET /api/status`" + ` - プロジェクト状態
- ` + "`GET /api/tasks`" + ` - タスク一覧
- ` + "`GET /api/graph`" + ` - 依存関係グラフ（Mermaid形式）
- ` + "`GET /api/predict`" + ` - 予測分析結果
- ` + "`GET /api/wbs`" + ` - WBS階層構造
- ` + "`GET /api/timeline`" + ` - タイムラインとクリティカルパス
- ` + "`GET /api/downstream?task_id=X`" + ` - 下流・上流タスク取得
- ` + "`GET /api/events`" + ` - SSE ストリーム

### 循環参照検出
ParentID の循環参照は自動検出され、エラーとして防止されます。

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

1. **WBS 作成**: タスクの分解と階層構造化
2. **見積もり**: 工数の見積もり
3. **依存関係分析**: タスク間の依存関係を特定
4. **タイムライン設計**: スケジュールの策定、クリティカルパス分析

## 基本コマンド

- ` + "`zeus add task <name> [options]`" + ` - タスクを追加
- ` + "`zeus list tasks [--status <status>]`" + ` - タスク一覧
- ` + "`zeus graph [--format mermaid]`" + ` - 依存関係グラフ
- ` + "`zeus predict`" + ` - 予測分析

## Phase 6 対応

### WBS階層の作成

1. 親タスクを作成:
` + "```bash" + `
zeus add task "フェーズ1: 設計" --wbs "1"
` + "```" + `

2. 子タスクを追加:
` + "```bash" + `
zeus add task "要件定義" --parent <親のID> --wbs "1.1"
zeus add task "アーキテクチャ設計" --parent <親のID> --wbs "1.2"
` + "```" + `

3. さらに孫タスクを追加:
` + "```bash" + `
zeus add task "DB設計" --parent <1.2のID> --wbs "1.2.1"
zeus add task "API設計" --parent <1.2のID> --wbs "1.2.2"
` + "```" + `

### タイムライン設計

` + "```bash" + `
zeus add task "実装" \
  --start 2026-01-20 \
  --due 2026-01-31 \
  --progress 0 \
  --assignee "開発チーム" \
  --priority high
` + "```" + `

### 依存関係の指定

Dependencies フィールドで依存関係を指定すると、
クリティカルパス計算とタイムライン表示に反映されます。

` + "```yaml" + `
tasks:
  - id: task-design
    title: "設計"
    dependencies: []
  - id: task-implement
    title: "実装"
    dependencies: ["task-design"]  # 設計完了後に開始
  - id: task-test
    title: "テスト"
    dependencies: ["task-implement"]  # 実装完了後に開始
` + "```" + `

## タスク追加オプション一覧

| オプション | 説明 | 例 |
|-----------|------|-----|
| ` + "`--parent <id>`" + ` | 親タスクID（WBS階層） | ` + "`--parent abc123`" + ` |
| ` + "`--start <date>`" + ` | 開始日（ISO8601） | ` + "`--start 2026-01-20`" + ` |
| ` + "`--due <date>`" + ` | 期限日（ISO8601） | ` + "`--due 2026-01-31`" + ` |
| ` + "`--progress <0-100>`" + ` | 進捗率 | ` + "`--progress 50`" + ` |
| ` + "`--wbs <code>`" + ` | WBSコード | ` + "`--wbs 1.2.3`" + ` |
| ` + "`--priority <level>`" + ` | 優先度 | ` + "`--priority high`" + ` |
| ` + "`--assignee <name>`" + ` | 担当者 | ` + "`--assignee \"山田\"`" + ` |

## 計画の原則

1. **保守的な見積もり**: バッファを確保
2. **段階的計画**: 大きなタスクは分割（WBS活用）
3. **柔軟性**: 変更に対応できる余地を残す
4. **クリティカルパス**: 遅延が許されないタスクを特定

## 出力形式

` + "```yaml" + `
tasks:
  - id: task-1
    title: "タスク名"
    parent_id: ""
    wbs_code: "1.1"
    start_date: "2026-01-20"
    due_date: "2026-01-31"
    progress: 0
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

1. **進捗レビュー**: タスクの進捗を評価、予測分析の活用
2. **品質チェック**: 成果物の品質を確認
3. **リスク評価**: 潜在的な問題を評価、クリティカルパス監視
4. **改善提案**: 改善点を提案

## コマンド

### 基本レビュー
- ` + "`zeus status`" + ` - 状態を確認
- ` + "`zeus pending`" + ` - 承認待ちアイテムを確認
- ` + "`zeus approve <id>`" + ` - アイテムを承認
- ` + "`zeus reject <id> [--reason \"\"]`" + ` - 却下

### 分析ツール
- ` + "`zeus graph [--format text|dot|mermaid]`" + ` - 依存関係グラフ表示
- ` + "`zeus predict completion`" + ` - 完了日予測
- ` + "`zeus predict risk`" + ` - リスク分析
- ` + "`zeus predict velocity`" + ` - ベロシティ分析
- ` + "`zeus predict all`" + ` - 全予測分析
- ` + "`zeus report [--format text|html|markdown]`" + ` - プロジェクトレポート生成

### リアルタイム監視
- ` + "`zeus dashboard`" + ` - Webダッシュボードで監視
  - タスクグラフ表示
  - WBS階層ビュー
  - タイムライン・クリティカルパス表示
  - 影響範囲ハイライト（下流/上流タスク）

## Phase 6 レビュー項目

### WBS階層のチェック
- 循環参照がないか（自動検出される）
- 適切な粒度で分割されているか
- WBSコードが一貫しているか

### タイムラインのチェック
- 開始日・期限日が設定されているか
- クリティカルパスが特定されているか
- 依存関係が正しく設定されているか
- スラック（余裕時間）が適切か

### 進捗確認
- 進捗率が正確に更新されているか
- 遅延タスクが特定されているか
- ボトルネックが把握されているか

## レビュー基準

1. **完了の定義**: 明確な完了条件を確認
2. **品質基準**: 定義された品質基準を満たしているか
3. **依存関係**: 後続タスクへの影響

## 承認レベル

| レベル | 説明 | 動作 |
|--------|------|------|
| **auto** | 自動承認 | 低リスク操作、即時実行 |
| **notify** | 通知のみ | 中リスク操作、ログ記録して実行 |
| **approve** | 明示的承認必要 | 高リスク操作、承認待ちキューに追加 |

## レポート活用

` + "```bash" + `
# テキスト形式のレポート
zeus report

# HTML形式でファイル出力
zeus report --format html -o report.html

# Markdown形式
zeus report --format markdown -o report.md
` + "```" + `
`

// Skill Templates

const projectScanTemplate = `---
description: プロジェクト全体をスキャンし、現在の状態を分析するスキル
---

# zeus-project-scan

プロジェクト全体をスキャンし、現在の状態を分析するスキル。

## 概要

Zeus プロジェクト（{{.ProjectName}}）の構造、タスク、進捗、依存関係、WBS階層、タイムラインを分析します。

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
  graph:
    cycles: []            # 循環参照リスト
    isolated: []          # 孤立タスクリスト
    max_depth: 3          # 依存関係の最大深度
  prediction:
    estimated_completion: "2026-03-31"
    risk_level: "medium"
    velocity_trend: "stable|improving|declining"
  wbs:
    max_depth: 3          # WBS階層の最大深度
    total_nodes: 15       # WBSノード総数
    orphan_tasks: []      # 親が見つからないタスク
  timeline:
    project_start: "2026-01-01"
    project_end: "2026-03-31"
    critical_path_length: 5     # クリティカルパス上のタスク数
    overdue_tasks: []           # 期限超過タスク
  risks: []
` + "```" + `

## 使用方法

1. ` + "`zeus status`" + ` コマンドで基本情報取得
2. ` + "`zeus graph`" + ` で依存関係グラフ確認
3. ` + "`zeus predict all`" + ` で予測分析実行
4. ` + "`zeus dashboard`" + ` で可視化（推奨）
5. 改善提案を生成

## コマンド実行例

` + "```bash" + `
# 基本状態の確認
zeus status

# 依存関係グラフ（Mermaid形式）
zeus graph --format mermaid

# 全予測分析
zeus predict all

# Webダッシュボードで可視化
zeus dashboard
` + "```" + `

## ダッシュボードAPI

スキャン結果をプログラムで取得する場合:

` + "```bash" + `
# ダッシュボード起動後
curl http://localhost:8080/api/status
curl http://localhost:8080/api/tasks
curl http://localhost:8080/api/graph
curl http://localhost:8080/api/predict
curl http://localhost:8080/api/wbs
curl http://localhost:8080/api/timeline
` + "```" + `

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

Zeus プロジェクト（{{.ProjectName}}）の状態を分析し、次に取り組むべきタスクや改善案を提案します。

## 実行方法

` + "```bash" + `
# 提案生成
zeus suggest [--limit N] [--impact high|medium|low]

# 例: 高影響の提案を5件表示
zeus suggest --limit 5 --impact high
` + "```" + `

## 出力

` + "```yaml" + `
suggestions:
  - id: suggestion-1
    type: "new_task|priority_change|dependency|schedule"
    description: "提案の説明"
    rationale: "理由"
    impact: "high|medium|low"
  - id: suggestion-2
    type: "schedule"
    description: "クリティカルパス上のタスクに遅延リスク"
    rationale: "タスクXのスラックが0で、依存タスクに影響"
    impact: "high"
` + "```" + `

## Phase 6 対応

### WBS階層を考慮した提案
- 親タスクの完了度に基づく子タスクの優先度調整
- 階層のバランスチェック（深すぎる/浅すぎる階層の検出）
- 親が未設定のタスクに対する整理提案

### タイムライン最適化
- クリティカルパス上のタスクの優先度向上
- 期限切れタスクの警告
- スラックが少ないタスクの注意喚起
- 依存関係のボトルネック特定

### 進捗整合性
- 親タスクと子タスクの進捗率の不整合検出
- 長期間更新されていないタスクの警告

## 提案の適用

` + "```bash" + `
# 個別適用
zeus apply <suggestion-id>

# 全提案適用（ドライラン可能）
zeus apply --all [--dry-run]
` + "```" + `

## アルゴリズム

1. 現在のタスク状態を分析
2. WBS階層と依存関係を考慮
3. クリティカルパスを計算
4. 目標との差分を計算
5. 優先度に基づいて提案を生成

## 関連スキル

- zeus-project-scan - プロジェクト状態のスキャン
- zeus-risk-analysis - リスク分析
`

const riskAnalysisTemplate = `---
description: プロジェクトのリスクを分析し、対策を提案するスキル
---

# zeus-risk-analysis

プロジェクトのリスクを分析するスキル。

## 概要

Zeus プロジェクト（{{.ProjectName}}）のリスク要因を特定し、対策を提案します。

## 実行方法

` + "```bash" + `
# リスク分析
zeus predict risk

# 全予測分析（完了日・リスク・ベロシティ）
zeus predict all

# グラフで循環参照・孤立タスク確認
zeus graph
` + "```" + `

## 出力

` + "```yaml" + `
risks:
  - id: risk-1
    category: "schedule|resource|technical|external|wbs|dependency"
    description: "リスクの説明"
    probability: "high|medium|low"
    impact: "high|medium|low"
    mitigation: "対策"
` + "```" + `

## predict コマンド出力例

` + "```json" + `
{
  "risk": {
    "overall_level": "medium",
    "factors": [
      {
        "name": "Schedule Pressure",
        "description": "クリティカルパス上のタスクに遅延",
        "impact": 0.7
      },
      {
        "name": "Dependency Bottleneck",
        "description": "タスクXに5つの依存タスクが集中",
        "impact": 0.5
      }
    ],
    "score": 0.65
  }
}
` + "```" + `

## リスクカテゴリ

### 従来のリスク
1. **スケジュールリスク**: 遅延、見積もり誤差
2. **リソースリスク**: 人員不足、スキル不足
3. **技術リスク**: 技術的困難、依存関係
4. **外部リスク**: 外部要因による影響

### Phase 6 固有のリスク

#### WBS階層のリスク
- **循環参照**: ParentID の循環（自動検出・防止済み）
- **階層の深さ不均衡**: 一部が深すぎる/浅すぎる
- **親タスクの進捗と子タスクの不整合**: 子タスク完了済みだが親が未完了
- **孤立タスク**: 親が削除されて参照切れ

#### タイムラインのリスク
- **クリティカルパス上の遅延**: スラック0のタスクが遅延
- **依存関係のボトルネック**: 特定タスクに依存が集中
- **スラック不足によるバッファ欠如**: 全体的に余裕がない
- **期限超過タスクの累積**: 未対処の遅延タスク

## 対策優先度

| 確率 | 影響 | 優先度 |
|------|------|--------|
| 高 | 高 | **最優先** |
| 高 | 中 | 優先対応 |
| 中 | 高 | 優先対応 |
| 中 | 中 | 計画的対応 |
| 低 | 低 | 監視のみ |

## 分析結果の活用

1. ` + "`zeus predict risk`" + ` でリスク要因を特定
2. ` + "`zeus graph`" + ` で依存関係を可視化
3. ` + "`zeus dashboard`" + ` でリアルタイム監視
4. ` + "`zeus suggest`" + ` で対策提案を取得
5. ` + "`zeus apply <id>`" + ` で対策を適用

## 関連

- zeus-project-scan - プロジェクトスキャン
- zeus-task-suggest - タスク提案
`
