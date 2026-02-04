// Stores のエクスポート
export * from './connection';
export * from './status';
export * from './tasks';
export * from './view';

// 全データを一括更新
import { refreshStatus } from './status';
import { refreshTasks } from './tasks';

export async function refreshAllData(): Promise<void> {
	await Promise.all([refreshStatus(), refreshTasks()]);
}
