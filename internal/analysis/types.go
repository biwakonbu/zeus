// Package analysis は Zeus の高度な分析機能を提供する。
// グラフ分析、予測、リスク評価などの機能を含む。
package analysis

// TaskInfo は分析に必要なタスク情報
// core.Task からの変換を前提とした軽量構造体
type TaskInfo struct {
	ID           string   // タスクID
	Title        string   // タイトル
	Status       string   // ステータス ("pending", "in_progress", "completed", "blocked")
	Dependencies []string // 依存タスクID
}

// ProjectState は分析に必要なプロジェクト状態
type ProjectState struct {
	Health  string    // プロジェクト健全性
	Summary TaskStats // タスク統計
}

// TaskStats はタスク統計
type TaskStats struct {
	TotalTasks int // 全タスク数
	Completed  int // 完了済み
	InProgress int // 進行中
	Pending    int // 保留中
}

// Snapshot はスナップショット情報
type Snapshot struct {
	Timestamp string        // タイムスタンプ
	State     ProjectState // 状態
}

// DependencyGraph は依存関係グラフを表現
type DependencyGraph struct {
	Nodes    map[string]*GraphNode // タスクIDをキーとするノードマップ
	Edges    []Edge                // 依存関係のエッジリスト
	Cycles   [][]string            // 循環依存のパス（検出された場合）
	Isolated []string              // 孤立ノード（依存関係なし）
	Stats    GraphStats            // グラフ統計
}

// GraphNode はグラフのノード（タスク）
type GraphNode struct {
	Task     *TaskInfo // タスクデータ
	Children []string  // 依存先タスクID（このタスクが依存するタスク）
	Parents  []string  // 依存元タスクID（このタスクに依存するタスク）
	Depth    int       // ルートからの深さ
}

// Edge はグラフのエッジ（依存関係）
type Edge struct {
	From string // 依存元タスクID
	To   string // 依存先タスクID
}

// GraphStats はグラフの統計情報
type GraphStats struct {
	TotalNodes       int // 総ノード数
	WithDependencies int // 依存関係を持つノード数
	IsolatedCount    int // 孤立ノード数
	CycleCount       int // 循環依存の数
	MaxDepth         int // 最大深さ
}

// CompletionPrediction は完了日予測
type CompletionPrediction struct {
	RemainingTasks    int     // 残タスク数
	AverageVelocity   float64 // 平均ベロシティ (tasks/week)
	EstimatedDate     string  // 予測完了日 (ISO8601)
	ConfidenceLevel   int     // 信頼度 (0-100%)
	MarginDays        int     // 誤差範囲 (+/- days)
	HasSufficientData bool    // 十分なデータがあるか
}

// RiskPrediction はリスク予測
type RiskPrediction struct {
	OverallLevel RiskLevel    // 全体リスクレベル
	Factors      []RiskFactor // リスク要因リスト
	Score        int          // リスクスコア (0-100)
}

// RiskLevel はリスクレベル
type RiskLevel string

const (
	RiskLow    RiskLevel = "Low"
	RiskMedium RiskLevel = "Medium"
	RiskHigh   RiskLevel = "High"
)

// RiskFactor はリスク要因
type RiskFactor struct {
	Name        string // 要因名
	Description string // 説明
	Impact      int    // 影響度 (1-10)
}

// VelocityReport はベロシティレポート
type VelocityReport struct {
	Last7Days     int           // 過去7日間の完了タスク数
	Last14Days    int           // 過去14日間の完了タスク数
	Last30Days    int           // 過去30日間の完了タスク数
	WeeklyAverage float64       // 週平均ベロシティ
	Trend         VelocityTrend // トレンド
	DataPoints    int           // 分析に使用したデータポイント数
}

// VelocityTrend はベロシティのトレンド
type VelocityTrend string

const (
	TrendIncreasing VelocityTrend = "Increasing"
	TrendStable     VelocityTrend = "Stable"
	TrendDecreasing VelocityTrend = "Decreasing"
	TrendUnknown    VelocityTrend = "Unknown"
)

// AnalysisResult は全分析結果を集約
type AnalysisResult struct {
	Graph      *DependencyGraph      // 依存関係グラフ
	Completion *CompletionPrediction // 完了日予測
	Risk       *RiskPrediction       // リスク予測
	Velocity   *VelocityReport       // ベロシティレポート
}

// GraphFormat はグラフ出力形式
type GraphFormat string

const (
	FormatText    GraphFormat = "text"
	FormatDot     GraphFormat = "dot"
	FormatMermaid GraphFormat = "mermaid"
)

// ReportFormat はレポート出力形式
type ReportFormat string

const (
	ReportFormatText     ReportFormat = "text"
	ReportFormatHTML     ReportFormat = "html"
	ReportFormatMarkdown ReportFormat = "markdown"
)

// PredictType は予測タイプ
type PredictType string

const (
	PredictCompletion PredictType = "completion"
	PredictRisk       PredictType = "risk"
	PredictVelocity   PredictType = "velocity"
	PredictAll        PredictType = "all"
)

// タスクステータス定数
const (
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusBlocked    = "blocked"
)
