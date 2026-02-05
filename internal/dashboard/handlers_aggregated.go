package dashboard

import (
	"net/http"

	"github.com/biwakonbu/zeus/internal/core"
)

// =============================================================================
// WBS Aggregated API 型定義（4視点用の集約データ）
// =============================================================================

// WBSAggregatedResponse は WBS 集約 API のレスポンス
type WBSAggregatedResponse struct {
	Progress  *ProgressAggregation `json:"progress"`
	Issues    *IssueAggregation    `json:"issues"`
	Coverage  *CoverageAggregation `json:"coverage"`
	Resources *ResourceAggregation `json:"resources"`
}

// ProgressAggregation は進捗集約データ（ツリーマップ用）
type ProgressAggregation struct {
	Vision        *ProgressNode   `json:"vision,omitempty"`
	Objectives    []*ProgressNode `json:"objectives"`
	TotalProgress int             `json:"total_progress"`
}

// ProgressNode は進捗ツリーのノード
type ProgressNode struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	NodeType      string          `json:"node_type"`
	Progress      int             `json:"progress"`
	Status        string          `json:"status"`
	ChildrenCount int             `json:"children_count"`
	Children      []*ProgressNode `json:"children,omitempty"`
}

// IssueAggregation は問題集中データ（バブルチャート用）
type IssueAggregation struct {
	Items       []*IssueBubble `json:"items"`
	TotalIssues int            `json:"total_issues"`
	MaxSeverity string         `json:"max_severity"`
}

// IssueBubble はバブルチャート用のデータ
type IssueBubble struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	NodeType     string  `json:"node_type"`
	ProblemCount int     `json:"problem_count"`
	RiskCount    int     `json:"risk_count"`
	TotalIssues  int     `json:"total_issues"`
	MaxSeverity  string  `json:"max_severity"`
	RiskScore    float64 `json:"risk_score"`
	Progress     int     `json:"progress"`
}

// CoverageAggregation はカバレッジデータ（サンバースト用）
type CoverageAggregation struct {
	Root          *CoverageNode `json:"root"`
	CoverageScore int           `json:"coverage_score"`
	OrphanedTasks []string      `json:"orphaned_tasks"`
	MissingLinks  []string      `json:"missing_links"`
}

// CoverageNode はサンバースト用のノード
type CoverageNode struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	NodeType  string          `json:"node_type"`
	HasIssue  bool            `json:"has_issue"`
	IssueType string          `json:"issue_type,omitempty"`
	Value     int             `json:"value"`
	Children  []*CoverageNode `json:"children,omitempty"`
}

// ResourceAggregation はリソース配分データ（ヒートマップ用）
type ResourceAggregation struct {
	Assignees  []string         `json:"assignees"`
	Objectives []string         `json:"objectives"`
	Matrix     [][]ResourceCell `json:"matrix"`
}

// ResourceCell はヒートマップのセル
type ResourceCell struct {
	TaskCount    int `json:"task_count"`
	Progress     int `json:"progress"`
	BlockedCount int `json:"blocked_count"`
}

// =============================================================================
// WBS Aggregated API ハンドラー
// =============================================================================

