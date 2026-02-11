package core

import (
	"fmt"
	"strings"
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

// ItemStatus はリスト項目のステータス
type ItemStatus string

const (
	ItemStatusPending    ItemStatus = "pending"
	ItemStatusInProgress ItemStatus = "in_progress"
	ItemStatusCompleted  ItemStatus = "completed"
	ItemStatusBlocked    ItemStatus = "blocked"
)

// ApprovalLevel は承認レベル
type ApprovalLevel string

const (
	ApprovalAuto    ApprovalLevel = "auto"
	ApprovalNotify  ApprovalLevel = "notify"
	ApprovalApprove ApprovalLevel = "approve"
)

// ItemPriority はリスト項目の優先度
type ItemPriority string

const (
	PriorityHigh   ItemPriority = "high"
	PriorityMedium ItemPriority = "medium"
	PriorityLow    ItemPriority = "low"
)

// ListItem は一覧表示用の汎用項目
type ListItem struct {
	ID            string        `yaml:"id"`
	Title         string        `yaml:"title"`
	Description   string        `yaml:"description,omitempty"`
	Status        ItemStatus    `yaml:"status"`
	Priority      ItemPriority  `yaml:"priority,omitempty"`
	Assignee      string        `yaml:"assignee,omitempty"`
	Dependencies  []string      `yaml:"dependencies"`
	ApprovalLevel ApprovalLevel `yaml:"approval_level"`
	CreatedAt     string        `yaml:"created_at"`
	UpdatedAt     string        `yaml:"updated_at"`
	ParentID      string        `yaml:"parent_id,omitempty"`
}

// HealthStatus は健全性ステータス
type HealthStatus string

const (
	HealthGood    HealthStatus = "good"
	HealthFair    HealthStatus = "fair"
	HealthPoor    HealthStatus = "poor"
	HealthUnknown HealthStatus = "unknown"
)

// SummaryStats はサマリー統計（Activity 統計）
type SummaryStats struct {
	TotalActivities int `yaml:"total_activities"`
	Completed       int `yaml:"completed"`
	InProgress      int `yaml:"in_progress"`
	Pending         int `yaml:"pending"`
}

// ProjectState はプロジェクト状態
type ProjectState struct {
	Timestamp string       `yaml:"timestamp"`
	Summary   SummaryStats `yaml:"summary"`
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
	Items  []ListItem
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
	// 注意: TargetTaskID は後方互換性のために残しているが、Activity ID を指定する
	TargetTaskID string          `yaml:"target_task_id,omitempty"` // priority_change, dependency用（Activity ID を指定）
	NewPriority  string          `yaml:"new_priority,omitempty"`   // priority_change用
	Dependencies []string        `yaml:"dependencies,omitempty"`   // dependency用
	TaskData     *ListItem       `yaml:"task_data,omitempty"`      // new_task用（非推奨: ActivityData を使用）
	ActivityData *ActivityEntity `yaml:"activity_data,omitempty"`  // new_task用（推奨）
}

// SuggestionStore は提案ストア
type SuggestionStore struct {
	Suggestions []Suggestion `yaml:"suggestions"`
}

// ApplyResult は提案適用結果
type ApplyResult struct {
	Applied           int
	Skipped           int
	Failed            int
	AppliedIDs        []string
	FailedIDs         []string
	CreatedTaskID     string // 非推奨: CreatedActivityID を使用
	CreatedActivityID string // 作成された Activity の ID
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

// Validate は ListItem の妥当性を検証
func (t *ListItem) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("item ID is required")
	}
	if t.Title == "" {
		return fmt.Errorf("item title is required")
	}
	if t.Status == "" {
		return fmt.Errorf("item status is required")
	}
	if t.ApprovalLevel != "" &&
		t.ApprovalLevel != ApprovalAuto &&
		t.ApprovalLevel != ApprovalNotify &&
		t.ApprovalLevel != ApprovalApprove {
		return fmt.Errorf("invalid approval level: %s", t.ApprovalLevel)
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
		// ActivityData を優先、なければ TaskData をチェック（後方互換性）
		if s.ActivityData != nil {
			if err := s.ActivityData.Validate(); err != nil {
				return fmt.Errorf("invalid activity data: %w", err)
			}
		} else if s.TaskData != nil {
			if err := s.TaskData.Validate(); err != nil {
				return fmt.Errorf("invalid task data: %w", err)
			}
		} else {
			return fmt.Errorf("new_task suggestion must have ActivityData or TaskData")
		}
	case SuggestionPriorityChange:
		// TargetTaskID は Activity ID を指定（名前は後方互換性のため維持）
		if s.TargetTaskID == "" {
			return fmt.Errorf("priority_change suggestion must have TargetTaskID (Activity ID)")
		}
		if s.NewPriority == "" {
			return fmt.Errorf("priority_change suggestion must have NewPriority")
		}
	case SuggestionDependency:
		// TargetTaskID は Activity ID を指定（名前は後方互換性のため維持）
		if s.TargetTaskID == "" {
			return fmt.Errorf("dependency suggestion must have TargetTaskID (Activity ID)")
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
// 10 概念モデル型定義 (Phase 1: Vision, Objective)
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
	Goals       []string        `yaml:"goals,omitempty"`
	Status      ObjectiveStatus `yaml:"status"`
	Owner       string          `yaml:"owner,omitempty"`
	Tags        []string        `yaml:"tags,omitempty"`
	Metadata    Metadata        `yaml:"metadata"`
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
	// Goals の空文字列要素を除外
	var cleanGoals []string
	for _, g := range o.Goals {
		trimmed := strings.TrimSpace(g)
		if trimmed != "" {
			cleanGoals = append(cleanGoals, trimmed)
		}
	}
	o.Goals = cleanGoals
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

// ============================================================
// 10 概念モデル型定義 (Phase 2: Consideration, Decision, Problem, Risk, Assumption)
// ============================================================

// === Consideration ===

// ConsiderationStatus は Consideration の状態
type ConsiderationStatus string

const (
	ConsiderationStatusOpen     ConsiderationStatus = "open"
	ConsiderationStatusDecided  ConsiderationStatus = "decided"
	ConsiderationStatusDeferred ConsiderationStatus = "deferred"
)

// ConsiderationOption は検討事項の選択肢
type ConsiderationOption struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description,omitempty"`
	Pros        []string `yaml:"pros,omitempty"`
	Cons        []string `yaml:"cons,omitempty"`
}

// ConsiderationEntity は 10 概念モデルの検討事項
// considerations/con-NNN.yaml で管理
type ConsiderationEntity struct {
	ID          string                `yaml:"id"`
	Title       string                `yaml:"title"`
	Status      ConsiderationStatus   `yaml:"status"`
	ObjectiveID string                `yaml:"objective_id,omitempty"`
	Context     string                `yaml:"context,omitempty"`
	Options     []ConsiderationOption `yaml:"options,omitempty"`
	DecisionID  string                `yaml:"decision_id,omitempty"`
	RaisedBy    string                `yaml:"raised_by,omitempty"`
	DueDate     string                `yaml:"due_date,omitempty"`
	Metadata    Metadata              `yaml:"metadata"`
}

// Validate は ConsiderationEntity の妥当性を検証
func (c *ConsiderationEntity) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("consideration ID is required")
	}
	if err := ValidateID("consideration", c.ID); err != nil {
		return err
	}
	if c.Title == "" {
		return fmt.Errorf("consideration title is required")
	}
	if c.Status == "" {
		c.Status = ConsiderationStatusOpen
	}
	switch c.Status {
	case ConsiderationStatusOpen, ConsiderationStatusDecided, ConsiderationStatusDeferred:
		// 有効
	default:
		return fmt.Errorf("invalid consideration status: %s", c.Status)
	}
	// Option ID のユニーク性チェック
	optionIDs := make(map[string]bool)
	for _, opt := range c.Options {
		if opt.ID == "" {
			return fmt.Errorf("option ID is required")
		}
		if optionIDs[opt.ID] {
			return fmt.Errorf("duplicate option ID: %s", opt.ID)
		}
		optionIDs[opt.ID] = true
	}
	return nil
}

