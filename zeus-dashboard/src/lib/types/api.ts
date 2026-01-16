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
	summary: TaskStats;
}

export type HealthStatus = 'good' | 'fair' | 'poor';

export interface TaskStats {
	total_tasks: number;
	completed: number;
	in_progress: number;
	pending: number;
}

// タスク API レスポンス
export interface TasksResponse {
	tasks: TaskItem[];
	total: number;
}

export interface TaskItem {
	id: string;
	title: string;
	status: TaskStatus;
	priority: Priority;
	assignee: string;
	dependencies: string[];

	// Phase 6A: WBS・タイムライン機能用フィールド
	parent_id?: string;
	start_date?: string;
	due_date?: string;
	progress: number;
	wbs_code?: string;
}

export type TaskStatus = 'completed' | 'in_progress' | 'pending' | 'blocked';
export type Priority = 'high' | 'medium' | 'low';

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

// 予測 API レスポンス
export interface PredictResponse {
	completion?: CompletionPrediction;
	risk?: RiskPrediction;
	velocity?: VelocityReport;
}

export interface CompletionPrediction {
	remaining_tasks: number;
	average_velocity: number;
	estimated_date: string;
	confidence_level: number;
	margin_days: number;
	has_sufficient_data: boolean;
}

export interface RiskPrediction {
	overall_level: RiskLevel;
	factors: RiskFactor[];
	score: number;
}

export type RiskLevel = 'low' | 'medium' | 'high' | 'critical';

export interface RiskFactor {
	name: string;
	description: string;
	impact: number;
}

export interface VelocityReport {
	last_7_days: number;
	last_14_days: number;
	last_30_days: number;
	weekly_average: number;
	trend: VelocityTrend;
	data_points: number;
}

export type VelocityTrend = 'increasing' | 'stable' | 'decreasing' | 'insufficient_data';

// SSE イベント型
export type SSEEventType = 'status' | 'task' | 'approval' | 'graph' | 'prediction';

export interface SSEEvent<T = unknown> {
	type: SSEEventType;
	data: T;
}

// エラーレスポンス
export interface ErrorResponse {
	error: string;
	message: string;
}

// Phase 6B: WBS API レスポンス
export interface WBSResponse {
	roots: WBSNode[];
	max_depth: number;
	stats: WBSStats;
}

export interface WBSNode {
	id: string;
	title: string;
	wbs_code: string;
	status: TaskStatus;
	progress: number;
	priority: Priority;
	assignee: string;
	children?: WBSNode[];
	depth: number;
}

export interface WBSStats {
	total_nodes: number;
	root_count: number;
	leaf_count: number;
	max_depth: number;
	avg_progress: number;
	completed_pct: number;
}

// Phase 6C: タイムライン API レスポンス
export interface TimelineResponse {
	items: TimelineItem[];
	critical_path: string[];
	project_start: string;
	project_end: string;
	total_duration: number;
	stats: TimelineStats;
}

export interface TimelineItem {
	task_id: string;
	title: string;
	start_date: string;
	end_date: string;
	progress: number;
	status: TaskStatus;
	priority: Priority;
	assignee: string;
	is_on_critical_path: boolean;
	slack: number | null;
	dependencies: string[];
}

export interface TimelineStats {
	total_tasks: number;
	tasks_with_dates: number;
	on_critical_path: number;
	average_slack: number;
	overdue_tasks: number;
	completed_on_time: number;
}
