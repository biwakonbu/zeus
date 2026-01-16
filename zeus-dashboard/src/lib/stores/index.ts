// Stores のエクスポート
export * from './connection';
export * from './status';
export * from './tasks';
export * from './graph';
export * from './prediction';

// 全データを一括更新
import { refreshStatus } from './status';
import { refreshTasks } from './tasks';
import { refreshGraph } from './graph';
import { refreshPrediction } from './prediction';

export async function refreshAllData(): Promise<void> {
	await Promise.all([refreshStatus(), refreshTasks(), refreshGraph(), refreshPrediction()]);
}