// GetID は Entity インターフェースを実装（ConsiderationEntity）
func (c *ConsiderationEntity) GetID() string { return c.ID }

// GetTitle は Entity インターフェースを実装（ConsiderationEntity）
func (c *ConsiderationEntity) GetTitle() string { return c.Title }

// === Decision（イミュータブル）===

// SelectedOption は選択された選択肢
type SelectedOption struct {
	OptionID string `yaml:"option_id"`
	Title    string `yaml:"title"`
}

// RejectedOption は却下された選択肢
type RejectedOption struct {
	OptionID string `yaml:"option_id"`
	Title    string `yaml:"title"`
	Reason   string `yaml:"reason,omitempty"`
}

// DecisionEntity は 10 概念モデルの決定事項（イミュータブル）
// decisions/dec-NNN.yaml で管理
// 一度作成されると更新不可
type DecisionEntity struct {
	ID              string           `yaml:"id"`
	Title           string           `yaml:"title"`
	ConsiderationID string           `yaml:"consideration_id"`
	Selected        SelectedOption   `yaml:"selected"`
	Rejected        []RejectedOption `yaml:"rejected,omitempty"`
	Rationale       string           `yaml:"rationale"`
	Impact          []string         `yaml:"impact,omitempty"`
	DecidedAt       string           `yaml:"decided_at"`
	DecidedBy       string           `yaml:"decided_by,omitempty"`
}

// Validate は DecisionEntity の妥当性を検証
func (d *DecisionEntity) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("decision ID is required")
	}
	if err := ValidateID("decision", d.ID); err != nil {
		return err
	}
	if d.Title == "" {
		return fmt.Errorf("decision title is required")
	}
	if d.ConsiderationID == "" {
		return fmt.Errorf("decision consideration_id is required")
	}
	if d.Selected.OptionID == "" {
		return fmt.Errorf("decision selected option_id is required")
	}
	if d.Rationale == "" {
		return fmt.Errorf("decision rationale is required")
	}
	if d.DecidedAt == "" {
		return fmt.Errorf("decision decided_at is required")
	}
	return nil
}

