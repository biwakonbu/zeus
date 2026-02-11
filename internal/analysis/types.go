// Package analysis は Zeus の高度な分析機能を提供する。
// グラフ分析などの機能を含む。
package analysis

// TaskInfo は分析に必要なタスク情報
// core.Task からの変換を前提とした軽量構造体
type TaskInfo struct {
	ID           string   // タスクID
	Title        string   // タイトル
	Status       string   // ステータス ("pending", "in_progress", "completed", "blocked")
	Dependencies []string // 依存タスクID

	// 依存関係
	ParentID string // 親タスクID
	Priority string // 優先度 ("high", "medium", "low")
	Assignee string // 担当者

	// 陳腐化分析用フィールド
	CreatedAt   string // 作成日時（ISO8601）
	UpdatedAt   string // 更新日時（ISO8601）
	CompletedAt string // 完了日時（ISO8601）
}

// ProjectState は分析に必要なプロジェクト状態
type ProjectState struct {
	Health  string       // プロジェクト健全性
	Summary SummaryStats // サマリー統計
}

// SummaryStats はサマリー統計（Activity 統計）
type SummaryStats struct {
	TotalActivities int // 全 Activity 数（JSON 互換のため "TotalActivities" を維持）
	Completed       int // 完了済み
	InProgress      int // 進行中
	Pending         int // 保留中
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

// AnalysisResult は全分析結果を集約
type AnalysisResult struct {
	Graph *DependencyGraph // 依存関係グラフ
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

// ステータス定数（Activity + Objective 共用）
const (
	// Activity ステータス
	TaskStatusDraft      = "draft"
	TaskStatusActive     = "active"
	TaskStatusDeprecated = "deprecated"
	// Objective ステータス
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusBlocked    = "blocked"
	TaskStatusOnHold     = "on_hold"
)

// ObjectiveInfo は分析に必要な Objective 情報
type ObjectiveInfo struct {
	ID        string // Objective ID
	Title     string // タイトル
	Status    string // ステータス
	CreatedAt string // 作成日時（ISO8601）
	UpdatedAt string // 更新日時（ISO8601）
}

// VisionInfo は分析に必要な Vision 情報
type VisionInfo struct {
	ID        string // Vision ID
	Title     string // タイトル
	Statement string // ステートメント
	Status    string // ステータス
}

// RiskInfo はリスク情報
// 使用コンテキスト:
//   - アフィニティ分析: Objective との関連付けによるクラスタリング
//
// core.RiskEntity からの変換時に handlers.go で生成される
type RiskInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Probability string `json:"probability"`  // low, medium, high
	Impact      string `json:"impact"`       // low, medium, high
	Score       int    `json:"score"`        // 計算されたスコア（Probability × Impact）
	Status      string `json:"status"`       // identified, mitigating, mitigated, accepted
	ObjectiveID string `json:"objective_id"` // 関連 Objective（Affinity クラスタリング用）
}

// QualityInfo は Quality エンティティ情報
// 使用コンテキスト:
//   - アフィニティ分析: Objective との関連付けによる品質基準クラスタリング
//
// core.QualityEntity からの変換時に handlers.go で生成される
// Note: QualityEntity には Status フィールドがないため、状態追跡が必要な場合は
// Gates の Pass/Fail や Metrics の達成度から派生させる
type QualityInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ObjectiveID string `json:"objective_id"` // 関連 Objective（必須）
}

// ============================================================
// UnifiedGraph 型定義 (Task/Activity 統合)
// ============================================================

// EntityType は統合グラフのエンティティタイプ
type EntityType string

const (
	EntityTypeActivity  EntityType = "activity"
	EntityTypeUseCase   EntityType = "usecase"
	EntityTypeObjective EntityType = "objective"
)

// ActivityInfo は分析に必要な Activity 情報
// core.ActivityEntity からの変換を前提とした軽量構造体
type ActivityInfo struct {
	ID        string // Activity ID
	Title     string // タイトル
	Status    string // ステータス
	UseCaseID string // 関連 UseCase ID
	CreatedAt string // 作成日時
	UpdatedAt string // 更新日時
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
	ID     string     // ノード ID
	Type   EntityType // エンティティタイプ
	Title  string     // タイトル
	Status string     // ステータス

	// 構造層（structural layer）情報
	StructuralParents  []string // 構造上の親ノード ID
	StructuralChildren []string // 構造上の子ノード ID
	StructuralDepth    int      // 構造層のルートからの深さ
}

// UnifiedEdgeLayer はエッジのレイヤー種別
type UnifiedEdgeLayer string

const (
	EdgeLayerStructural UnifiedEdgeLayer = "structural" // 階層構造を形成するレイヤー
)

// UnifiedEdgeRelation はエッジの意味論（関係種別）
type UnifiedEdgeRelation string

const (
	RelationImplements  UnifiedEdgeRelation = "implements"
	RelationContributes UnifiedEdgeRelation = "contributes"
)

// UnifiedEdge は統合グラフのエッジ（2層モデル）
type UnifiedEdge struct {
	From     string              // ソースノード ID
	To       string              // ターゲットノード ID
	Layer    UnifiedEdgeLayer    // レイヤー（structural）
	Relation UnifiedEdgeRelation // 関係種別
}

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
	TotalNodes          int                         // 総ノード数
	NodesByType         map[EntityType]int          // タイプ別ノード数
	TotalEdges          int                         // 総エッジ数
	EdgesByLayer        map[UnifiedEdgeLayer]int    // レイヤー別エッジ数
	EdgesByRelation     map[UnifiedEdgeRelation]int // 関係別エッジ数
	IsolatedCount       int                         // 孤立ノード数
	CycleCount          int                         // 循環依存の数
	MaxStructuralDepth  int                         // 構造層の最大深さ
	CompletedActivities int                         // 完了 Activity 数
	TotalActivities     int                         // 総 Activity 数
}

// GraphFilter はグラフ表示のフィルタリングオプション
type GraphFilter struct {
	FocusID          string                // 中心ノード ID
	FocusDepth       int                   // 中心からの深さ（0 = 無制限）
	IncludeTypes     []EntityType          // 含めるエンティティタイプ
	ExcludeTypes     []EntityType          // 除外するエンティティタイプ
	IncludeLayers    []UnifiedEdgeLayer    // 含めるエッジレイヤー
	IncludeRelations []UnifiedEdgeRelation // 含めるエッジ関係種別
	HideCompleted    bool                  // 完了済みを非表示
	HideDraft        bool                  // ドラフトを非表示
	HideUnrelated    bool                  // 無関係ノードを非表示
}

// NewGraphFilter はデフォルトのフィルタを作成
func NewGraphFilter() *GraphFilter {
	return &GraphFilter{
		FocusDepth:       0,
		IncludeTypes:     []EntityType{},
		ExcludeTypes:     []EntityType{},
		IncludeLayers:    []UnifiedEdgeLayer{},
		IncludeRelations: []UnifiedEdgeRelation{},
		HideCompleted:    false,
		HideDraft:        false,
		HideUnrelated:    false,
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

// WithIncludeLayers は含めるエッジレイヤーを設定
func (f *GraphFilter) WithIncludeLayers(layers ...UnifiedEdgeLayer) *GraphFilter {
	f.IncludeLayers = layers
	return f
}

// WithIncludeRelations は含めるエッジ関係種別を設定
func (f *GraphFilter) WithIncludeRelations(relations ...UnifiedEdgeRelation) *GraphFilter {
	f.IncludeRelations = relations
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
