// API クライアント
import type {
	StatusResponse,
	TasksResponse,
	GraphResponse,
	PredictResponse,
	WBSResponse,
	WBSNode,
	TimelineResponse,
	DownstreamResponse,
	ErrorResponse,
	GraphNode,
	GraphEdge,
	WBSGraphData,
	ActorsResponse,
	UseCasesResponse,
	UseCaseDiagramResponse
} from '$lib/types/api';

// API ベース URL（開発時は Vite Proxy 経由、本番時は同一オリジン）
const API_BASE = '/api';

// カスタムエラークラス
export class APIError extends Error {
	constructor(
		public status: number,
		public response: ErrorResponse
	) {
		super(response.message);
		this.name = 'APIError';
	}
}

// 共通 fetch ラッパー
async function fetchJSON<T>(endpoint: string): Promise<T> {
	const url = `${API_BASE}${endpoint}`;
	const response = await fetch(url, {
		method: 'GET',
		headers: {
			Accept: 'application/json'
		}
	});

	if (!response.ok) {
		let errorResponse: ErrorResponse;
		try {
			errorResponse = await response.json();
		} catch {
			errorResponse = {
				error: response.statusText,
				message: `HTTP ${response.status}: ${response.statusText}`
			};
		}
		throw new APIError(response.status, errorResponse);
	}

	return response.json();
}

// ステータス取得
export async function fetchStatus(): Promise<StatusResponse> {
	return fetchJSON<StatusResponse>('/status');
}

// タスク一覧取得
export async function fetchTasks(): Promise<TasksResponse> {
	return fetchJSON<TasksResponse>('/tasks');
}

// グラフ取得
export async function fetchGraph(): Promise<GraphResponse> {
	return fetchJSON<GraphResponse>('/graph');
}

// 予測取得
export async function fetchPredict(): Promise<PredictResponse> {
	return fetchJSON<PredictResponse>('/predict');
}

// WBS 取得
export async function fetchWBS(): Promise<WBSResponse> {
	return fetchJSON<WBSResponse>('/wbs');
}

// タイムライン取得
export async function fetchTimeline(): Promise<TimelineResponse> {
	return fetchJSON<TimelineResponse>('/timeline');
}

// 下流タスク取得（影響範囲の可視化用）
export async function fetchDownstream(taskId: string): Promise<DownstreamResponse> {
	return fetchJSON<DownstreamResponse>(`/downstream?task_id=${encodeURIComponent(taskId)}`);
}


// =============================================================================
// UML UseCase API
// =============================================================================

// アクター一覧取得
export async function fetchActors(): Promise<ActorsResponse> {
	return fetchJSON<ActorsResponse>('/actors');
}

// ユースケース一覧取得
export async function fetchUseCases(): Promise<UseCasesResponse> {
	return fetchJSON<UseCasesResponse>('/usecases');
}

// ユースケース図取得
export async function fetchUseCaseDiagram(boundary?: string): Promise<UseCaseDiagramResponse> {
	const params = boundary ? `?boundary=${encodeURIComponent(boundary)}` : '';
	return fetchJSON<UseCaseDiagramResponse>(`/uml/usecase${params}`);
}

// =============================================================================
// WBS → GraphData 変換ユーティリティ
// =============================================================================

/**
 * WBS 階層データをフラットな GraphNode/GraphEdge に変換
 * @param wbs WBSResponse (階層構造)
 * @returns WBSGraphData (フラットなノード・エッジ)
 */
export function convertWBSToGraphData(wbs: WBSResponse): WBSGraphData {
	const nodes: GraphNode[] = [];
	const edges: GraphEdge[] = [];

	/**
	 * 再帰的にノードを抽出し、親子関係をエッジとして記録
	 */
	function traverse(node: WBSNode, parentId: string | null): void {
		// GraphNode に変換
		const graphNode: GraphNode = {
			id: node.id,
			title: node.title,
			node_type: node.node_type,
			status: node.status,
			progress: node.progress,
			priority: node.priority || undefined,
			assignee: node.assignee || undefined,
			wbs_code: node.wbs_code || undefined,
			dependencies: parentId ? [parentId] : []
		};
		nodes.push(graphNode);

		// 親子関係をエッジとして記録（親 → 子）
		if (parentId) {
			edges.push({
				from: parentId,
				to: node.id
			});
		}

		// 子ノードを再帰処理
		if (node.children && node.children.length > 0) {
			for (const child of node.children) {
				traverse(child, node.id);
			}
		}
	}

	// ルートノードから開始
	for (const root of wbs.roots) {
		traverse(root, null);
	}

	return { nodes, edges };
}

/**
 * WBS データを取得して GraphData に変換
 * @returns WBSGraphData
 */
export async function fetchWBSAsGraphData(): Promise<WBSGraphData> {
	const wbs = await fetchWBS();
	return convertWBSToGraphData(wbs);
}

// 全データ取得（並列実行）
export interface DashboardData {
	status: StatusResponse | null;
	tasks: TasksResponse | null;
	graph: GraphResponse | null;
	predict: PredictResponse | null;
}

export async function fetchAllData(): Promise<DashboardData> {
	const results = await Promise.allSettled([
		fetchStatus(),
		fetchTasks(),
		fetchGraph(),
		fetchPredict()
	]);

	return {
		status: results[0].status === 'fulfilled' ? results[0].value : null,
		tasks: results[1].status === 'fulfilled' ? results[1].value : null,
		graph: results[2].status === 'fulfilled' ? results[2].value : null,
		predict: results[3].status === 'fulfilled' ? results[3].value : null
	};
}