// GetID は Entity インターフェースを実装（DecisionEntity）
func (d *DecisionEntity) GetID() string { return d.ID }

// GetTitle は Entity インターフェースを実装（DecisionEntity）
func (d *DecisionEntity) GetTitle() string { return d.Title }

// === Problem ===

// ProblemStatus は Problem の状態
type ProblemStatus string

const (
	ProblemStatusOpen       ProblemStatus = "open"
	ProblemStatusInProgress ProblemStatus = "in_progress"
	ProblemStatusResolved   ProblemStatus = "resolved"
	ProblemStatusWontFix    ProblemStatus = "wont_fix"
)

// ProblemSeverity は Problem の重大度
type ProblemSeverity string

const (
	ProblemSeverityCritical ProblemSeverity = "critical"
	ProblemSeverityHigh     ProblemSeverity = "high"
	ProblemSeverityMedium   ProblemSeverity = "medium"
	ProblemSeverityLow      ProblemSeverity = "low"
)

// ProblemEntity は 10 概念モデルの問題
// problems/prob-NNN.yaml で管理
type ProblemEntity struct {
	ID                 string          `yaml:"id"`
	Title              string          `yaml:"title"`
	Status             ProblemStatus   `yaml:"status"`
	Severity           ProblemSeverity `yaml:"severity"`
	ObjectiveID        string          `yaml:"objective_id,omitempty"`
	Description        string          `yaml:"description,omitempty"`
	Impact             string          `yaml:"impact,omitempty"`
	RootCause          string          `yaml:"root_cause,omitempty"`
	PotentialSolutions []string        `yaml:"potential_solutions,omitempty"`
	ReportedBy         string          `yaml:"reported_by,omitempty"`
	AssignedTo         string          `yaml:"assigned_to,omitempty"`
	Metadata           Metadata        `yaml:"metadata"`
}

// Validate は ProblemEntity の妥当性を検証
func (p *ProblemEntity) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("problem ID is required")
	}
	if err := ValidateID("problem", p.ID); err != nil {
		return err
	}
	if p.Title == "" {
		return fmt.Errorf("problem title is required")
	}
	if p.Status == "" {
		p.Status = ProblemStatusOpen
	}
	switch p.Status {
	case ProblemStatusOpen, ProblemStatusInProgress, ProblemStatusResolved, ProblemStatusWontFix:
		// 有効
	default:
		return fmt.Errorf("invalid problem status: %s", p.Status)
	}
	if p.Severity == "" {
		p.Severity = ProblemSeverityMedium
	}
	switch p.Severity {
	case ProblemSeverityCritical, ProblemSeverityHigh, ProblemSeverityMedium, ProblemSeverityLow:
		// 有効
	default:
		return fmt.Errorf("invalid problem severity: %s", p.Severity)
	}
	return nil
}

// GetID は Entity インターフェースを実装（ProblemEntity）
func (p *ProblemEntity) GetID() string { return p.ID }

// GetTitle は Entity インターフェースを実装（ProblemEntity）
func (p *ProblemEntity) GetTitle() string { return p.Title }

// === Risk ===

// RiskStatus は Risk の状態
type RiskStatus string

const (
	RiskStatusIdentified RiskStatus = "identified"
	RiskStatusMitigating RiskStatus = "mitigating"
	RiskStatusMitigated  RiskStatus = "mitigated"
	RiskStatusOccurred   RiskStatus = "occurred"
	RiskStatusClosed     RiskStatus = "closed"
)

// RiskProbability は Risk の発生確率
type RiskProbability string

const (
	RiskProbabilityHigh   RiskProbability = "high"
	RiskProbabilityMedium RiskProbability = "medium"
	RiskProbabilityLow    RiskProbability = "low"
)

// RiskImpact は Risk の影響度
type RiskImpact string

const (
	RiskImpactCritical RiskImpact = "critical"
	RiskImpactHigh     RiskImpact = "high"
	RiskImpactMedium   RiskImpact = "medium"
	RiskImpactLow      RiskImpact = "low"
)

// RiskScore は Risk の総合スコア（自動計算）
type RiskScore string

const (
	RiskScoreCritical RiskScore = "critical"
	RiskScoreHigh     RiskScore = "high"
	RiskScoreMedium   RiskScore = "medium"
	RiskScoreLow      RiskScore = "low"
)

// RiskMitigation は Risk の軽減策
type RiskMitigation struct {
	Preventive []string `yaml:"preventive,omitempty"`
	Contingent []string `yaml:"contingent,omitempty"`
}

