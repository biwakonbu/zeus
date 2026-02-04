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

	// Phase 6A: WBS・タイムライン機能用フィールド
	ParentID      string  // 親タスクID
	StartDate     string  // 開始日（ISO8601）
	DueDate       string  // 期限日（ISO8601）
	Progress      int     // 進捗率（0-100）
	WBSCode       string  // WBS番号（例: "1.2.3"）
	Priority      string  // 優先度 ("high", "medium", "low")
	Assignee      string  // 担当者
	EstimateHours float64 // 見積もり時間

	// 陳腐化分析用フィールド
	CreatedAt   string // 作成日時（ISO8601）
	UpdatedAt   string // 更新日時（ISO8601）
	CompletedAt string // 完了日時（ISO8601）
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
	Timestamp string       // タイムスタンプ
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

// ObjectiveInfo は分析に必要な Objective 情報
type ObjectiveInfo struct {
	ID        string // Objective ID
	Title     string // タイトル
	WBSCode   string // WBS コード
	Progress  int    // 進捗率
	Status    string // ステータス
	ParentID  string // 親 Objective ID（L3 の場合）
	CreatedAt string // 作成日時（ISO8601）
	UpdatedAt string // 更新日時（ISO8601）
}

// DeliverableInfo は分析に必要な Deliverable 情報
type DeliverableInfo struct {
	ID          string // Deliverable ID
	Title       string // タイトル
	ObjectiveID string // 紐づく Objective ID
	Progress    int    // 進捗率
	Status      string // ステータス
	CreatedAt   string // 作成日時（ISO8601）
	UpdatedAt   string // 更新日時（ISO8601）
}

// VisionInfo は分析に必要な Vision 情報
type VisionInfo struct {
	ID        string // Vision ID
	Title     string // タイトル
	Statement string // ステートメント
	Status    string // ステータス
}

// ============================================================
// UnifiedGraph 型定義 (Task/Activity 統合)
// ============================================================

// EntityType は統合グラフのエンティティタイプ
type EntityType string

const (
	EntityTypeActivity    EntityType = "activity"
	EntityTypeUseCase     EntityType = "usecase"
	EntityTypeDeliverable EntityType = "deliverable"
	EntityTypeObjective   EntityType = "objective"
)

// ActivityInfo は分析に必要な Activity 情報
// core.ActivityEntity からの変換を前提とした軽量構造体
type ActivityInfo struct {
	ID           string   // Activity ID
	Title        string   // タイトル
	Status       string   // ステータス
	Mode         string   // モード ("simple" or "flow")
	Dependencies []string // 依存 Activity ID
	ParentID     string   // 親 Activity ID
	UseCaseID    string   // 関連 UseCase ID

	// 作業管理フィールド
	Progress      int      // 進捗率（0-100）
	Priority      string   // 優先度
	Assignee      string   // 担当者
	WBSCode       string   // WBS コード
	StartDate     string   // 開始日
	DueDate       string   // 期限日
	EstimateHours float64  // 見積もり時間
	ActualHours   float64  // 実績時間

	// 関連情報
	RelatedDeliverables []string // 関連 Deliverable ID
	CreatedAt           string   // 作成日時
	UpdatedAt           string   // 更新日時
}

// UseCaseInfo は分析に必要な UseCase 情報
type UseCaseInfo struct {
	ID          string   // UseCase ID
	Title       string   // タイトル
	Status      string   // ステータス
	ObjectiveID string   // 関連 Objective ID
	SubsystemID string   // サブシステム ID
	ActorIDs    []string // 関連 Actor ID
}

// UnifiedGraphNode は統合グラフのノード
type UnifiedGraphNode struct {
	ID         string     // ノード ID
	Type       EntityType // エンティティタイプ
	Title      string     // タイトル
	Status     string     // ステータス
	Progress   int        // 進捗率（該当する場合）
	Priority   string     // 優先度（Activity のみ）
	Assignee   string     // 担当者（Activity のみ）
	Mode       string     // モード（Activity のみ: "simple" or "flow"）

	// 関連情報
	Parents  []string // 親ノード ID
	Children []string // 子ノード ID
	Depth    int      // ルートからの深さ
}

