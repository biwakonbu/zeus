package core

import (
	"fmt"
	"time"
)

// ZeusConfig はメイン設定
type ZeusConfig struct {
	Version    string      `yaml:"version"`
	Project    ProjectInfo `yaml:"project"`
	Objectives []Objective `yaml:"objectives"`
	Settings   Settings    `yaml:"settings"`
}

// ProjectInfo はプロジェクト情報
type ProjectInfo struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	StartDate   string `yaml:"start_date"`
}

// Objective は目標
type Objective struct {
	ID       string `yaml:"id"`
	Title    string `yaml:"title"`
	Deadline string `yaml:"deadline"`
	Priority string `yaml:"priority"` // high, medium, low
}

// Settings は設定
type Settings struct {
	AutomationLevel string `yaml:"automation_level"` // auto, notify, approve
	ApprovalMode    string `yaml:"approval_mode"`    // default, strict, loose
	AIProvider      string `yaml:"ai_provider"`      // claude-code, gemini, codex
}

// TaskStatus はタスクステータス
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusBlocked    TaskStatus = "blocked"
)

// ApprovalLevel は承認レベル
type ApprovalLevel string

const (
	ApprovalAuto    ApprovalLevel = "auto"
	ApprovalNotify  ApprovalLevel = "notify"
	ApprovalApprove ApprovalLevel = "approve"
)

// TaskPriority はタスク優先度
type TaskPriority string

const (
	PriorityHigh   TaskPriority = "high"
	PriorityMedium TaskPriority = "medium"
	PriorityLow    TaskPriority = "low"
)

// Task はタスク
type Task struct {
	ID            string        `yaml:"id"`
	Title         string        `yaml:"title"`
	Description   string        `yaml:"description,omitempty"`
	Status        TaskStatus    `yaml:"status"`
	Priority      TaskPriority  `yaml:"priority,omitempty"`
	Assignee      string        `yaml:"assignee,omitempty"`
	EstimateHours float64       `yaml:"estimate_hours,omitempty"`
	ActualHours   float64       `yaml:"actual_hours,omitempty"`
	Dependencies  []string      `yaml:"dependencies"`
	ApprovalLevel ApprovalLevel `yaml:"approval_level"`
	CreatedAt     string        `yaml:"created_at"`
	UpdatedAt     string        `yaml:"updated_at"`

	// Phase 6A: WBS・タイムライン機能用の新規フィールド（全て optional）
	ParentID  string `yaml:"parent_id,omitempty"`  // 親タスクID
	StartDate string `yaml:"start_date,omitempty"` // 開始日（ISO8601）
	DueDate   string `yaml:"due_date,omitempty"`   // 期限日（ISO8601）
	Progress  int    `yaml:"progress,omitempty"`   // 進捗率（0-100）
	WBSCode   string `yaml:"wbs_code,omitempty"`   // WBS番号（例: "1.2.3"）
}

// TaskStore はタスクストア
type TaskStore struct {
	Tasks []Task `yaml:"tasks"`
}

// HealthStatus は健全性ステータス
type HealthStatus string

const (
	HealthGood    HealthStatus = "good"
	HealthFair    HealthStatus = "fair"
	HealthPoor    HealthStatus = "poor"
	HealthUnknown HealthStatus = "unknown"
)

// TaskStats はタスク統計
type TaskStats struct {
	TotalTasks int `yaml:"total_tasks"`
	Completed  int `yaml:"completed"`
	InProgress int `yaml:"in_progress"`
	Pending    int `yaml:"pending"`
}

// ProjectState はプロジェクト状態
type ProjectState struct {
	Timestamp string       `yaml:"timestamp"`
	Summary   TaskStats    `yaml:"summary"`
	Health    HealthStatus `yaml:"health"`
	Risks     []string     `yaml:"risks"`
}

// Snapshot はスナップショット
type Snapshot struct {
	Timestamp string       `yaml:"timestamp"`
	Label     string       `yaml:"label,omitempty"`
	State     ProjectState `yaml:"state"`
}

// InitResult は初期化結果
type InitResult struct {
	Success    bool
	ZeusPath   string
	ClaudePath string
}

// StatusResult はステータス結果
type StatusResult struct {
	Project          ProjectInfo
	State            ProjectState
	PendingApprovals int
}

