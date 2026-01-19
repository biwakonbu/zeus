package analysis

import (
	"context"
	"time"
)

// StaleType は陳腐化の種類
type StaleType string

const (
	StaleTypeCompletedOld StaleType = "completed_old"
	StaleTypeOrphaned     StaleType = "orphaned"
	StaleTypeBlockedLong  StaleType = "blocked_long"
	StaleTypeNoProgress   StaleType = "no_progress"
)

// StaleRecommendation は推奨アクション
type StaleRecommendation string

const (
	StaleRecommendArchive StaleRecommendation = "archive"
	StaleRecommendReview  StaleRecommendation = "review"
	StaleRecommendDelete  StaleRecommendation = "delete"
)

// StaleEntity は陳腐化したエンティティ
type StaleEntity struct {
	Type           StaleType           `json:"type"`
	EntityID       string              `json:"entity_id"`
	EntityTitle    string              `json:"entity_title"`
	EntityType     string              `json:"entity_type"` // task, objective, deliverable
	Recommendation StaleRecommendation `json:"recommendation"`
	Message        string              `json:"message"`
	DaysStale      int                 `json:"days_stale"`
}

// StaleAnalysis は陳腐化分析結果
type StaleAnalysis struct {
	StaleEntities []StaleEntity `json:"stale_entities"`
	TotalStale    int           `json:"total_stale"`
	ArchiveCount  int           `json:"archive_count"`
	ReviewCount   int           `json:"review_count"`
	DeleteCount   int           `json:"delete_count"`
}

// StaleAnalyzerConfig は陳腐化分析の設定
type StaleAnalyzerConfig struct {
	CompletedStaleDays int // 完了後何日で陳腐化とみなすか（デフォルト: 30）
	BlockedStaleDays   int // ブロック状態が何日で陳腐化とみなすか（デフォルト: 14）
	NoProgressDays     int // 進捗がない状態が何日で陳腐化とみなすか（デフォルト: 21）
}

// DefaultStaleConfig はデフォルト設定
var DefaultStaleConfig = StaleAnalyzerConfig{
	CompletedStaleDays: 30,
	BlockedStaleDays:   14,
	NoProgressDays:     21,
}

// StaleAnalyzer は陳腐化分析を行う
type StaleAnalyzer struct {
	tasks        []TaskInfo
	objectives   []ObjectiveInfo
	deliverables []DeliverableInfo
	config       StaleAnalyzerConfig
	now          time.Time
}

// NewStaleAnalyzer は新しい StaleAnalyzer を作成
func NewStaleAnalyzer(
	tasks []TaskInfo,
	objectives []ObjectiveInfo,
	deliverables []DeliverableInfo,
	config *StaleAnalyzerConfig,
) *StaleAnalyzer {
	cfg := DefaultStaleConfig
	if config != nil {
		cfg = *config
	}
	return &StaleAnalyzer{
		tasks:        tasks,
		objectives:   objectives,
		deliverables: deliverables,
		config:       cfg,
		now:          time.Now(),
	}
}

// Analyze は陳腐化分析を実行
func (s *StaleAnalyzer) Analyze(ctx context.Context) (*StaleAnalysis, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &StaleAnalysis{
		StaleEntities: []StaleEntity{},
	}

	// 参照マップを作成（孤立検出用）
	referenced := make(map[string]bool)
	for _, task := range s.tasks {
		for _, dep := range task.Dependencies {
			referenced[dep] = true
		}
		if task.ParentID != "" {
			referenced[task.ParentID] = true
		}
	}
	for _, del := range s.deliverables {
		if del.ObjectiveID != "" {
			referenced[del.ObjectiveID] = true
		}
	}

	// タスクの陳腐化チェック
	for _, task := range s.tasks {
		stale := s.checkTaskStale(task, referenced)
		if stale != nil {
			result.StaleEntities = append(result.StaleEntities, *stale)
			s.countRecommendation(result, stale.Recommendation)
		}
	}

	// Objective の陳腐化チェック
	for _, obj := range s.objectives {
		stale := s.checkObjectiveStale(obj, referenced)
		if stale != nil {
			result.StaleEntities = append(result.StaleEntities, *stale)
			s.countRecommendation(result, stale.Recommendation)
		}
	}

	// Deliverable の陳腐化チェック
	for _, del := range s.deliverables {
		stale := s.checkDeliverableStale(del, referenced)
		if stale != nil {
			result.StaleEntities = append(result.StaleEntities, *stale)
			s.countRecommendation(result, stale.Recommendation)
		}
	}

	result.TotalStale = len(result.StaleEntities)
	return result, nil
}

