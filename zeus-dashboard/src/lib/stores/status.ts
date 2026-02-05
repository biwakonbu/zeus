import { writable, derived } from 'svelte/store';
import type { StatusResponse, ProjectInfo, ProjectState, HealthStatus } from '$lib/types/api';
import { fetchStatus } from '$lib/api/client';

// ステータスデータのストア
export const statusData = writable<StatusResponse | null>(null);

// ステータスの読み込み状態
export const statusLoading = writable<boolean>(false);

// ステータスのエラー
export const statusError = writable<string | null>(null);

// プロジェクト情報の派生ストア
export const projectInfo = derived(statusData, ($status): ProjectInfo | null => {
	return $status?.project ?? null;
});

// プロジェクト状態の派生ストア
export const projectState = derived(statusData, ($status): ProjectState | null => {
	return $status?.state ?? null;
});

// 健全性の派生ストア
export const health = derived(projectState, ($state): HealthStatus | null => {
	return ($state?.health as HealthStatus) ?? null;
});

// 承認待ち件数の派生ストア
export const pendingApprovals = derived(statusData, ($status): number => {
	return $status?.pending_approvals ?? 0;
});

// 進捗率の派生ストア
export const progressPercent = derived(projectState, ($state): number => {
	if (!$state?.summary || $state.summary.total_activities === 0) {
		return 0;
	}
	return Math.round(($state.summary.completed / $state.summary.total_activities) * 100);
});

// ステータスを更新
export async function refreshStatus(): Promise<void> {
	statusLoading.set(true);
	statusError.set(null);

	try {
		const data = await fetchStatus();
		statusData.set(data);
	} catch (err) {
		const message = err instanceof Error ? err.message : 'Unknown error';
		statusError.set(message);
	} finally {
		statusLoading.set(false);
	}
}

// ステータスを直接設定（SSE 用）
export function setStatus(data: StatusResponse): void {
	statusData.set(data);
	statusError.set(null);
}