// AddResult は追加結果
type AddResult struct {
	Success       bool
	ID            string
	Entity        string
	NeedsApproval bool   // 承認が必要な場合 true
	ApprovalID    string // 承認待ち ID（NeedsApproval が true の場合）
}

// ListResult は一覧結果
type ListResult struct {
	Entity string
	Items  []Task
	Total  int
}

// Now は現在時刻を ISO8601 形式で返す
func Now() string {
	return time.Now().Format(time.RFC3339)
}

// Today は今日の日付を返す
func Today() string {
	return time.Now().Format("2006-01-02")
}

// SuggestionType は提案タイプ
type SuggestionType string

const (
	SuggestionNewTask        SuggestionType = "new_task"
	SuggestionPriorityChange SuggestionType = "priority_change"
	SuggestionDependency     SuggestionType = "dependency"
	SuggestionRiskMitigation SuggestionType = "risk_mitigation"
)

// SuggestionImpact は提案の影響度
type SuggestionImpact string

const (
	ImpactHigh   SuggestionImpact = "high"
	ImpactMedium SuggestionImpact = "medium"
	ImpactLow    SuggestionImpact = "low"
)

// SuggestionStatus は提案ステータス
type SuggestionStatus string

const (
	SuggestionPending  SuggestionStatus = "pending"
	SuggestionApplied  SuggestionStatus = "applied"
	SuggestionRejected SuggestionStatus = "rejected"
)

// Suggestion はAI提案
type Suggestion struct {
	ID          string           `yaml:"id"`
	Type        SuggestionType   `yaml:"type"`
	Description string           `yaml:"description"`
	Rationale   string           `yaml:"rationale"`
	Impact      SuggestionImpact `yaml:"impact"`
	Status      SuggestionStatus `yaml:"status"`
	CreatedAt   string           `yaml:"created_at"`
	UpdatedAt   string           `yaml:"updated_at,omitempty"`
	// タイプ固有のデータ
	TargetTaskID string   `yaml:"target_task_id,omitempty"` // priority_change, dependency用
	NewPriority  string   `yaml:"new_priority,omitempty"`   // priority_change用
	Dependencies []string `yaml:"dependencies,omitempty"`   // dependency用
	TaskData     *Task    `yaml:"task_data,omitempty"`      // new_task用
}

// SuggestionStore は提案ストア
type SuggestionStore struct {
	Suggestions []Suggestion `yaml:"suggestions"`
}

// ApplyResult は提案適用結果
type ApplyResult struct {
	Applied       int
	Skipped       int
	Failed        int
	AppliedIDs    []string
	FailedIDs     []string
	CreatedTaskID string
}

// ExplainResult は説明結果
type ExplainResult struct {
	EntityID    string            // 対象エンティティID
	EntityType  string            // エンティティタイプ (project, task, etc.)
	Summary     string            // 要約説明
	Details     string            // 詳細説明
	Context     map[string]string // コンテキスト情報
	Suggestions []string          // 改善提案
}

// Validate は Task の妥当性を検証
func (t *Task) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task ID is required")
	}
	if t.Title == "" {
		return fmt.Errorf("task title is required")
	}
	if t.Status == "" {
		return fmt.Errorf("task status is required")
	}
	if t.EstimateHours < 0 {
		return fmt.Errorf("estimate_hours must be non-negative, got %f", t.EstimateHours)
	}
	if t.ActualHours < 0 {
		return fmt.Errorf("actual_hours must be non-negative, got %f", t.ActualHours)
	}
	if t.ApprovalLevel != "" &&
		t.ApprovalLevel != ApprovalAuto &&
		t.ApprovalLevel != ApprovalNotify &&
		t.ApprovalLevel != ApprovalApprove {
		return fmt.Errorf("invalid approval level: %s", t.ApprovalLevel)
	}
	// Progress は 0-100 の範囲
	if t.Progress < 0 || t.Progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100, got %d", t.Progress)
	}
	// 自己参照の禁止
	if t.ParentID != "" && t.ParentID == t.ID {
		return fmt.Errorf("task cannot be its own parent")
	}
	return nil
}

