import { writable, derived } from 'svelte/store';
import type { TasksResponse, TaskItem, TaskStatus } from '$lib/types/api';
import { fetchTasks } from '$lib/api/client';

// タスクデータのストア
export const tasksData = writable<TasksResponse | null>(null);

// タスクの読み込み状態
export const tasksLoading = writable<boolean>(false);

// タスクのエラー
export const tasksError = writable<string | null>(null);

// タスク一覧の派生ストア
export const tasks = derived(tasksData, ($data): TaskItem[] => {
	return $data?.tasks ?? [];
});

// タスク総数の派生ストア
export const totalTasks = derived(tasksData, ($data): number => {
	return $data?.total ?? 0;
});

// ステータス別タスク数の派生ストア
export const tasksByStatus = derived(tasks, ($tasks): Record<TaskStatus, number> => {
	const counts: Record<TaskStatus, number> = {
		completed: 0,
		in_progress: 0,
		pending: 0,
		blocked: 0
	};

	for (const task of $tasks) {
		const status = task.status as TaskStatus;
		if (status in counts) {
			counts[status]++;
		}
	}

	return counts;
});

// フィルタリング用の派生ストア
export function getTasksByStatus(status: TaskStatus) {
	return derived(tasks, ($tasks) => {
		return $tasks.filter((task) => task.status === status);
	});
}

// タスクを更新
export async function refreshTasks(): Promise<void> {
	tasksLoading.set(true);
	tasksError.set(null);

	try {
		const data = await fetchTasks();
		tasksData.set(data);
	} catch (err) {
		const message = err instanceof Error ? err.message : 'Unknown error';
		tasksError.set(message);
	} finally {
		tasksLoading.set(false);
	}
}

// タスクを直接設定（SSE 用）
export function setTasks(data: TasksResponse): void {
	tasksData.set(data);
	tasksError.set(null);
}