// handleAPIWBSAggregated は WBS 集約 API を処理
func (s *Server) handleAPIWBSAggregated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "GET メソッドのみ許可されています")
		return
	}

	ctx := r.Context()
	fileStore := s.zeus.FileStore()

	// Activity を取得（ListItem に変換）
	activities := []core.ListItem{}
	actFiles, _ := fileStore.ListDir(ctx, "activities")
	for _, file := range actFiles {
		if !hasYamlSuffix(file) {
			continue
		}
		var act core.ActivityEntity
		if err := fileStore.ReadYaml(ctx, "activities/"+file, &act); err == nil {
			activities = append(activities, act.ToListItem())
		}
	}

	// Objective を取得
	objectives := []core.ObjectiveEntity{}
	objFiles, err := fileStore.ListDir(ctx, "objectives")
	if err == nil {
		for _, file := range objFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var obj core.ObjectiveEntity
			if err := fileStore.ReadYaml(ctx, "objectives/"+file, &obj); err == nil {
				objectives = append(objectives, obj)
			}
		}
	}

	// Deliverable を取得
	deliverables := []core.DeliverableEntity{}
	delFiles, err := fileStore.ListDir(ctx, "deliverables")
	if err == nil {
		for _, file := range delFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var del core.DeliverableEntity
			if err := fileStore.ReadYaml(ctx, "deliverables/"+file, &del); err == nil {
				deliverables = append(deliverables, del)
			}
		}
	}

	// Vision を取得
	var vision core.Vision
	_ = fileStore.ReadYaml(ctx, "vision.yaml", &vision)

	// Problem を取得
	problems := []core.ProblemEntity{}
	probFiles, err := fileStore.ListDir(ctx, "problems")
	if err == nil {
		for _, file := range probFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var prob core.ProblemEntity
			if err := fileStore.ReadYaml(ctx, "problems/"+file, &prob); err == nil {
				problems = append(problems, prob)
			}
		}
	}

	// Risk を取得
	risks := []core.RiskEntity{}
	riskFiles, err := fileStore.ListDir(ctx, "risks")
	if err == nil {
		for _, file := range riskFiles {
			if !hasYamlSuffix(file) {
				continue
			}
			var risk core.RiskEntity
			if err := fileStore.ReadYaml(ctx, "risks/"+file, &risk); err == nil {
				risks = append(risks, risk)
			}
		}
	}

	// 集約データを構築
	response := WBSAggregatedResponse{
		Progress:  buildProgressAggregation(vision, objectives, deliverables, activities),
		Issues:    buildIssueAggregation(objectives, deliverables, problems, risks),
		Coverage:  buildCoverageAggregation(vision, objectives, deliverables, activities),
		Resources: buildResourceAggregation(objectives, activities),
	}

	writeJSON(w, http.StatusOK, response)
}

// =============================================================================
// Aggregation ヘルパー関数
// =============================================================================

// buildProgressAggregation は進捗集約データを構築
// items は Activity から変換された ListItem のリスト
func buildProgressAggregation(vision core.Vision, objectives []core.ObjectiveEntity, deliverables []core.DeliverableEntity, items []core.ListItem) *ProgressAggregation {
	result := &ProgressAggregation{
		Objectives: []*ProgressNode{},
	}

	// Objective ID → Deliverables マップ
	objDeliverables := make(map[string][]core.DeliverableEntity)
	for _, del := range deliverables {
		objDeliverables[del.ObjectiveID] = append(objDeliverables[del.ObjectiveID], del)
	}

	// Deliverable ID → Items マップ
	// （Item の DeliverableID がないので、ParentID から判定）
	delItems := make(map[string][]core.ListItem)
	for _, item := range items {
		if item.ParentID != "" {
			delItems[item.ParentID] = append(delItems[item.ParentID], item)
		}
	}

	// Objective ごとの進捗計算
	totalProgress := 0
	for _, obj := range objectives {
		objNode := &ProgressNode{
			ID:       obj.ID,
			Title:    obj.Title,
			NodeType: "objective",
			Progress: obj.Progress,
			Status:   string(obj.Status),
			Children: []*ProgressNode{},
		}

		// 関連 Deliverable を追加
		for _, del := range objDeliverables[obj.ID] {
			delNode := &ProgressNode{
				ID:       del.ID,
				Title:    del.Title,
				NodeType: "deliverable",
				Progress: del.Progress,
				Status:   string(del.Status),
				Children: []*ProgressNode{},
			}

			// 関連 Activity を追加
			for _, item := range delItems[del.ID] {
				itemNode := &ProgressNode{
					ID:       item.ID,
					Title:    item.Title,
					NodeType: "task",
					Progress: item.Progress,
					Status:   string(item.Status),
				}
				delNode.Children = append(delNode.Children, itemNode)
			}
			delNode.ChildrenCount = len(delNode.Children)

			objNode.Children = append(objNode.Children, delNode)
		}
		objNode.ChildrenCount = len(objNode.Children)

		result.Objectives = append(result.Objectives, objNode)
		totalProgress += obj.Progress
	}

	// 平均進捗
	if len(objectives) > 0 {
		result.TotalProgress = totalProgress / len(objectives)
	}

	// Vision を追加（存在する場合）
	if vision.Title != "" {
		result.Vision = &ProgressNode{
			ID:            "vision",
			Title:         vision.Title,
			NodeType:      "vision",
			Progress:      result.TotalProgress,
			Status:        "active",
			ChildrenCount: len(objectives),
		}
	}

	return result
}