// Validate は Suggestion の妥当性を検証
func (s *Suggestion) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("suggestion ID is required")
	}
	if s.Description == "" {
		return fmt.Errorf("suggestion description is required")
	}
	if s.Impact == "" {
		return fmt.Errorf("suggestion impact is required")
	}

	switch s.Type {
	case SuggestionNewTask:
		if s.TaskData == nil {
			return fmt.Errorf("new_task suggestion must have TaskData")
		}
		if err := s.TaskData.Validate(); err != nil {
			return fmt.Errorf("invalid task data: %w", err)
		}
	case SuggestionPriorityChange:
		if s.TargetTaskID == "" {
			return fmt.Errorf("priority_change suggestion must have TargetTaskID")
		}
		if s.NewPriority == "" {
			return fmt.Errorf("priority_change suggestion must have NewPriority")
		}
	case SuggestionDependency:
		if s.TargetTaskID == "" {
			return fmt.Errorf("dependency suggestion must have TargetTaskID")
		}
		if len(s.Dependencies) == 0 {
			return fmt.Errorf("dependency suggestion must have at least one dependency")
		}
	case SuggestionRiskMitigation:
		// リスク対策は追加検証不要
	default:
		return fmt.Errorf("unknown suggestion type: %s", s.Type)
	}

	return nil
}

// ============================================================
// 10 概念モデル型定義 (Phase 1: Vision, Objective, Deliverable)
// ============================================================

// Metadata は 10 概念モデルの共通メタデータ
type Metadata struct {
	CreatedAt string   `yaml:"created_at"`
	UpdatedAt string   `yaml:"updated_at,omitempty"`
	Owner     string   `yaml:"owner,omitempty"`
	Tags      []string `yaml:"tags,omitempty"`
}

// VisionStatus は Vision の状態
type VisionStatus string

const (
	VisionStatusDraft    VisionStatus = "draft"
	VisionStatusActive   VisionStatus = "active"
	VisionStatusArchived VisionStatus = "archived"
)

// Vision は 10 概念モデルのビジョン
// 単一ファイル (vision.yaml) で管理
type Vision struct {
	ID              string       `yaml:"id"`
	Title           string       `yaml:"title"`
	Statement       string       `yaml:"statement"`
	SuccessCriteria []string     `yaml:"success_criteria,omitempty"`
	Status          VisionStatus `yaml:"status"`
	Metadata        Metadata     `yaml:"metadata"`
}

// ObjectiveStatus は Objective の状態
type ObjectiveStatus string

const (
	ObjectiveStatusNotStarted ObjectiveStatus = "not_started"
	ObjectiveStatusInProgress ObjectiveStatus = "in_progress"
	ObjectiveStatusCompleted  ObjectiveStatus = "completed"
	ObjectiveStatusOnHold     ObjectiveStatus = "on_hold"
	ObjectiveStatusCancelled  ObjectiveStatus = "cancelled"
)

// ObjectiveEntity は 10 概念モデルの目標（個別ファイル管理）
// objectives/obj-NNN.yaml で管理
// 注: ZeusConfig 内の Objective とは別の構造
type ObjectiveEntity struct {
	ID          string          `yaml:"id"`
	Title       string          `yaml:"title"`
	Description string          `yaml:"description,omitempty"`
	Status      ObjectiveStatus `yaml:"status"`
	Progress    int             `yaml:"progress"`
	Owner       string          `yaml:"owner,omitempty"`
	ParentID    string          `yaml:"parent_id,omitempty"` // 親 Objective ID（階層化用）
	WBSCode     string          `yaml:"wbs_code,omitempty"`
	StartDate   string          `yaml:"start_date,omitempty"`
	DueDate     string          `yaml:"due_date,omitempty"`
	Tags        []string        `yaml:"tags,omitempty"`
	Metadata    Metadata        `yaml:"metadata"`
}

// DeliverableStatus は Deliverable の状態
type DeliverableStatus string

const (
	DeliverableStatusPlanned    DeliverableStatus = "planned"
	DeliverableStatusInProgress DeliverableStatus = "in_progress"
	DeliverableStatusReview     DeliverableStatus = "review"
	DeliverableStatusCompleted  DeliverableStatus = "completed"
	DeliverableStatusCancelled  DeliverableStatus = "cancelled"
)

// DeliverableFormat は成果物のフォーマット
type DeliverableFormat string

const (
	DeliverableFormatDocument    DeliverableFormat = "document"
	DeliverableFormatCode        DeliverableFormat = "code"
	DeliverableFormatData        DeliverableFormat = "data"
	DeliverableFormatDesign      DeliverableFormat = "design"
	DeliverableFormatPresentation DeliverableFormat = "presentation"
	DeliverableFormatOther       DeliverableFormat = "other"
)

