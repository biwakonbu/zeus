// API レスポンス型定義（Go の型と同期）

// 接続状態
export type ConnectionState = 'connected' | 'connecting' | 'disconnected';

// ステータス API レスポンス
export interface StatusResponse {
	project: ProjectInfo;
	state: ProjectState;
	pending_approvals: number;
}

export interface ProjectInfo {
	id: string;
	name: string;
	description: string;
	start_date: string;
}

export interface ProjectState {
	health: HealthStatus;
	summary: SummaryStats;
}

export type HealthStatus = 'good' | 'fair' | 'poor';

export interface SummaryStats {
	total_activities: number;
	completed: number;
	in_progress: number;
	pending: number;
}

// EntityStatus は各ビューで使用される共通型
export type EntityStatus = 'draft' | 'active' | 'deprecated';

// グラフ API レスポンス
export interface GraphResponse {
	mermaid: string;
	stats: GraphStats;
	cycles: string[][];
	isolated: string[];
}

export interface GraphStats {
	total_nodes: number;
	with_dependencies: number;
	isolated_count: number;
	cycle_count: number;
	max_depth: number;
}

// SSE イベント型
export type SSEEventType = 'status' | 'approval' | 'graph' | 'prediction';

export interface SSEEvent<T = unknown> {
	type: SSEEventType;
	data: T;
}

// エラーレスポンス
export interface ErrorResponse {
	error: string;
	message: string;
}

// =============================================================================
// 10概念モデル API レスポンス
// =============================================================================

// Vision
export interface VisionResponse {
	vision: Vision | null;
}

export interface Vision {
	title: string;
	statement: string;
	created_at: string;
	updated_at: string;
}

// Objective
export interface ObjectivesResponse {
	objectives: Objective[];
	total: number;
}

export interface Objective {
	id: string;
	title: string;
	description?: string;
	status: ObjectiveStatus;
	parent_id?: string;
	created_at: string;
	updated_at: string;
}

export type ObjectiveStatus = 'not_started' | 'in_progress' | 'completed' | 'on_hold';

// Consideration
export interface ConsiderationsResponse {
	considerations: Consideration[];
	total: number;
}

export interface Consideration {
	id: string;
	title: string;
	description?: string;
	status: ConsiderationStatus;
	objective_id?: string;
	options: ConsiderationOption[];
	due_date?: string;
	decision_id?: string;
	created_at: string;
	updated_at: string;
}

export interface ConsiderationOption {
	id: string;
	title: string;
	description?: string;
	pros?: string[];
	cons?: string[];
}

export type ConsiderationStatus = 'open' | 'decided' | 'deferred';

// Decision
export interface DecisionsResponse {
	decisions: Decision[];
	total: number;
}

export interface Decision {
	id: string;
	title: string;
	consideration_id: string;
	selected_option: {
		id: string;
		title: string;
	};
	rationale: string;
	decided_at: string;
	created_at: string;
}

// Problem
export interface ProblemsResponse {
	problems: Problem[];
	total: number;
}

export interface Problem {
	id: string;
	title: string;
	description?: string;
	severity: Severity;
	status: ProblemStatus;
	objective_id?: string;
	created_at: string;
	updated_at: string;
}

export type Severity = 'low' | 'medium' | 'high' | 'critical';
export type ProblemStatus = 'open' | 'investigating' | 'resolved' | 'wont_fix';

// Risk
export interface RisksResponse {
	risks: Risk[];
	total: number;
}

export interface Risk {
	id: string;
	title: string;
	description?: string;
	probability: RiskProbability;
	impact: RiskImpact;
	score: number;
	status: RiskStatus;
	objective_id?: string;
	mitigation?: string;
	created_at: string;
	updated_at: string;
}

export type RiskProbability = 'low' | 'medium' | 'high';
export type RiskImpact = 'low' | 'medium' | 'high';
export type RiskStatus = 'identified' | 'mitigating' | 'mitigated' | 'accepted';

// Assumption
export interface AssumptionsResponse {
	assumptions: Assumption[];
	total: number;
}

