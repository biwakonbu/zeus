// Stores のエクスポート
export * from './connection';
export * from './status';
export * from './view';

// 全データを一括更新
import { refreshStatus } from './status';

export async function refreshAllData(): Promise<void> {
	await refreshStatus();
}