// buildIssueAggregation は問題集中データを構築
func buildIssueAggregation(objectives []core.ObjectiveEntity, deliverables []core.DeliverableEntity, problems []core.ProblemEntity, risks []core.RiskEntity) *IssueAggregation {
	result := &IssueAggregation{
		Items:       []*IssueBubble{},
		MaxSeverity: "low",
	}

	// エンティティ ID → 問題/リスク集計
	issueMap := make(map[string]*IssueBubble)

	// Objective 用のバブルを作成
	for _, obj := range objectives {
		issueMap[obj.ID] = &IssueBubble{
			ID:       obj.ID,
			Title:    obj.Title,
			NodeType: "objective",
			Progress: obj.Progress,
		}
	}

	// Deliverable 用のバブルを作成
	for _, del := range deliverables {
		issueMap[del.ID] = &IssueBubble{
			ID:       del.ID,
			Title:    del.Title,
			NodeType: "deliverable",
			Progress: del.Progress,
		}
	}

	// Problem を集計
	for _, prob := range problems {
		targetID := prob.ObjectiveID
		if targetID == "" {
			targetID = prob.DeliverableID
		}
		if targetID == "" {
			continue
		}
		if bubble, ok := issueMap[targetID]; ok {
			bubble.ProblemCount++
			bubble.TotalIssues++
			// 深刻度を更新
			if severityRank(string(prob.Severity)) > severityRank(bubble.MaxSeverity) {
				bubble.MaxSeverity = string(prob.Severity)
			}
		}
	}

	// Risk を集計
	for _, risk := range risks {
		targetID := risk.ObjectiveID
		if targetID == "" {
			targetID = risk.DeliverableID
		}
		if targetID == "" {
			continue
		}
		if bubble, ok := issueMap[targetID]; ok {
			bubble.RiskCount++
			bubble.TotalIssues++
			// リスクスコアを加算
			bubble.RiskScore += float64(riskScoreToInt(string(risk.RiskScore)))
			// 深刻度を更新
			riskSeverity := riskScoreToSeverity(string(risk.RiskScore))
			if severityRank(riskSeverity) > severityRank(bubble.MaxSeverity) {
				bubble.MaxSeverity = riskSeverity
			}
		}
	}

	// 結果に追加（問題がある項目のみ）
	for _, bubble := range issueMap {
		if bubble.TotalIssues > 0 {
			result.Items = append(result.Items, bubble)
			result.TotalIssues += bubble.TotalIssues
			if severityRank(bubble.MaxSeverity) > severityRank(result.MaxSeverity) {
				result.MaxSeverity = bubble.MaxSeverity
			}
		}
	}

	return result
}