export interface Assumption {
	id: string;
	title: string;
	description?: string;
	status: AssumptionStatus;
	objective_id?: string;
	verified_at?: string;
	created_at: string;
	updated_at: string;
}

export type AssumptionStatus = 'unverified' | 'verified' | 'invalid';

// Constraint
export interface ConstraintsResponse {
	constraints: Constraint[];
	total: number;
}

export interface Constraint {
	id: string;
	title: string;
	description?: string;
	category: ConstraintCategory;
	non_negotiable: boolean;
	created_at: string;
	updated_at: string;
}

export type ConstraintCategory = 'technical' | 'business' | 'legal' | 'resource' | 'time';

// Quality
export interface QualityResponse {
	quality_items: QualityItem[];
	total: number;
}

export interface QualityItem {
	id: string;
	title: string;
	description?: string;
	objective_id: string;
	metric?: QualityMetric;
	gate?: QualityGate;
	status: QualityStatus;
	created_at: string;
	updated_at: string;
}

export interface QualityMetric {
	name: string;
	target: number;
	current?: number;
	unit: string;
}

export interface QualityGate {
	name: string;
	passed: boolean;
	checked_at?: string;
}

export type QualityStatus = 'not_checked' | 'passing' | 'failing';

// =============================================================================
// グラフビュー用統一ノード型
// =============================================================================

// グラフノードの種別（UnifiedGraph 対応）
// Note: 'task' は Activity に統合されたため除去。色・ラベル定義は $lib/viewer/config/nodeTypes.ts で一元管理。
export type GraphNodeType = 'vision' | 'activity' | 'usecase';

// グラフビュー用の統一ノードデータ
export interface GraphNode {
	id: string;
	title: string;
	node_type: GraphNodeType;
	status: string;
	structural_depth?: number;
}

export type GraphEdgeLayer = 'structural';
export type GraphEdgeRelation =
	| 'parent'
	| 'implements';

// グラフビュー用のエッジデータ
export interface GraphEdge {
	from: string;
	to: string;
	layer: GraphEdgeLayer;
	relation: GraphEdgeRelation;
}


// =============================================================================
// UML Subsystem API レスポンス（TASK-017）
// =============================================================================

// サブシステム
export interface SubsystemItem {
	id: string;
	name: string;
	description?: string;
}

// サブシステム一覧 API レスポンス
export interface SubsystemsResponse {
	subsystems: SubsystemItem[];
	total: number;
}

// =============================================================================
// UML UseCase API レスポンス
// =============================================================================

// アクタータイプ
export type ActorType = 'human' | 'system' | 'time' | 'device' | 'external';

// アクター
export interface ActorItem {
	id: string;
	title: string;
	type: ActorType;
	description?: string;
}

// アクター一覧 API レスポンス
export interface ActorsResponse {
	actors: ActorItem[];
	total: number;
}

// アクターロール
export type ActorRole = 'primary' | 'secondary';

// ユースケース - アクター参照
export interface UseCaseActorRef {
	actor_id: string;
	role: ActorRole;
}

// リレーションタイプ
export type RelationType = 'include' | 'extend' | 'generalize';

// ユースケースリレーション
export interface UseCaseRelation {
	type: RelationType;
	target_id: string;
	condition?: string;
	extension_point?: string;
}

// ユースケースステータス
export type UseCaseStatus = 'draft' | 'active' | 'deprecated';

/** 代替フロー（UML 2.5 準拠） */
export interface AlternativeFlow {
	id: string;
	name: string;
	condition: string;
	steps: string[];
	/** メインフローに戻るステップ */
	rejoins_at?: string;
}

/** 例外フロー（UML 2.5 準拠） */
export interface ExceptionFlow {
	id: string;
	name: string;
	/** 例外発生条件 */
	trigger: string;
	steps: string[];
	/** 結果（例: "ステップ2へ戻る"） */
	outcome?: string;
}

/** シナリオ（UML 2.5 準拠の拡張版） */
export interface UseCaseScenario {
	/** 事前条件 */
	preconditions?: string[];
	/** 開始イベント */
	trigger?: string;
	/** メインフロー */
	main_flow?: string[];
	/** 代替フロー */
	alternative_flows?: AlternativeFlow[];
	/** 例外フロー */
	exception_flows?: ExceptionFlow[];
	/** 事後条件 */
	postconditions?: string[];
}

