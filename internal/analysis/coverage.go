package analysis

import (
	"context"
)

// CoverageIssueType はカバレッジ問題の種類
type CoverageIssueType string

const (
	CoverageIssueNoTasks  CoverageIssueType = "no_tasks"
	CoverageIssueOrphaned CoverageIssueType = "orphaned"
)

// CoverageIssueSeverity は問題の深刻度
type CoverageIssueSeverity string

const (
	CoverageSeverityWarning CoverageIssueSeverity = "warning"
	CoverageSeverityError   CoverageIssueSeverity = "error"
)

// CoverageIssue はカバレッジ問題を表す
type CoverageIssue struct {
	Type        CoverageIssueType     `json:"type"`
	EntityID    string                `json:"entity_id"`
	EntityTitle string                `json:"entity_title"`
	EntityType  string                `json:"entity_type"` // objective, task
	Severity    CoverageIssueSeverity `json:"severity"`
	Message     string                `json:"message"`
}

// CoverageAnalysis はカバレッジ分析結果
type CoverageAnalysis struct {
	Issues          []CoverageIssue `json:"issues"`
	CoverageScore   int             `json:"coverage_score"` // 0-100
	ObjectivesCover int             `json:"objectives_covered"`
	ObjectivesTotal int             `json:"objectives_total"`
}

// CoverageAnalyzer はカバレッジ分析を行う
type CoverageAnalyzer struct {
	objectives []ObjectiveInfo
	tasks      []TaskInfo
}

// NewCoverageAnalyzer は新しい CoverageAnalyzer を作成
func NewCoverageAnalyzer(
	objectives []ObjectiveInfo,
	tasks []TaskInfo,
) *CoverageAnalyzer {
	return &CoverageAnalyzer{
		objectives: objectives,
		tasks:      tasks,
	}
}

// Analyze はカバレッジ分析を実行
func (c *CoverageAnalyzer) Analyze(ctx context.Context) (*CoverageAnalysis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &CoverageAnalysis{
		Issues:          []CoverageIssue{},
		ObjectivesTotal: len(c.objectives),
	}

	// Task.ParentID を使って親子関係を追跡
	// NOTE: Activity は ParentID を持たない（UseCase 経由で Objective に紐づく）。
	// Activity のカバレッジは Unified Graph の構造解析で判定する。
	taskParents := make(map[string][]TaskInfo)
	orphanTasks := []TaskInfo{}
	for _, task := range c.tasks {
		if task.ParentID != "" {
			taskParents[task.ParentID] = append(taskParents[task.ParentID], task)
		} else {
			// 孤立タスクの候補
			orphanTasks = append(orphanTasks, task)
		}
	}

	// 1. Objective に Task が紐づいているかチェック（エラー）
	for _, obj := range c.objectives {
		if tasks, ok := taskParents[obj.ID]; ok && len(tasks) > 0 {
			result.ObjectivesCover++
		} else {
			result.Issues = append(result.Issues, CoverageIssue{
				Type:        CoverageIssueNoTasks,
				EntityID:    obj.ID,
				EntityTitle: obj.Title,
				EntityType:  "objective",
				Severity:    CoverageSeverityError,
				Message:     "Objective に Task が紐づいていません",
			})
		}
	}

	// 2. 孤立タスクをチェック（警告）
	// Objective に属さないトップレベルタスク
	for _, task := range orphanTasks {
		isOrphan := true
		// Objective の ID と一致する親を持つか確認
		for _, obj := range c.objectives {
			if task.ParentID == obj.ID {
				isOrphan = false
				break
			}
		}
		if isOrphan && task.ParentID == "" {
			result.Issues = append(result.Issues, CoverageIssue{
				Type:        CoverageIssueOrphaned,
				EntityID:    task.ID,
				EntityTitle: task.Title,
				EntityType:  "task",
				Severity:    CoverageSeverityWarning,
				Message:     "タスクが Objective に紐づいていません",
			})
		}
	}

	// カバレッジスコアを計算（Objective + Task ベース）
	if result.ObjectivesTotal > 0 {
		result.CoverageScore = (result.ObjectivesCover * 100) / result.ObjectivesTotal
	} else if len(c.tasks) > 0 {
		// Objective がない場合、タスクの孤立率で計算
		orphanCount := 0
		for _, issue := range result.Issues {
			if issue.Type == CoverageIssueOrphaned {
				orphanCount++
			}
		}
		result.CoverageScore = 100 - ((orphanCount * 100) / len(c.tasks))
	} else {
		result.CoverageScore = 100
	}

	return result, nil
}
