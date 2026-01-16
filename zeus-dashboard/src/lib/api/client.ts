// API クライアント
import type {
	StatusResponse,
	TasksResponse,
	GraphResponse,
	PredictResponse,
	WBSResponse,
	TimelineResponse,
	ErrorResponse
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