// checkTaskStale はタスクの陳腐化をチェック
func (s *StaleAnalyzer) checkTaskStale(task TaskInfo, referenced map[string]bool) *StaleEntity {
	// 1. 完了後 30 日以上経過
	if task.Status == TaskStatusCompleted {
		completedAt := s.parseDate(task.CompletedAt)
		if completedAt != nil {
			days := int(s.now.Sub(*completedAt).Hours() / 24)
			if days >= s.config.CompletedStaleDays {
				return &StaleEntity{
					Type:           StaleTypeCompletedOld,
					EntityID:       task.ID,
					EntityTitle:    task.Title,
					EntityType:     "task",
					Recommendation: StaleRecommendArchive,
					Message:        "完了後 " + itoa(days) + " 日が経過しています",
					DaysStale:      days,
				}
			}
		}
	}

	// 2. ブロック状態が 14 日以上継続
	if task.Status == TaskStatusBlocked {
		// ブロック開始日は UpdatedAt を使用（暫定）
		blockedAt := s.parseDate(task.UpdatedAt)
		if blockedAt != nil {
			days := int(s.now.Sub(*blockedAt).Hours() / 24)
			if days >= s.config.BlockedStaleDays {
				return &StaleEntity{
					Type:           StaleTypeBlockedLong,
					EntityID:       task.ID,
					EntityTitle:    task.Title,
					EntityType:     "task",
					Recommendation: StaleRecommendReview,
					Message:        "ブロック状態が " + itoa(days) + " 日継続しています",
					DaysStale:      days,
				}
			}
		}
	}

	// 3. 孤立タスク（誰からも参照されていない完了済みタスク）
	if task.Status == TaskStatusCompleted && !referenced[task.ID] && task.ParentID == "" && len(task.Dependencies) == 0 {
		return &StaleEntity{
			Type:           StaleTypeOrphaned,
			EntityID:       task.ID,
			EntityTitle:    task.Title,
			EntityType:     "task",
			Recommendation: StaleRecommendReview,
			Message:        "孤立したタスクです（参照なし）",
			DaysStale:      0,
		}
	}

	return nil
}

// checkObjectiveStale は Objective の陳腐化をチェック
func (s *StaleAnalyzer) checkObjectiveStale(obj ObjectiveInfo, referenced map[string]bool) *StaleEntity {
	// 完了した Objective で、更新から 30 日以上経過
	if obj.Status == "completed" {
		updatedAt := s.parseDate(obj.UpdatedAt)
		if updatedAt != nil {
			days := int(s.now.Sub(*updatedAt).Hours() / 24)
			if days >= s.config.CompletedStaleDays {
				return &StaleEntity{
					Type:           StaleTypeCompletedOld,
					EntityID:       obj.ID,
					EntityTitle:    obj.Title,
					EntityType:     "objective",
					Recommendation: StaleRecommendArchive,
					Message:        "完了後 " + itoa(days) + " 日が経過しています",
					DaysStale:      days,
				}
			}
		}
	}

	return nil
}

// checkDeliverableStale は Deliverable の陳腐化をチェック
func (s *StaleAnalyzer) checkDeliverableStale(del DeliverableInfo, referenced map[string]bool) *StaleEntity {
	// 完了した Deliverable で、更新から 30 日以上経過
	if del.Status == "delivered" || del.Status == "approved" {
		updatedAt := s.parseDate(del.UpdatedAt)
		if updatedAt != nil {
			days := int(s.now.Sub(*updatedAt).Hours() / 24)
			if days >= s.config.CompletedStaleDays {
				return &StaleEntity{
					Type:           StaleTypeCompletedOld,
					EntityID:       del.ID,
					EntityTitle:    del.Title,
					EntityType:     "deliverable",
					Recommendation: StaleRecommendArchive,
					Message:        "完了後 " + itoa(days) + " 日が経過しています",
					DaysStale:      days,
				}
			}
		}
	}

	return nil
}

// parseDate は日付文字列をパース
func (s *StaleAnalyzer) parseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	// 複数のフォーマットを試す
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t
		}
	}
	return nil
}

// countRecommendation は推奨アクションをカウント
func (s *StaleAnalyzer) countRecommendation(result *StaleAnalysis, rec StaleRecommendation) {
	switch rec {
	case StaleRecommendArchive:
		result.ArchiveCount++
	case StaleRecommendReview:
		result.ReviewCount++
	case StaleRecommendDelete:
		result.DeleteCount++
	}
}
