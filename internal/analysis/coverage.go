package analysis

import (
	"context"
)

// CoverageIssueType はカバレッジ問題の種類
type CoverageIssueType string

const (
	CoverageIssueNoDeliverables CoverageIssueType = "no_deliverables"
	CoverageIssueNoTasks        CoverageIssueType = "no_tasks"
	CoverageIssueUnlinkedTasks  CoverageIssueType = "unlinked_tasks"
	CoverageIssueOrphaned       CoverageIssueType = "orphaned"
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
	EntityType  string                `json:"entity_type"` // objective, deliverable, task
	Severity    CoverageIssueSeverity `json:"severity"`
	Message     string                `json:"message"`
}

// CoverageAnalysis はカバレッジ分析結果
type CoverageAnalysis struct {
	Issues          []CoverageIssue `json:"issues"`
	CoverageScore   int             `json:"coverage_score"` // 0-100
	ObjectivesCover int             `json:"objectives_covered"`
	ObjectivesTotal int             `json:"objectives_total"`
	DeliverablesOk  int             `json:"deliverables_ok"`
	DeliverablesErr int             `json:"deliverables_err"`
}

// CoverageAnalyzer はカバレッジ分析を行う
type CoverageAnalyzer struct {
	objectives   []ObjectiveInfo
	deliverables []DeliverableInfo
	tasks        []TaskInfo
}

// NewCoverageAnalyzer は新しい CoverageAnalyzer を作成
func NewCoverageAnalyzer(
	objectives []ObjectiveInfo,
	deliverables []DeliverableInfo,
	tasks []TaskInfo,
) *CoverageAnalyzer {
	return &CoverageAnalyzer{
		objectives:   objectives,
		deliverables: deliverables,
		tasks:        tasks,
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

	// Objective → Deliverable のマッピング
	objToDeliverables := make(map[string][]DeliverableInfo)
	for _, del := range c.deliverables {
		if del.ObjectiveID != "" {
			objToDeliverables[del.ObjectiveID] = append(objToDeliverables[del.ObjectiveID], del)
		}
	}

	// Deliverable → Task のマッピング（現在の実装では Task の deliverable_id が使用されていない場合がある）
	// Task.ParentID を使って親子関係を追跡
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

	// 1. Objective に Deliverable が紐づいていないかチェック（エラー）
	for _, obj := range c.objectives {
		deliverables := objToDeliverables[obj.ID]
		if len(deliverables) == 0 {
			result.Issues = append(result.Issues, CoverageIssue{
				Type:        CoverageIssueNoDeliverables,
				EntityID:    obj.ID,
				EntityTitle: obj.Title,
				EntityType:  "objective",
				Severity:    CoverageSeverityError,
				Message:     "Objective に Deliverable が紐づいていません",
			})
		} else {
			result.ObjectivesCover++
		}
	}

	// 2. Deliverable に Task が紐づいていないかチェック（警告）
	// 現在の実装では Task から Deliverable への直接参照がないため、
	// Deliverable の子ノードとして Task があるかを確認
	deliverableHasTasks := make(map[string]bool)
	for _, del := range c.deliverables {
		// Task.ParentID が Deliverable.ID と一致するかをチェック
		if tasks, ok := taskParents[del.ID]; ok && len(tasks) > 0 {
			deliverableHasTasks[del.ID] = true
			result.DeliverablesOk++
		} else {
			deliverableHasTasks[del.ID] = false
			result.DeliverablesErr++
			result.Issues = append(result.Issues, CoverageIssue{
				Type:        CoverageIssueNoTasks,
				EntityID:    del.ID,
				EntityTitle: del.Title,
				EntityType:  "deliverable",
				Severity:    CoverageSeverityWarning,
				Message:     "Deliverable に Task が紐づいていません",
			})
		}
	}

	// 3. 孤立タスクをチェック（警告）
	// Objective や Deliverable に属さないトップレベルタスク
	for _, task := range orphanTasks {
		// 親がいない場合は孤立の可能性
		isOrphan := true
		// Deliverable の ID と一致する親を持つか確認
		for _, del := range c.deliverables {
			if task.ParentID == del.ID {
				isOrphan = false
				break
			}
		}
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
				Message:     "タスクが Objective/Deliverable に紐づいていません",
			})
		}
	}

	// カバレッジスコアを計算
	if result.ObjectivesTotal > 0 {
		score := (result.ObjectivesCover * 100) / result.ObjectivesTotal
		if len(c.deliverables) > 0 {
			delScore := (result.DeliverablesOk * 100) / len(c.deliverables)
			score = (score + delScore) / 2
		}
		result.CoverageScore = score
	} else if len(c.deliverables) > 0 {
		result.CoverageScore = (result.DeliverablesOk * 100) / len(c.deliverables)
	} else if len(c.tasks) > 0 {
		// Objective/Deliverable がない場合、タスクの孤立率で計算
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