// RiskEntity は 10 概念モデルのリスク
// risks/risk-NNN.yaml で管理
type RiskEntity struct {
	ID          string          `yaml:"id"`
	Title       string          `yaml:"title"`
	Status      RiskStatus      `yaml:"status"`
	Probability RiskProbability `yaml:"probability"`
	Impact      RiskImpact      `yaml:"impact"`
	RiskScore   RiskScore       `yaml:"risk_score"` // 自動計算
	ObjectiveID string          `yaml:"objective_id,omitempty"`
	Description string          `yaml:"description,omitempty"`
	Trigger     string          `yaml:"trigger,omitempty"`
	Mitigation  RiskMitigation  `yaml:"mitigation,omitempty"`
	Owner       string          `yaml:"owner,omitempty"`
	ReviewDate  string          `yaml:"review_date,omitempty"`
	Metadata    Metadata        `yaml:"metadata"`
}

// CalculateRiskScore は probability × impact から risk_score を計算
func CalculateRiskScore(probability RiskProbability, impact RiskImpact) RiskScore {
	matrix := map[RiskProbability]map[RiskImpact]RiskScore{
		RiskProbabilityHigh: {
			RiskImpactCritical: RiskScoreCritical,
			RiskImpactHigh:     RiskScoreCritical,
			RiskImpactMedium:   RiskScoreHigh,
			RiskImpactLow:      RiskScoreMedium,
		},
		RiskProbabilityMedium: {
			RiskImpactCritical: RiskScoreCritical,
			RiskImpactHigh:     RiskScoreHigh,
			RiskImpactMedium:   RiskScoreMedium,
			RiskImpactLow:      RiskScoreLow,
		},
		RiskProbabilityLow: {
			RiskImpactCritical: RiskScoreHigh,
			RiskImpactHigh:     RiskScoreMedium,
			RiskImpactMedium:   RiskScoreLow,
			RiskImpactLow:      RiskScoreLow,
		},
	}
	if impactMap, ok := matrix[probability]; ok {
		if score, ok := impactMap[impact]; ok {
			return score
		}
	}
	return RiskScoreMedium // デフォルト
}

// Validate は RiskEntity の妥当性を検証
func (r *RiskEntity) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("risk ID is required")
	}
	if err := ValidateID("risk", r.ID); err != nil {
		return err
	}
	if r.Title == "" {
		return fmt.Errorf("risk title is required")
	}
	if r.Status == "" {
		r.Status = RiskStatusIdentified
	}
	switch r.Status {
	case RiskStatusIdentified, RiskStatusMitigating, RiskStatusMitigated,
		RiskStatusOccurred, RiskStatusClosed:
		// 有効
	default:
		return fmt.Errorf("invalid risk status: %s", r.Status)
	}
	if r.Probability == "" {
		r.Probability = RiskProbabilityMedium
	}
	switch r.Probability {
	case RiskProbabilityHigh, RiskProbabilityMedium, RiskProbabilityLow:
		// 有効
	default:
		return fmt.Errorf("invalid risk probability: %s", r.Probability)
	}
	if r.Impact == "" {
		r.Impact = RiskImpactMedium
	}
	switch r.Impact {
	case RiskImpactCritical, RiskImpactHigh, RiskImpactMedium, RiskImpactLow:
		// 有効
	default:
		return fmt.Errorf("invalid risk impact: %s", r.Impact)
	}
	// RiskScore を自動計算
	r.RiskScore = CalculateRiskScore(r.Probability, r.Impact)
	return nil
}

// GetID は Entity インターフェースを実装（RiskEntity）
func (r *RiskEntity) GetID() string { return r.ID }

// GetTitle は Entity インターフェースを実装（RiskEntity）
func (r *RiskEntity) GetTitle() string { return r.Title }

// === Assumption ===

// AssumptionStatus は Assumption の状態
type AssumptionStatus string

const (
	AssumptionStatusAssumed     AssumptionStatus = "assumed"
	AssumptionStatusValidated   AssumptionStatus = "validated"
	AssumptionStatusInvalidated AssumptionStatus = "invalidated"
)

// AssumptionValidation は Assumption の検証情報
type AssumptionValidation struct {
	Method      string `yaml:"method,omitempty"`
	Result      string `yaml:"result,omitempty"`
	ValidatedAt string `yaml:"validated_at,omitempty"`
}

// AssumptionEntity は 10 概念モデルの前提条件
// assumptions/assum-NNN.yaml で管理
type AssumptionEntity struct {
	ID          string               `yaml:"id"`
	Title       string               `yaml:"title"`
	Status      AssumptionStatus     `yaml:"status"`
	ObjectiveID string               `yaml:"objective_id,omitempty"`
	Description string               `yaml:"description,omitempty"`
	IfInvalid   string               `yaml:"if_invalid,omitempty"`
	Validation  AssumptionValidation `yaml:"validation,omitempty"`
	Metadata    Metadata             `yaml:"metadata"`
}

// Validate は AssumptionEntity の妥当性を検証
func (a *AssumptionEntity) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("assumption ID is required")
	}
	if err := ValidateID("assumption", a.ID); err != nil {
		return err
	}
	if a.Title == "" {
		return fmt.Errorf("assumption title is required")
	}
	if a.Status == "" {
		a.Status = AssumptionStatusAssumed
	}
	switch a.Status {
	case AssumptionStatusAssumed, AssumptionStatusValidated, AssumptionStatusInvalidated:
		// 有効
	default:
		return fmt.Errorf("invalid assumption status: %s", a.Status)
	}
	return nil
}