// buildCoverageAggregation はカバレッジデータを構築
// items は Activity から変換された ListItem のリスト
func buildCoverageAggregation(vision core.Vision, objectives []core.ObjectiveEntity, deliverables []core.DeliverableEntity, items []core.ListItem) *CoverageAggregation {
	result := &CoverageAggregation{
		OrphanedTasks: []string{},
		MissingLinks:  []string{},
	}

	// Objective ID → Deliverables マップ
	objDeliverables := make(map[string][]core.DeliverableEntity)
	for _, del := range deliverables {
		objDeliverables[del.ObjectiveID] = append(objDeliverables[del.ObjectiveID], del)
	}

	// Deliverable ID → Items マップ
	delItems := make(map[string][]core.ListItem)
	linkedItemIDs := make(map[string]bool)
	for _, item := range items {
		if item.ParentID != "" {
			delItems[item.ParentID] = append(delItems[item.ParentID], item)
			linkedItemIDs[item.ID] = true
		}
	}

	// 孤立 Activity を検出
	for _, item := range items {
		if !linkedItemIDs[item.ID] {
			result.OrphanedTasks = append(result.OrphanedTasks, item.ID)
		}
	}

	// ルートノードを構築
	root := &CoverageNode{
		ID:       "vision",
		Title:    vision.Title,
		NodeType: "vision",
		Value:    1,
		Children: []*CoverageNode{},
	}
	if root.Title == "" {
		root.Title = "Project"
	}

	coveredCount := 0
	for _, obj := range objectives {
		objNode := &CoverageNode{
			ID:       obj.ID,
			Title:    obj.Title,
			NodeType: "objective",
			Value:    1,
			Children: []*CoverageNode{},
		}

		// Deliverable なしの Objective をマーク
		dels := objDeliverables[obj.ID]
		if len(dels) == 0 {
			objNode.HasIssue = true
			objNode.IssueType = "no_deliverables"
			result.MissingLinks = append(result.MissingLinks, obj.ID)
		} else {
			coveredCount++
		}

		// Deliverable を追加
		for _, del := range dels {
			delNode := &CoverageNode{
				ID:       del.ID,
				Title:    del.Title,
				NodeType: "deliverable",
				Value:    1,
				Children: []*CoverageNode{},
			}

			// Activity なしの Deliverable をマーク
			itemList := delItems[del.ID]
			if len(itemList) == 0 {
				delNode.HasIssue = true
				delNode.IssueType = "no_tasks"
				result.MissingLinks = append(result.MissingLinks, del.ID)
			}

			// Activity を追加
			for _, item := range itemList {
				itemNode := &CoverageNode{
					ID:       item.ID,
					Title:    item.Title,
					NodeType: "task",
					Value:    1,
				}
				delNode.Children = append(delNode.Children, itemNode)
			}

			objNode.Children = append(objNode.Children, delNode)
		}

		root.Children = append(root.Children, objNode)
	}

	result.Root = root

	// カバレッジスコアを計算
	if len(objectives) > 0 {
		result.CoverageScore = (coveredCount * 100) / len(objectives)
	} else {
		result.CoverageScore = 100
	}

	return result
}

// buildResourceAggregation はリソース配分データを構築
// items は Activity から変換された ListItem のリスト
func buildResourceAggregation(objectives []core.ObjectiveEntity, items []core.ListItem) *ResourceAggregation {
	result := &ResourceAggregation{
		Assignees:  []string{},
		Objectives: []string{},
		Matrix:     [][]ResourceCell{},
	}

	// 担当者一覧を収集
	assigneeSet := make(map[string]bool)
	for _, item := range items {
		if item.Assignee != "" {
			assigneeSet[item.Assignee] = true
		}
	}
	for assignee := range assigneeSet {
		result.Assignees = append(result.Assignees, assignee)
	}

	// Objective 一覧
	objIDs := make(map[string]int) // ID → index
	for i, obj := range objectives {
		result.Objectives = append(result.Objectives, obj.Title)
		objIDs[obj.ID] = i
	}

	// 担当者がいない場合は空のマトリクスを返す
	if len(result.Assignees) == 0 || len(objectives) == 0 {
		return result
	}

	// マトリクスを初期化
	result.Matrix = make([][]ResourceCell, len(result.Assignees))
	for i := range result.Matrix {
		result.Matrix[i] = make([]ResourceCell, len(objectives))
	}

	// Activity → Objective のマッピング（簡易版: ParentID から推定）
	// 実際には Deliverable → Objective のリンクを辿る必要がある
	assigneeIdx := make(map[string]int)
	for i, a := range result.Assignees {
		assigneeIdx[a] = i
	}

	for _, item := range items {
		if item.Assignee == "" {
			continue
		}
		aIdx, ok := assigneeIdx[item.Assignee]
		if !ok {
			continue
		}

		// 仮: 最初の Objective に割り当て（実際はリンクを辿る）
		oIdx := 0
		if item.ParentID != "" {
			if idx, ok := objIDs[item.ParentID]; ok {
				oIdx = idx
			}
		}
		if oIdx >= len(objectives) {
			oIdx = 0
		}

		result.Matrix[aIdx][oIdx].TaskCount++
		result.Matrix[aIdx][oIdx].Progress += item.Progress
		if item.Status == core.ItemStatusBlocked {
			result.Matrix[aIdx][oIdx].BlockedCount++
		}
	}

	// 進捗を平均化
	for i := range result.Matrix {
		for j := range result.Matrix[i] {
			if result.Matrix[i][j].TaskCount > 0 {
				result.Matrix[i][j].Progress /= result.Matrix[i][j].TaskCount
			}
		}
	}

	return result
}