// DeliverableEntity は 10 概念モデルの成果物
// deliverables/del-NNN.yaml で管理
type DeliverableEntity struct {
	ID                 string            `yaml:"id"`
	Title              string            `yaml:"title"`
	Description        string            `yaml:"description,omitempty"`
	ObjectiveID        string            `yaml:"objective_id"` // 紐づく Objective
	Format             DeliverableFormat `yaml:"format"`
	AcceptanceCriteria []string          `yaml:"acceptance_criteria,omitempty"`
	Status             DeliverableStatus `yaml:"status"`
	Progress           int               `yaml:"progress"`
	Metadata           Metadata          `yaml:"metadata"`
}

// Validate は Vision の妥当性を検証
func (v *Vision) Validate() error {
	if v.ID == "" {
		return fmt.Errorf("vision ID is required")
	}
	if err := ValidateID("vision", v.ID); err != nil {
		return err
	}
	if v.Title == "" {
		return fmt.Errorf("vision title is required")
	}
	if v.Statement == "" {
		return fmt.Errorf("vision statement is required")
	}
	if v.Status == "" {
		v.Status = VisionStatusDraft
	}
	switch v.Status {
	case VisionStatusDraft, VisionStatusActive, VisionStatusArchived:
		// 有効
	default:
		return fmt.Errorf("invalid vision status: %s", v.Status)
	}
	return nil
}

// Validate は ObjectiveEntity の妥当性を検証
func (o *ObjectiveEntity) Validate() error {
	if o.ID == "" {
		return fmt.Errorf("objective ID is required")
	}
	if err := ValidateID("objective", o.ID); err != nil {
		return err
	}
	if o.Title == "" {
		return fmt.Errorf("objective title is required")
	}
	if o.Status == "" {
		o.Status = ObjectiveStatusNotStarted
	}
	switch o.Status {
	case ObjectiveStatusNotStarted, ObjectiveStatusInProgress, ObjectiveStatusCompleted,
		ObjectiveStatusOnHold, ObjectiveStatusCancelled:
		// 有効
	default:
		return fmt.Errorf("invalid objective status: %s", o.Status)
	}
	if o.Progress < 0 || o.Progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100, got %d", o.Progress)
	}
	// 自己参照の禁止
	if o.ParentID != "" && o.ParentID == o.ID {
		return fmt.Errorf("objective cannot be its own parent")
	}
	return nil
}

// Validate は DeliverableEntity の妥当性を検証
func (d *DeliverableEntity) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("deliverable ID is required")
	}
	if err := ValidateID("deliverable", d.ID); err != nil {
		return err
	}
	if d.Title == "" {
		return fmt.Errorf("deliverable title is required")
	}
	if d.ObjectiveID == "" {
		return fmt.Errorf("deliverable objective_id is required")
	}
	if d.Format == "" {
		d.Format = DeliverableFormatOther
	}
	switch d.Format {
	case DeliverableFormatDocument, DeliverableFormatCode, DeliverableFormatData,
		DeliverableFormatDesign, DeliverableFormatPresentation, DeliverableFormatOther:
		// 有効
	default:
		return fmt.Errorf("invalid deliverable format: %s", d.Format)
	}
	if d.Status == "" {
		d.Status = DeliverableStatusPlanned
	}
	switch d.Status {
	case DeliverableStatusPlanned, DeliverableStatusInProgress, DeliverableStatusReview,
		DeliverableStatusCompleted, DeliverableStatusCancelled:
		// 有効
	default:
		return fmt.Errorf("invalid deliverable status: %s", d.Status)
	}
	if d.Progress < 0 || d.Progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100, got %d", d.Progress)
	}
	return nil
}

// GetID は Entity インターフェースを実装（Vision）
func (v *Vision) GetID() string { return v.ID }

// GetTitle は Entity インターフェースを実装（Vision）
func (v *Vision) GetTitle() string { return v.Title }

// GetID は Entity インターフェースを実装（ObjectiveEntity）
func (o *ObjectiveEntity) GetID() string { return o.ID }

// GetTitle は Entity インターフェースを実装（ObjectiveEntity）
func (o *ObjectiveEntity) GetTitle() string { return o.Title }

// GetID は Entity インターフェースを実装（DeliverableEntity）
func (d *DeliverableEntity) GetID() string { return d.ID }

// GetTitle は Entity インターフェースを実装（DeliverableEntity）
func (d *DeliverableEntity) GetTitle() string { return d.Title }