// UnifiedEdge は統合グラフのエッジ
type UnifiedEdge struct {
	From     string         // ソースノード ID
	To       string         // ターゲットノード ID
	Type     UnifiedEdgeType // エッジタイプ
	Label    string         // ラベル（オプション）
}

// UnifiedEdgeType は統合グラフのエッジタイプ
type UnifiedEdgeType string

const (
	EdgeTypeDependency  UnifiedEdgeType = "dependency"  // 依存関係
	EdgeTypeParent      UnifiedEdgeType = "parent"      // 親子関係
	EdgeTypeRelates     UnifiedEdgeType = "relates"     // 関連（Activity → UseCase/Deliverable）
	EdgeTypeContributes UnifiedEdgeType = "contributes" // 貢献（UseCase → Objective）
)

// UnifiedGraph は統合グラフ
type UnifiedGraph struct {
	Nodes    map[string]*UnifiedGraphNode // ID をキーとするノードマップ
	Edges    []UnifiedEdge                // エッジリスト
	Cycles   [][]string                   // 循環依存（検出された場合）
	Isolated []string                     // 孤立ノード
	Stats    UnifiedGraphStats            // 統計情報
}

// UnifiedGraphStats は統合グラフの統計情報
type UnifiedGraphStats struct {
	TotalNodes          int            // 総ノード数
	NodesByType         map[EntityType]int // タイプ別ノード数
	TotalEdges          int            // 総エッジ数
	EdgesByType         map[UnifiedEdgeType]int // タイプ別エッジ数
	IsolatedCount       int            // 孤立ノード数
	CycleCount          int            // 循環依存の数
	MaxDepth            int            // 最大深さ
	CompletedActivities int            // 完了 Activity 数
	TotalActivities     int            // 総 Activity 数
}

// GraphFilter はグラフ表示のフィルタリングオプション
type GraphFilter struct {
	FocusID       string       // 中心ノード ID
	FocusDepth    int          // 中心からの深さ（0 = 無制限）
	IncludeTypes  []EntityType // 含めるエンティティタイプ
	ExcludeTypes  []EntityType // 除外するエンティティタイプ
	HideCompleted bool         // 完了済みを非表示
	HideDraft     bool         // ドラフトを非表示
	HideUnrelated bool         // 無関係ノードを非表示
}

// NewGraphFilter はデフォルトのフィルタを作成
func NewGraphFilter() *GraphFilter {
	return &GraphFilter{
		FocusDepth:    0,
		IncludeTypes:  []EntityType{},
		ExcludeTypes:  []EntityType{},
		HideCompleted: false,
		HideDraft:     false,
		HideUnrelated: false,
	}
}

// WithFocus はフォーカスノードを設定
func (f *GraphFilter) WithFocus(id string, depth int) *GraphFilter {
	f.FocusID = id
	f.FocusDepth = depth
	return f
}

// WithIncludeTypes は含めるタイプを設定
func (f *GraphFilter) WithIncludeTypes(types ...EntityType) *GraphFilter {
	f.IncludeTypes = types
	return f
}

// WithExcludeTypes は除外するタイプを設定
func (f *GraphFilter) WithExcludeTypes(types ...EntityType) *GraphFilter {
	f.ExcludeTypes = types
	return f
}

// WithHideCompleted は完了済み非表示を設定
func (f *GraphFilter) WithHideCompleted(hide bool) *GraphFilter {
	f.HideCompleted = hide
	return f
}

// WithHideDraft はドラフト非表示を設定
func (f *GraphFilter) WithHideDraft(hide bool) *GraphFilter {
	f.HideDraft = hide
	return f
}

// WithHideUnrelated は無関係ノード非表示を設定
func (f *GraphFilter) WithHideUnrelated(hide bool) *GraphFilter {
	f.HideUnrelated = hide
	return f
}
