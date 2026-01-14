package core

import "time"

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
	AutomationLevel string `yaml:"automation_level"` // simple, standard, advanced
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

// Task はタスク
type Task struct {
	ID            string        `yaml:"id"`
	Title         string        `yaml:"title"`
	Description   string        `yaml:"description,omitempty"`
	Status        TaskStatus    `yaml:"status"`
	Assignee      string        `yaml:"assignee,omitempty"`
	EstimateHours float64       `yaml:"estimate_hours,omitempty"`
	ActualHours   float64       `yaml:"actual_hours,omitempty"`
	Dependencies  []string      `yaml:"dependencies"`
	ApprovalLevel ApprovalLevel `yaml:"approval_level"`
	CreatedAt     string        `yaml:"created_at"`
	UpdatedAt     string        `yaml:"updated_at"`
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
	Level      string
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
	Success bool
	ID      string
	Entity  string
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
