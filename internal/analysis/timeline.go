package analysis

import (
	"context"
	"sort"
	"time"
)

// TimelineItem はタイムライン上のアイテム
type TimelineItem struct {
	TaskID           string   `json:"task_id"`
	Title            string   `json:"title"`
	StartDate        string   `json:"start_date"`
	EndDate          string   `json:"end_date"`
	Progress         int      `json:"progress"`
	Status           string   `json:"status"`
	Priority         string   `json:"priority"`
	Assignee         string   `json:"assignee"`
	IsOnCriticalPath bool     `json:"is_on_critical_path"`
	Slack            int      `json:"slack"` // 余裕日数
	Dependencies     []string `json:"dependencies"`
}

// Timeline はプロジェクト全体のタイムライン
type Timeline struct {
	Items         []TimelineItem `json:"items"`
	CriticalPath  []string       `json:"critical_path"`
	ProjectStart  string         `json:"project_start"`
	ProjectEnd    string         `json:"project_end"`
	TotalDuration int            `json:"total_duration"` // 日数
	Stats         TimelineStats  `json:"stats"`
}

// TimelineStats はタイムライン統計
type TimelineStats struct {
	TotalTasks      int     `json:"total_tasks"`
	TasksWithDates  int     `json:"tasks_with_dates"`
	OnCriticalPath  int     `json:"on_critical_path"`
	AverageSlack    float64 `json:"average_slack"`
	OverdueTasks    int     `json:"overdue_tasks"`
	CompletedOnTime int     `json:"completed_on_time"`
}

// TimelineBuilder はタイムラインを構築するビルダー
type TimelineBuilder struct {
	tasks map[string]*TaskInfo
}

// NewTimelineBuilder は新しい TimelineBuilder を作成
func NewTimelineBuilder(tasks []TaskInfo) *TimelineBuilder {
	taskMap := make(map[string]*TaskInfo)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}
	return &TimelineBuilder{tasks: taskMap}
}

// Build はタイムラインを構築
func (tb *TimelineBuilder) Build(ctx context.Context) (*Timeline, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 日付が設定されているタスクを抽出
	var itemsWithDates []TimelineItem
	var projectStart, projectEnd time.Time
	firstDate := true

	today := time.Now().Truncate(24 * time.Hour)
	overdue := 0
	completedOnTime := 0

	for _, task := range tb.tasks {
		item := TimelineItem{
			TaskID:       task.ID,
			Title:        task.Title,
			StartDate:    task.StartDate,
			EndDate:      task.DueDate,
			Progress:     task.Progress,
			Status:       task.Status,
			Priority:     task.Priority,
			Assignee:     task.Assignee,
			Dependencies: task.Dependencies,
		}

		// 日付のパース
		var start, end time.Time
		if task.StartDate != "" {
			if t, err := time.Parse("2006-01-02", task.StartDate); err == nil {
				start = t
			}
		}
		if task.DueDate != "" {
			if t, err := time.Parse("2006-01-02", task.DueDate); err == nil {
				end = t
			}
		}

		// 期限切れチェック
		if task.Status != TaskStatusCompleted && !end.IsZero() && end.Before(today) {
			overdue++
		}

		// 期限内完了チェック
		if task.Status == TaskStatusCompleted && !end.IsZero() {
			completedOnTime++
		}

		// 有効な日付範囲を持つタスクのみ追加
		if !start.IsZero() || !end.IsZero() {
			itemsWithDates = append(itemsWithDates, item)

			// プロジェクト期間の計算
			if !start.IsZero() {
				if firstDate || start.Before(projectStart) {
					projectStart = start
				}
			}
			if !end.IsZero() {
				if firstDate || end.After(projectEnd) {
					projectEnd = end
				}
			}
			firstDate = false
		}
	}

	// 開始日でソート
	sort.Slice(itemsWithDates, func(i, j int) bool {
		return itemsWithDates[i].StartDate < itemsWithDates[j].StartDate
	})

	// クリティカルパスを計算
	criticalPath, slack := tb.calculateCriticalPath(ctx, itemsWithDates)

	// スラックを各アイテムに設定
	for i := range itemsWithDates {
		if s, ok := slack[itemsWithDates[i].TaskID]; ok {
			itemsWithDates[i].Slack = s
		}
		// クリティカルパスに含まれるか判定
		for _, cpID := range criticalPath {
			if itemsWithDates[i].TaskID == cpID {
				itemsWithDates[i].IsOnCriticalPath = true
				break
			}
		}
	}

	// 統計計算
	var totalSlack float64
	slackCount := 0
	for _, item := range itemsWithDates {
		if item.Slack >= 0 {
			totalSlack += float64(item.Slack)
			slackCount++
		}
	}
	avgSlack := 0.0
	if slackCount > 0 {
		avgSlack = totalSlack / float64(slackCount)
	}

	// プロジェクト期間の計算
	var totalDuration int
	if !projectStart.IsZero() && !projectEnd.IsZero() {
		totalDuration = int(projectEnd.Sub(projectStart).Hours() / 24)
	}

	// 日付文字列の設定
	var projectStartStr, projectEndStr string
	if !projectStart.IsZero() {
		projectStartStr = projectStart.Format("2006-01-02")
	}
	if !projectEnd.IsZero() {
		projectEndStr = projectEnd.Format("2006-01-02")
	}

	return &Timeline{
		Items:         itemsWithDates,
		CriticalPath:  criticalPath,
		ProjectStart:  projectStartStr,
		ProjectEnd:    projectEndStr,
		TotalDuration: totalDuration,
		Stats: TimelineStats{
			TotalTasks:      len(tb.tasks),
			TasksWithDates:  len(itemsWithDates),
			OnCriticalPath:  len(criticalPath),
			AverageSlack:    avgSlack,
			OverdueTasks:    overdue,
			CompletedOnTime: completedOnTime,
		},
	}, nil
}