// GetID は Entity インターフェースを実装（AssumptionEntity）
func (a *AssumptionEntity) GetID() string { return a.ID }

// GetTitle は Entity インターフェースを実装（AssumptionEntity）
func (a *AssumptionEntity) GetTitle() string { return a.Title }

// ============================================================
// 10 概念モデル型定義 (Phase 3: Constraint, Quality)
// ============================================================

// === Constraint（単一ファイル）===

// ConstraintCategory は Constraint のカテゴリ
type ConstraintCategory string

const (
	ConstraintCategoryTechnical ConstraintCategory = "technical"
	ConstraintCategoryBusiness  ConstraintCategory = "business"
	ConstraintCategoryLegal     ConstraintCategory = "legal"
	ConstraintCategoryResource  ConstraintCategory = "resource"
)

// ConstraintEntity は 10 概念モデルの制約条件
type ConstraintEntity struct {
	ID            string             `yaml:"id"`
	Title         string             `yaml:"title"`
	Category      ConstraintCategory `yaml:"category"`
	Description   string             `yaml:"description,omitempty"`
	Source        string             `yaml:"source,omitempty"`
	Impact        []string           `yaml:"impact,omitempty"`
	NonNegotiable bool               `yaml:"non_negotiable"`
}

// Validate は ConstraintEntity の妥当性を検証
func (c *ConstraintEntity) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("constraint ID is required")
	}
	if err := ValidateID("constraint", c.ID); err != nil {
		return err
	}
	if c.Title == "" {
		return fmt.Errorf("constraint title is required")
	}
	if c.Category == "" {
		c.Category = ConstraintCategoryTechnical
	}
	switch c.Category {
	case ConstraintCategoryTechnical, ConstraintCategoryBusiness,
		ConstraintCategoryLegal, ConstraintCategoryResource:
		// 有効
	default:
		return fmt.Errorf("invalid constraint category: %s", c.Category)
	}
	return nil
}

// GetID は Entity インターフェースを実装（ConstraintEntity）
func (c *ConstraintEntity) GetID() string { return c.ID }

// GetTitle は Entity インターフェースを実装（ConstraintEntity）
func (c *ConstraintEntity) GetTitle() string { return c.Title }

// ConstraintsFile は制約条件ファイルの構造
// constraints.yaml で管理（単一ファイル）
type ConstraintsFile struct {
	Constraints []ConstraintEntity `yaml:"constraints"`
	Metadata    Metadata           `yaml:"metadata"`
}

// === Quality ===

// MetricStatus は QualityMetric の状態
type MetricStatus string

const (
	MetricStatusMet        MetricStatus = "met"
	MetricStatusNotMet     MetricStatus = "not_met"
	MetricStatusInProgress MetricStatus = "in_progress"
)

// GateStatus は QualityGate の状態
type GateStatus string

const (
	GateStatusPassed  GateStatus = "passed"
	GateStatusFailed  GateStatus = "failed"
	GateStatusPending GateStatus = "pending"
)

// QualityMetric は品質メトリクス
type QualityMetric struct {
	ID      string       `yaml:"id"`
	Name    string       `yaml:"name"`
	Target  float64      `yaml:"target"`
	Unit    string       `yaml:"unit,omitempty"`
	Current float64      `yaml:"current,omitempty"`
	Status  MetricStatus `yaml:"status"`
}

// QualityGate は品質ゲート
type QualityGate struct {
	Name     string     `yaml:"name"`
	Criteria []string   `yaml:"criteria"`
	Status   GateStatus `yaml:"status"`
}

// QualityEntity は 10 概念モデルの品質基準
// quality/qual-NNN.yaml で管理
type QualityEntity struct {
	ID          string          `yaml:"id"`
	Title       string          `yaml:"title"`
	ObjectiveID string          `yaml:"objective_id"`
	Metrics     []QualityMetric `yaml:"metrics"`
	Gates       []QualityGate   `yaml:"gates,omitempty"`
	Reviewer    string          `yaml:"reviewer,omitempty"`
	Metadata    Metadata        `yaml:"metadata"`
}

