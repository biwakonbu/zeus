package report

import (
	"bytes"
	"context"
	"strings"
	"text/template"
	"time"

	"github.com/biwakonbu/zeus/internal/analysis"
)

// ZeusConfig はレポート生成に必要な設定情報
type ZeusConfig struct {
	Project ProjectInfo
}

// ProjectInfo はプロジェクト情報
type ProjectInfo struct {
	ID          string
	Name        string
	Description string
	StartDate   string
}

// ProjectState はプロジェクト状態
type ProjectState struct {
	Health  string
	Summary SummaryStats
}

// SummaryStats はサマリー統計（Activity 統計）
type SummaryStats struct {
	TotalActivities int // JSON 互換のため "TotalActivities" を維持
	Completed       int
	InProgress      int
	Pending         int
}

// Generator はレポートを生成
type Generator struct {
	config   *ZeusConfig
	state    *ProjectState
	analysis *analysis.AnalysisResult
}

// NewGenerator は新しい Generator を作成
func NewGenerator(config *ZeusConfig, state *ProjectState, analysisResult *analysis.AnalysisResult) *Generator {
	return &Generator{
		config:   config,
		state:    state,
		analysis: analysisResult,
	}
}

// ReportData はテンプレートに渡すデータ
type ReportData struct {
	// 基本情報
	Timestamp   string
	Separator   string
	ProjectName string
	Health      string
	HealthClass string

	// サマリー統計
	TaskStats         SummaryStats
	CompletionPercent int
	CompletedPercent  int
	InProgressPercent int
	PendingPercent    int

	// 予測
	HasPrediction bool
	Completion    *analysis.CompletionPrediction

	// リスク
	HasRisk   bool
	Risk      *analysis.RiskPrediction
	RiskClass string

	// ベロシティ
	HasVelocity bool
	Velocity    *analysis.VelocityReport

	// グラフ
	HasGraph     bool
	GraphMermaid string

	// 推奨事項
	Recommendations []string
}

// GenerateText は TEXT 形式でレポートを生成
func (g *Generator) GenerateText(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	data := g.buildReportData()
	data.Separator = strings.Repeat("=", 60)

	tmpl, err := template.New("text").Parse(TextTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateHTML は HTML 形式でレポートを生成
func (g *Generator) GenerateHTML(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	data := g.buildReportData()

	tmpl, err := template.New("html").Parse(HTMLTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateMarkdown は Markdown 形式でレポートを生成
func (g *Generator) GenerateMarkdown(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	data := g.buildReportData()

	// グラフがある場合は Mermaid 形式を追加
	if g.analysis != nil && g.analysis.Graph != nil {
		data.HasGraph = true
		data.GraphMermaid = g.analysis.Graph.ToMermaid()
	}

	tmpl, err := template.New("markdown").Parse(MarkdownTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// buildReportData はレポートデータを構築
func (g *Generator) buildReportData() *ReportData {
	data := &ReportData{
		Timestamp:       time.Now().Format("2006-01-02 15:04:05"),
		ProjectName:     g.config.Project.Name,
		Health:          g.state.Health,
		HealthClass:     strings.ToLower(g.state.Health),
		TaskStats:       g.state.Summary,
		Recommendations: []string{},
	}

	// 完了率を計算
	if g.state.Summary.TotalActivities > 0 {
		total := float64(g.state.Summary.TotalActivities)
		data.CompletionPercent = int(float64(g.state.Summary.Completed) / total * 100)
		data.CompletedPercent = data.CompletionPercent
		data.InProgressPercent = int(float64(g.state.Summary.InProgress) / total * 100)
		data.PendingPercent = int(float64(g.state.Summary.Pending) / total * 100)
	}

	// 分析結果を追加
	if g.analysis != nil {
		if g.analysis.Completion != nil {
			data.HasPrediction = true
			data.Completion = g.analysis.Completion
		}
		if g.analysis.Risk != nil {
			data.HasRisk = true
			data.Risk = g.analysis.Risk
			data.RiskClass = strings.ToLower(string(g.analysis.Risk.OverallLevel))
		}
		if g.analysis.Velocity != nil {
			data.HasVelocity = true
			data.Velocity = g.analysis.Velocity
		}
	}

	// 推奨事項を生成
	data.Recommendations = g.generateRecommendations()

	return data
}

// generateRecommendations は推奨事項を生成
func (g *Generator) generateRecommendations() []string {
	recommendations := []string{}

	// 健全性に基づく推奨
	switch g.state.Health {
	case "Poor":
		recommendations = append(recommendations,
			"Project health is poor. Review blocked tasks and resolve dependencies.")
	case "Fair":
		recommendations = append(recommendations,
			"Project health is fair. Focus on completing in-progress tasks.")
	}

	// リスク要因に基づく推奨
	if g.analysis != nil && g.analysis.Risk != nil {
		for _, factor := range g.analysis.Risk.Factors {
			switch factor.Name {
			case "Blocked Tasks":
				recommendations = append(recommendations,
					"Resolve blocked tasks to improve project flow.")
			case "High WIP":
				recommendations = append(recommendations,
					"Consider limiting work in progress to improve focus and throughput.")
			case "Low Completion Rate":
				recommendations = append(recommendations,
					"Break down large tasks into smaller, manageable pieces.")
			case "Stalled Progress":
				recommendations = append(recommendations,
					"Review team capacity and identify blockers preventing progress.")
			}
		}
	}

	// 保留タスクが多い場合
	if g.state.Summary.Pending > 10 {
		recommendations = append(recommendations,
			"Large backlog detected. Consider prioritizing or archiving low-priority tasks.")
	}

	// 重複を除去
	seen := make(map[string]bool)
	unique := []string{}
	for _, rec := range recommendations {
		if !seen[rec] {
			seen[rec] = true
			unique = append(unique, rec)
		}
	}

	return unique
}