// calculateCriticalPath はクリティカルパスを計算
// CPM (Critical Path Method) アルゴリズムを使用
func (tb *TimelineBuilder) calculateCriticalPath(_ context.Context, items []TimelineItem) ([]string, map[string]int) {
	if len(items) == 0 {
		return []string{}, map[string]int{}
	}

	// タスクマップを作成
	taskMap := make(map[string]*TimelineItem)
	for i := range items {
		taskMap[items[i].TaskID] = &items[i]
	}

	// 各タスクの期間を計算
	duration := make(map[string]int)
	for _, item := range items {
		var start, end time.Time
		if item.StartDate != "" {
			start, _ = time.Parse("2006-01-02", item.StartDate)
		}
		if item.EndDate != "" {
			end, _ = time.Parse("2006-01-02", item.EndDate)
		}
		if !start.IsZero() && !end.IsZero() {
			duration[item.TaskID] = int(end.Sub(start).Hours() / 24)
		} else {
			duration[item.TaskID] = 1 // デフォルト 1 日
		}
	}

	// フォワードパス: 最早開始時刻（ES）と最早終了時刻（EF）を計算
	es := make(map[string]int) // Early Start
	ef := make(map[string]int) // Early Finish

	// トポロジカルソート順で処理（依存関係を考慮）
	visited := make(map[string]bool)
	var order []string
	var visit func(id string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		if item, ok := taskMap[id]; ok {
			for _, depID := range item.Dependencies {
				visit(depID)
			}
		}
		order = append(order, id)
	}
	for _, item := range items {
		visit(item.TaskID)
	}

	// フォワードパス
	for _, id := range order {
		item, ok := taskMap[id]
		if !ok {
			continue
		}
		maxEF := 0
		for _, depID := range item.Dependencies {
			if ef[depID] > maxEF {
				maxEF = ef[depID]
			}
		}
		es[id] = maxEF
		ef[id] = maxEF + duration[id]
	}

	// プロジェクト最終日を取得
	projectEnd := 0
	for _, e := range ef {
		if e > projectEnd {
			projectEnd = e
		}
	}

	// バックワードパス: 最遅開始時刻（LS）と最遅終了時刻（LF）を計算
	ls := make(map[string]int) // Late Start
	lf := make(map[string]int) // Late Finish

	// 後続タスクを構築
	successors := make(map[string][]string)
	for _, item := range items {
		for _, depID := range item.Dependencies {
			successors[depID] = append(successors[depID], item.TaskID)
		}
	}

	// 逆順で処理
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		if len(successors[id]) == 0 {
			lf[id] = projectEnd
		} else {
			minLS := projectEnd
			for _, succID := range successors[id] {
				if ls[succID] < minLS {
					minLS = ls[succID]
				}
			}
			lf[id] = minLS
		}
		ls[id] = lf[id] - duration[id]
	}

	// スラック（余裕時間）を計算
	slack := make(map[string]int)
	for id := range taskMap {
		slack[id] = ls[id] - es[id]
	}

	// クリティカルパス（スラック = 0 のタスク）を抽出
	var criticalPath []string
	for _, id := range order {
		if slack[id] == 0 {
			criticalPath = append(criticalPath, id)
		}
	}

	return criticalPath, slack
}
