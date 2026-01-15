package report

import (
	"context"
	"strings"
	"testing"

	"github.com/biwakonbu/zeus/internal/analysis"
)

func TestNewGenerator(t *testing.T) {
	config := &ZeusConfig{
		Project: ProjectInfo{
			ID:          "test-project",
			Name:        "Test Project",
			Description: "A test project",
			StartDate:   "2024-01-01",
		},
	}

	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  5,
			InProgress: 3,
			Pending:    2,
		},
	}

	generator := NewGenerator(config, state, nil)

	if generator == nil {
		t.Fatal("NewGenerator returned nil")
	}

	if generator.config != config {
		t.Error("expected config to be set")
	}

	if generator.state != state {
		t.Error("expected state to be set")
	}
}

func TestGenerator_GenerateText(t *testing.T) {
	ctx := context.Background()

	config := &ZeusConfig{
		Project: ProjectInfo{
			Name: "Test Project",
		},
	}

	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  5,
			InProgress: 3,
			Pending:    2,
		},
	}

	generator := NewGenerator(config, state, nil)
	report, err := generator.GenerateText(ctx)

	if err != nil {
		t.Fatalf("GenerateText returned error: %v", err)
	}

	// 基本的な内容の確認
	if !strings.Contains(report, "Test Project") {
		t.Error("expected project name in report")
	}

	if !strings.Contains(report, "Good") {
		t.Error("expected health status in report")
	}

	// セパレーターを含むことを確認
	if !strings.Contains(report, "=") {
		t.Error("expected separator in text report")
	}
}

func TestGenerator_GenerateText_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	generator := NewGenerator(&ZeusConfig{}, &ProjectState{}, nil)
	_, err := generator.GenerateText(ctx)

	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestGenerator_GenerateHTML(t *testing.T) {
	ctx := context.Background()

	config := &ZeusConfig{
		Project: ProjectInfo{
			Name: "HTML Test Project",
		},
	}

	state := &ProjectState{
		Health: "Excellent",
		Summary: TaskStats{
			TotalTasks: 20,
			Completed:  15,
			InProgress: 3,
			Pending:    2,
		},
	}

	generator := NewGenerator(config, state, nil)
	report, err := generator.GenerateHTML(ctx)

	if err != nil {
		t.Fatalf("GenerateHTML returned error: %v", err)
	}

	// HTML構造の確認
	if !strings.Contains(report, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE declaration")
	}

	if !strings.Contains(report, "<html") {
		t.Error("expected html tag")
	}

	if !strings.Contains(report, "HTML Test Project") {
		t.Error("expected project name in HTML")
	}

	// CSSクラスの確認
	if !strings.Contains(report, "class=") {
		t.Error("expected CSS classes in HTML")
	}
}

func TestGenerator_GenerateHTML_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	generator := NewGenerator(&ZeusConfig{}, &ProjectState{}, nil)
	_, err := generator.GenerateHTML(ctx)

	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestGenerator_GenerateMarkdown(t *testing.T) {
	ctx := context.Background()

	config := &ZeusConfig{
		Project: ProjectInfo{
			Name: "Markdown Test",
		},
	}

	state := &ProjectState{
		Health: "Fair",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  3,
			InProgress: 4,
			Pending:    3,
		},
	}

	generator := NewGenerator(config, state, nil)
	report, err := generator.GenerateMarkdown(ctx)

	if err != nil {
		t.Fatalf("GenerateMarkdown returned error: %v", err)
	}

	// Markdown構造の確認
	if !strings.Contains(report, "#") {
		t.Error("expected markdown headers")
	}

	if !strings.Contains(report, "Markdown Test") {
		t.Error("expected project name in markdown")
	}
}

func TestGenerator_GenerateMarkdown_WithGraph(t *testing.T) {
	ctx := context.Background()

	config := &ZeusConfig{
		Project: ProjectInfo{
			Name: "Graph Test",
		},
	}

	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 5,
			Completed:  2,
			InProgress: 2,
			Pending:    1,
		},
	}

	// グラフ付き分析結果
	analysisResult := &analysis.AnalysisResult{
		Graph: &analysis.DependencyGraph{
			Nodes: map[string]*analysis.GraphNode{
				"task-1": {
					Task: &analysis.TaskInfo{
						ID:     "task-1",
						Title:  "Task 1",
						Status: "completed",
					},
				},
			},
			Edges: []analysis.Edge{},
		},
	}

	generator := NewGenerator(config, state, analysisResult)
	report, err := generator.GenerateMarkdown(ctx)

	if err != nil {
		t.Fatalf("GenerateMarkdown returned error: %v", err)
	}

	// Mermaidグラフを含むことを確認
	if !strings.Contains(report, "```mermaid") {
		t.Error("expected mermaid graph in markdown with graph analysis")
	}
}