// Validate は QualityEntity の妥当性を検証
func (q *QualityEntity) Validate() error {
	if q.ID == "" {
		return fmt.Errorf("quality ID is required")
	}
	if err := ValidateID("quality", q.ID); err != nil {
		return err
	}
	if q.Title == "" {
		return fmt.Errorf("quality title is required")
	}
	if q.ObjectiveID == "" {
		return fmt.Errorf("quality objective_id is required")
	}
	if len(q.Metrics) == 0 {
		return fmt.Errorf("quality must have at least one metric")
	}
	// Metric バリデーション
	metricIDs := make(map[string]bool)
	for _, m := range q.Metrics {
		if m.ID == "" {
			return fmt.Errorf("metric ID is required")
		}
		if metricIDs[m.ID] {
			return fmt.Errorf("duplicate metric ID: %s", m.ID)
		}
		metricIDs[m.ID] = true
		if m.Name == "" {
			return fmt.Errorf("metric name is required")
		}
		if m.Status == "" {
			m.Status = MetricStatusInProgress
		}
		switch m.Status {
		case MetricStatusMet, MetricStatusNotMet, MetricStatusInProgress:
			// 有効
		default:
			return fmt.Errorf("invalid metric status: %s", m.Status)
		}
	}
	// Gate バリデーション
	for _, g := range q.Gates {
		if g.Name == "" {
			return fmt.Errorf("gate name is required")
		}
		if g.Status == "" {
			g.Status = GateStatusPending
		}
		switch g.Status {
		case GateStatusPassed, GateStatusFailed, GateStatusPending:
			// 有効
		default:
			return fmt.Errorf("invalid gate status: %s", g.Status)
		}
	}
	return nil
}

// GetID は Entity インターフェースを実装（QualityEntity）
func (q *QualityEntity) GetID() string { return q.ID }

// GetTitle は Entity インターフェースを実装（QualityEntity）
func (q *QualityEntity) GetTitle() string { return q.Title }

// ============================================================
// UML ユースケース図型定義 (Actor, UseCase, Subsystem)
// ============================================================

// === Actor ===

// ActorType はアクターの種類
type ActorType string

const (
	ActorTypeHuman    ActorType = "human"
	ActorTypeSystem   ActorType = "system"
	ActorTypeTime     ActorType = "time"
	ActorTypeDevice   ActorType = "device"
	ActorTypeExternal ActorType = "external"
)

// ActorEntity はアクターエンティティ
type ActorEntity struct {
	ID          string    `yaml:"id"`
	Title       string    `yaml:"title"`
	Type        ActorType `yaml:"type"`
	Description string    `yaml:"description,omitempty"`
	Metadata    Metadata  `yaml:"metadata"`
}

// ActorsFile はアクターファイルの構造（単一ファイル管理）
type ActorsFile struct {
	Actors []ActorEntity `yaml:"actors"`
}

// Validate は ActorEntity の妥当性を検証
func (a *ActorEntity) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("actor ID is required")
	}
	if err := ValidateID("actor", a.ID); err != nil {
		return err
	}
	if a.Title == "" {
		return fmt.Errorf("actor title is required")
	}
	if a.Type == "" {
		a.Type = ActorTypeHuman
	}
	switch a.Type {
	case ActorTypeHuman, ActorTypeSystem, ActorTypeTime, ActorTypeDevice, ActorTypeExternal:
		// 有効
	default:
		return fmt.Errorf("invalid actor type: %s", a.Type)
	}
	return nil
}

// GetID は Entity インターフェースを実装（ActorEntity）
func (a *ActorEntity) GetID() string { return a.ID }

// GetTitle は Entity インターフェースを実装（ActorEntity）
func (a *ActorEntity) GetTitle() string { return a.Title }

// === Subsystem ===

// SubsystemEntity はサブシステムエンティティ（UML ユースケース図のシステム境界）
type SubsystemEntity struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Metadata    Metadata `yaml:"metadata"`
}

// SubsystemsFile はサブシステムファイルの構造（単一ファイル管理）
type SubsystemsFile struct {
	Subsystems []SubsystemEntity `yaml:"subsystems"`
}

// Validate は SubsystemEntity の妥当性を検証
func (s *SubsystemEntity) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("subsystem ID is required")
	}
	if err := ValidateID("subsystem", s.ID); err != nil {
		return err
	}
	if s.Name == "" {
		return fmt.Errorf("subsystem name is required")
	}
	return nil
}

// GetID は Entity インターフェースを実装（SubsystemEntity）
func (s *SubsystemEntity) GetID() string { return s.ID }

// GetTitle は Entity インターフェースを実装（SubsystemEntity）
func (s *SubsystemEntity) GetTitle() string { return s.Name }

// === UseCase ===

// UseCaseStatus はユースケースの状態
type UseCaseStatus string

const (
	UseCaseStatusDraft      UseCaseStatus = "draft"
	UseCaseStatusActive     UseCaseStatus = "active"
	UseCaseStatusDeprecated UseCaseStatus = "deprecated"
)

// ActorRole はアクターの役割
type ActorRole string

const (
	ActorRolePrimary   ActorRole = "primary"
	ActorRoleSecondary ActorRole = "secondary"
)

// UseCaseActorRef はユースケースとアクターの関連
type UseCaseActorRef struct {
	ActorID string    `yaml:"actor_id"`
	Role    ActorRole `yaml:"role"`
}

// RelationType は関係の種類
type RelationType string

const (
	RelationTypeInclude    RelationType = "include"
	RelationTypeExtend     RelationType = "extend"
	RelationTypeGeneralize RelationType = "generalize"
)