// ユースケース（TASK-018: subsystem_id 追加）
export interface UseCaseItem {
	id: string;
	title: string;
	description?: string;
	status: UseCaseStatus;
	objective_id?: string;
	subsystem_id?: string; // サブシステム参照（オプション）
	actors: UseCaseActorRef[];
	relations: UseCaseRelation[];
	scenario?: UseCaseScenario;
}

// ユースケース一覧 API レスポンス
export interface UseCasesResponse {
	usecases: UseCaseItem[];
	total: number;
}

// ユースケース図 API レスポンス
export interface UseCaseDiagramResponse {
	actors: ActorItem[];
	usecases: UseCaseItem[];
	boundary: string;
	mermaid: string;
}

// =============================================================================
// UML Activity API レスポンス
// =============================================================================

// アクティビティノードタイプ
export type ActivityNodeType =
	| 'initial' // 開始ノード（黒丸）
	| 'final' // 終了ノード（二重丸）
	| 'action' // アクション（角丸四角形）
	| 'decision' // 分岐（ひし形）
	| 'merge' // 合流（ひし形）
	| 'fork' // 並列分岐（太い横線）
	| 'join'; // 並列合流（太い横線）

// アクティビティステータス
export type ActivityStatus = 'draft' | 'active' | 'deprecated';

// アクティビティノード
export interface ActivityNodeItem {
	id: string;
	type: ActivityNodeType;
	name?: string;
}

// アクティビティ遷移
export interface ActivityTransitionItem {
	id: string;
	source: string;
	target: string;
	guard?: string;
}

// アクティビティ
export interface ActivityItem {
	id: string;
	title: string;
	description?: string;
	usecase_id?: string;
	usecase_title?: string;
	status: ActivityStatus;
	nodes: ActivityNodeItem[];
	transitions: ActivityTransitionItem[];
	created_at: string;
	updated_at: string;
}

// アクティビティ一覧 API レスポンス
export interface ActivitiesResponse {
	activities: ActivityItem[];
	total: number;
}

// アクティビティ図 API レスポンス
export interface ActivityDiagramResponse {
	activity?: ActivityItem;
	mermaid: string;
}

// =============================================================================
// UnifiedGraph API レスポンス（Graph View 用）
// =============================================================================

// UnifiedGraph ノード
export interface UnifiedGraphNodeItem {
	id: string;
	type: string;
	title: string;
	status: string;
	structural_depth: number;
	structural_parents?: string[];
	structural_children?: string[];
}

// UnifiedGraph エッジ
export interface UnifiedGraphEdgeItem {
	source: string;
	target: string;
	layer: GraphEdgeLayer;
	relation: GraphEdgeRelation;
}

// UnifiedGraph 統計情報
export interface UnifiedGraphStats {
	total_nodes: number;
	total_edges: number;
	total_activities: number;
	completed_activities: number;
	max_structural_depth: number;
	cycle_count: number;
	isolated_count: number;
	nodes_by_type: Record<string, number>;
	edges_by_layer: Record<string, number>;
	edges_by_relation: Record<string, number>;
}

// UnifiedGraph グループ（Objective ベース）
export interface UnifiedGraphGroupItem {
	id: string;
	title: string;
	description?: string;
	goals?: string[];
	status: string;
	owner?: string;
	tags?: string[];
	node_ids: string[];
}

// UnifiedGraph API レスポンス
export interface UnifiedGraphResponse {
	nodes: UnifiedGraphNodeItem[];
	edges: UnifiedGraphEdgeItem[];
	groups: UnifiedGraphGroupItem[];
	stats: UnifiedGraphStats;
	cycles: string[][];
	isolated: string[];
	mermaid: string;
}

// グラフデータ型（GraphNode/Edge + Groups の組み合わせ）
export interface GraphData {
	nodes: GraphNode[];
	edges: GraphEdge[];
	groups?: UnifiedGraphGroupItem[];
}