func TestGenerator_GenerateMarkdown_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	generator := NewGenerator(&ZeusConfig{}, &ProjectState{}, nil)
	_, err := generator.GenerateMarkdown(ctx)

	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestGenerator_BuildReportData(t *testing.T) {
	config := &ZeusConfig{
		Project: ProjectInfo{
			Name: "Data Test",
		},
	}

	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 100,
			Completed:  50,
			InProgress: 30,
			Pending:    20,
		},
	}

	generator := NewGenerator(config, state, nil)
	data := generator.buildReportData()

	// 基本データの確認
	if data.ProjectName != "Data Test" {
		t.Errorf("expected ProjectName='Data Test', got '%s'", data.ProjectName)
	}

	if data.Health != "Good" {
		t.Errorf("expected Health='Good', got '%s'", data.Health)
	}

	// パーセンテージの確認
	if data.CompletionPercent != 50 {
		t.Errorf("expected CompletionPercent=50, got %d", data.CompletionPercent)
	}

	if data.InProgressPercent != 30 {
		t.Errorf("expected InProgressPercent=30, got %d", data.InProgressPercent)
	}

	if data.PendingPercent != 20 {
		t.Errorf("expected PendingPercent=20, got %d", data.PendingPercent)
	}
}

func TestGenerator_BuildReportData_WithAnalysis(t *testing.T) {
	config := &ZeusConfig{
		Project: ProjectInfo{Name: "Analysis Test"},
	}

	state := &ProjectState{
		Health: "Fair",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  3,
			InProgress: 4,
			Pending:    3,
		},
	}

	analysisResult := &analysis.AnalysisResult{
		Completion: &analysis.CompletionPrediction{
			RemainingTasks:  7,
			AverageVelocity: 2.5,
			EstimatedDate:   "2024-02-15",
			ConfidenceLevel: 70,
			MarginDays:      5,
		},
		Risk: &analysis.RiskPrediction{
			OverallLevel: analysis.RiskMedium,
			Score:        45,
			Factors: []analysis.RiskFactor{
				{Name: "High WIP", Impact: 5},
			},
		},
		Velocity: &analysis.VelocityReport{
			Last7Days:     5,
			Last14Days:    8,
			Last30Days:    15,
			WeeklyAverage: 3.75,
			Trend:         analysis.TrendStable,
		},
	}

	generator := NewGenerator(config, state, analysisResult)
	data := generator.buildReportData()

	// 予測データの確認
	if !data.HasPrediction {
		t.Error("expected HasPrediction=true")
	}

	if data.Completion == nil {
		t.Error("expected Completion to be set")
	}

	// リスクデータの確認
	if !data.HasRisk {
		t.Error("expected HasRisk=true")
	}

	if data.Risk == nil {
		t.Error("expected Risk to be set")
	}

	if data.RiskClass != "medium" {
		t.Errorf("expected RiskClass='medium', got '%s'", data.RiskClass)
	}

	// ベロシティデータの確認
	if !data.HasVelocity {
		t.Error("expected HasVelocity=true")
	}

	if data.Velocity == nil {
		t.Error("expected Velocity to be set")
	}
}

func TestGenerator_GenerateRecommendations_PoorHealth(t *testing.T) {
	config := &ZeusConfig{Project: ProjectInfo{Name: "Test"}}
	state := &ProjectState{
		Health: "Poor",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  2,
			InProgress: 3,
			Pending:    5,
		},
	}

	generator := NewGenerator(config, state, nil)
	recommendations := generator.generateRecommendations()

	// Poor health の推奨事項を含むことを確認
	found := false
	for _, rec := range recommendations {
		if strings.Contains(rec, "poor") || strings.Contains(rec, "Poor") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected recommendation for poor health")
	}
}