// UseCaseRelation はユースケース間の関係
type UseCaseRelation struct {
	Type           RelationType `yaml:"type"`
	TargetID       string       `yaml:"target_id"`
	ExtensionPoint string       `yaml:"extension_point,omitempty"`
	Condition      string       `yaml:"condition,omitempty"`
}

// AlternativeFlow は代替フロー（UML 2.5 準拠）
type AlternativeFlow struct {
	ID        string   `yaml:"id"`
	Name      string   `yaml:"name"`
	Condition string   `yaml:"condition"`
	Steps     []string `yaml:"steps"`
	RejoinsAt string   `yaml:"rejoins_at,omitempty"` // メインフローに戻るステップ
}

// ExceptionFlow は例外フロー（UML 2.5 準拠）
type ExceptionFlow struct {
	ID      string   `yaml:"id"`
	Name    string   `yaml:"name"`
	Trigger string   `yaml:"trigger"` // 例外発生条件
	Steps   []string `yaml:"steps"`
	Outcome string   `yaml:"outcome,omitempty"` // 結果（例: "ステップ2へ戻る"）
}

// UseCaseScenario はシナリオ（UML 2.5 準拠の拡張版）
type UseCaseScenario struct {
	Preconditions    []string          `yaml:"preconditions,omitempty"`     // 事前条件
	Trigger          string            `yaml:"trigger,omitempty"`           // 開始イベント
	MainFlow         []string          `yaml:"main_flow,omitempty"`         // メインフロー
	AlternativeFlows []AlternativeFlow `yaml:"alternative_flows,omitempty"` // 代替フロー
	ExceptionFlows   []ExceptionFlow   `yaml:"exception_flows,omitempty"`   // 例外フロー
	Postconditions   []string          `yaml:"postconditions,omitempty"`    // 事後条件
}

// UseCaseEntity はユースケースエンティティ
// usecases/uc-NNN.yaml で管理（個別ファイル）
type UseCaseEntity struct {
	ID          string            `yaml:"id"`
	Title       string            `yaml:"title"`
	ObjectiveID string            `yaml:"objective_id"` // 必須
	Description string            `yaml:"description,omitempty"`
	SubsystemID string            `yaml:"subsystem_id,omitempty"` // サブシステム ID（オプション）
	Actors      []UseCaseActorRef `yaml:"actors,omitempty"`
	Relations   []UseCaseRelation `yaml:"relations,omitempty"`
	Scenario    UseCaseScenario   `yaml:"scenario,omitempty"`
	Status      UseCaseStatus     `yaml:"status"`
	Metadata    Metadata          `yaml:"metadata"`
}

// Validate は UseCaseEntity の妥当性を検証
func (u *UseCaseEntity) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("usecase ID is required")
	}
	if err := ValidateID("usecase", u.ID); err != nil {
		return err
	}
	if u.Title == "" {
		return fmt.Errorf("usecase title is required")
	}
	// ObjectiveID は必須（参照整合性のため）
	if u.ObjectiveID == "" {
		return fmt.Errorf("usecase objective_id is required")
	}
	if u.Status == "" {
		u.Status = UseCaseStatusDraft
	}
	switch u.Status {
	case UseCaseStatusDraft, UseCaseStatusActive, UseCaseStatusDeprecated:
		// 有効
	default:
		return fmt.Errorf("invalid usecase status: %s", u.Status)
	}
	// Actor 参照の検証
	for _, ar := range u.Actors {
		if ar.ActorID == "" {
			return fmt.Errorf("actor_id is required in actor reference")
		}
		if ar.Role == "" {
			ar.Role = ActorRolePrimary
		}
		switch ar.Role {
		case ActorRolePrimary, ActorRoleSecondary:
			// 有効
		default:
			return fmt.Errorf("invalid actor role: %s", ar.Role)
		}
	}
	// Relation の検証
	for _, rel := range u.Relations {
		if rel.TargetID == "" {
			return fmt.Errorf("target_id is required in relation")
		}
		if rel.Type == "" {
			return fmt.Errorf("relation type is required")
		}
		switch rel.Type {
		case RelationTypeInclude, RelationTypeExtend, RelationTypeGeneralize:
			// 有効
		default:
			return fmt.Errorf("invalid relation type: %s", rel.Type)
		}
		// extend の場合は extension_point が推奨
		// （必須ではない、condition も任意）
	}
	return nil
}

// GetID は Entity インターフェースを実装（UseCaseEntity）
func (u *UseCaseEntity) GetID() string { return u.ID }

// GetTitle は Entity インターフェースを実装（UseCaseEntity）
func (u *UseCaseEntity) GetTitle() string { return u.Title }

// ============================================================
// UML アクティビティ図型定義 (Activity)
// Task/Activity 統合により、Activity が「実行可能な作業単位」として機能
// ============================================================

// === Activity ===

// ActivityStatus はアクティビティの状態
type ActivityStatus string

const (
	ActivityStatusDraft      ActivityStatus = "draft"
	ActivityStatusActive     ActivityStatus = "active"
	ActivityStatusDeprecated ActivityStatus = "deprecated"
)

