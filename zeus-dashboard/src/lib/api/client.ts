// API クライアント
import type {
	StatusResponse,
	GraphResponse,
	ErrorResponse,
	ActorsResponse,
	UseCasesResponse,
	UseCaseDiagramResponse,
	ActivitiesResponse,
	ActivityDiagramResponse,
	SubsystemsResponse,
	UnifiedGraphResponse
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

// グラフ取得
export async function fetchGraph(): Promise<GraphResponse> {
	return fetchJSON<GraphResponse>('/graph');
}

// UnifiedGraph 取得（Activity, UseCase, Objective の統合グラフ）
export async function fetchUnifiedGraph(): Promise<UnifiedGraphResponse> {
	return fetchJSON<UnifiedGraphResponse>('/unified-graph');
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
// UML Subsystem API（TASK-023）
// =============================================================================

// サブシステム一覧取得
export async function fetchSubsystems(): Promise<SubsystemsResponse> {
	return fetchJSON<SubsystemsResponse>('/subsystems');
}

// =============================================================================
// UML Activity API
// =============================================================================

// アクティビティ一覧取得
export async function fetchActivities(): Promise<ActivitiesResponse> {
	return fetchJSON<ActivitiesResponse>('/activities');
}

// アクティビティ図取得
export async function fetchActivityDiagram(activityId: string): Promise<ActivityDiagramResponse> {
	return fetchJSON<ActivityDiagramResponse>(`/uml/activity?id=${encodeURIComponent(activityId)}`);
}

// 全データ取得（並列実行）
export interface DashboardData {
	status: StatusResponse | null;
	graph: GraphResponse | null;
}

export async function fetchAllData(): Promise<DashboardData> {
	const results = await Promise.allSettled([fetchStatus(), fetchGraph()]);

	return {
		status: results[0].status === 'fulfilled' ? results[0].value : null,
		graph: results[1].status === 'fulfilled' ? results[1].value : null
	};
}