func TestGenerator_GenerateRecommendations_FairHealth(t *testing.T) {
	config := &ZeusConfig{Project: ProjectInfo{Name: "Test"}}
	state := &ProjectState{
		Health: "Fair",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  4,
			InProgress: 3,
			Pending:    3,
		},
	}

	generator := NewGenerator(config, state, nil)
	recommendations := generator.generateRecommendations()

	found := false
	for _, rec := range recommendations {
		if strings.Contains(rec, "fair") || strings.Contains(rec, "Fair") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected recommendation for fair health")
	}
}

func TestGenerator_GenerateRecommendations_LargeBacklog(t *testing.T) {
	config := &ZeusConfig{Project: ProjectInfo{Name: "Test"}}
	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 50,
			Completed:  10,
			InProgress: 5,
			Pending:    35, // > 10
		},
	}

	generator := NewGenerator(config, state, nil)
	recommendations := generator.generateRecommendations()

	found := false
	for _, rec := range recommendations {
		if strings.Contains(rec, "backlog") || strings.Contains(rec, "Backlog") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected recommendation for large backlog")
	}
}

func TestGenerator_GenerateRecommendations_WithRiskFactors(t *testing.T) {
	config := &ZeusConfig{Project: ProjectInfo{Name: "Test"}}
	state := &ProjectState{
		Health: "Good",
		Summary: TaskStats{
			TotalTasks: 10,
			Completed:  5,
			InProgress: 3,
			Pending:    2,
		},
	}

	analysisResult := &analysis.AnalysisResult{
		Risk: &analysis.RiskPrediction{
			OverallLevel: analysis.RiskMedium,
			Score:        50,
			Factors: []analysis.RiskFactor{
				{Name: "Blocked Tasks", Impact: 7},
				{Name: "High WIP", Impact: 5},
			},
		},
	}

	generator := NewGenerator(config, state, analysisResult)
	recommendations := generator.generateRecommendations()

	// リスク要因に基づく推奨事項を確認
	foundBlockedRec := false
	foundWIPRec := false
	for _, rec := range recommendations {
		if strings.Contains(rec, "blocked") {
			foundBlockedRec = true
		}
		if strings.Contains(rec, "work in progress") || strings.Contains(rec, "WIP") {
			foundWIPRec = true
		}
	}

	if !foundBlockedRec {
		t.Error("expected recommendation for blocked tasks")
	}

	if !foundWIPRec {
		t.Error("expected recommendation for high WIP")
	}
}

func TestGenerator_GenerateRecommendations_NoDuplicates(t *testing.T) {
	config := &ZeusConfig{Project: ProjectInfo{Name: "Test"}}
	state := &ProjectState{
		Health: "Poor",
		Summary: TaskStats{
			TotalTasks: 50,
			Completed:  5,
			InProgress: 10,
			Pending:    35,
		},
	}

	analysisResult := &analysis.AnalysisResult{
		Risk: &analysis.RiskPrediction{
			Factors: []analysis.RiskFactor{
				{Name: "Blocked Tasks", Impact: 7},
				{Name: "Blocked Tasks", Impact: 7}, // 重複
			},
		},
	}

	generator := NewGenerator(config, state, analysisResult)
	recommendations := generator.generateRecommendations()

	// 重複がないことを確認
	seen := make(map[string]bool)
	for _, rec := range recommendations {
		if seen[rec] {
			t.Errorf("found duplicate recommendation: %s", rec)
		}
		seen[rec] = true
	}
}

func TestGenerator_ZeroTasks(t *testing.T) {
	ctx := context.Background()

	config := &ZeusConfig{
		Project: ProjectInfo{Name: "Empty Project"},
	}

	state := &ProjectState{
		Health: "N/A",
		Summary: TaskStats{
			TotalTasks: 0,
			Completed:  0,
			InProgress: 0,
			Pending:    0,
		},
	}

	generator := NewGenerator(config, state, nil)

	// TEXT形式
	text, err := generator.GenerateText(ctx)
	if err != nil {
		t.Fatalf("GenerateText returned error: %v", err)
	}
	if text == "" {
		t.Error("expected non-empty text report")
	}

	// HTML形式
	html, err := generator.GenerateHTML(ctx)
	if err != nil {
		t.Fatalf("GenerateHTML returned error: %v", err)
	}
	if html == "" {
		t.Error("expected non-empty HTML report")
	}

	// Markdown形式
	md, err := generator.GenerateMarkdown(ctx)
	if err != nil {
		t.Fatalf("GenerateMarkdown returned error: %v", err)
	}
	if md == "" {
		t.Error("expected non-empty markdown report")
	}
}