// ActivityNodeType はアクティビティノードの種類
type ActivityNodeType string

const (
	ActivityNodeTypeInitial  ActivityNodeType = "initial"  // 開始ノード（黒丸）
	ActivityNodeTypeFinal    ActivityNodeType = "final"    // 終了ノード（二重丸）
	ActivityNodeTypeAction   ActivityNodeType = "action"   // アクション（角丸四角形）
	ActivityNodeTypeDecision ActivityNodeType = "decision" // 分岐（ひし形）
	ActivityNodeTypeMerge    ActivityNodeType = "merge"    // 合流（ひし形）
	ActivityNodeTypeFork     ActivityNodeType = "fork"     // 並列分岐（太い横線）
	ActivityNodeTypeJoin     ActivityNodeType = "join"     // 並列合流（太い横線）
)

// ActivityNode はアクティビティ図のノード
type ActivityNode struct {
	ID   string           `yaml:"id"`
	Type ActivityNodeType `yaml:"type"`
	Name string           `yaml:"name,omitempty"` // initial/final では不要
}

// ActivityTransition はアクティビティ図の遷移
type ActivityTransition struct {
	ID     string `yaml:"id"`
	Source string `yaml:"source"`          // ソースノードID
	Target string `yaml:"target"`          // ターゲットノードID
	Guard  string `yaml:"guard,omitempty"` // ガード条件（例: "[認証成功]"）
}

// ActivityEntity はアクティビティ図エンティティ
// activities/act-NNN.yaml で管理（個別ファイル）
// Task/Activity 統合により、作業管理フィールドを追加
type ActivityEntity struct {
	ID          string               `yaml:"id"`
	Title       string               `yaml:"title"`
	Description string               `yaml:"description,omitempty"`
	UseCaseID   string               `yaml:"usecase_id,omitempty"` // 任意紐付け
	Status      ActivityStatus       `yaml:"status"`
	Nodes       []ActivityNode       `yaml:"nodes,omitempty"`
	Transitions []ActivityTransition `yaml:"transitions,omitempty"`
	Metadata    Metadata             `yaml:"metadata"`
}

// Validate は ActivityNode の妥当性を検証
func (n *ActivityNode) Validate() error {
	if n.ID == "" {
		return fmt.Errorf("activity node ID is required")
	}
	if n.Type == "" {
		return fmt.Errorf("activity node type is required")
	}
	switch n.Type {
	case ActivityNodeTypeInitial, ActivityNodeTypeFinal, ActivityNodeTypeAction,
		ActivityNodeTypeDecision, ActivityNodeTypeMerge, ActivityNodeTypeFork, ActivityNodeTypeJoin:
		// 有効
	default:
		return fmt.Errorf("invalid activity node type: %s", n.Type)
	}
	// action, decision には name が推奨（必須ではない）

	return nil
}

// Validate は ActivityTransition の妥当性を検証
func (t *ActivityTransition) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("activity transition ID is required")
	}
	if t.Source == "" {
		return fmt.Errorf("activity transition source is required")
	}
	if t.Target == "" {
		return fmt.Errorf("activity transition target is required")
	}
	return nil
}

// Validate は ActivityEntity の妥当性を検証
func (a *ActivityEntity) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("activity ID is required")
	}
	if err := ValidateID("activity", a.ID); err != nil {
		return err
	}
	if a.Title == "" {
		return fmt.Errorf("activity title is required")
	}
	if a.Status == "" {
		a.Status = ActivityStatusDraft
	}
	// ステータスのバリデーション
	switch a.Status {
	case ActivityStatusDraft, ActivityStatusActive, ActivityStatusDeprecated:
		// 有効
	default:
		return fmt.Errorf("invalid activity status: %s", a.Status)
	}

	// ノードのバリデーションとID重複チェック
	nodeIDs := make(map[string]bool)
	for _, node := range a.Nodes {
		if err := node.Validate(); err != nil {
			return fmt.Errorf("invalid node: %w", err)
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}
	// 遷移のバリデーションとID重複チェック
	transitionIDs := make(map[string]bool)
	for _, trans := range a.Transitions {
		if err := trans.Validate(); err != nil {
			return fmt.Errorf("invalid transition: %w", err)
		}
		if transitionIDs[trans.ID] {
			return fmt.Errorf("duplicate transition ID: %s", trans.ID)
		}
		transitionIDs[trans.ID] = true
		// ソース/ターゲットがノードに存在するか確認
		if !nodeIDs[trans.Source] {
			return fmt.Errorf("transition source not found: %s", trans.Source)
		}
		if !nodeIDs[trans.Target] {
			return fmt.Errorf("transition target not found: %s", trans.Target)
		}
	}

	return nil
}

// GetID は Entity インターフェースを実装（ActivityEntity）
func (a *ActivityEntity) GetID() string { return a.ID }

// GetTitle は Entity インターフェースを実装（ActivityEntity）
func (a *ActivityEntity) GetTitle() string { return a.Title }
