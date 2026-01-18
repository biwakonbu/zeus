/**
 * パフォーマンステスト用ヘルパー
 *
 * 非機能要件として閾値を定義し、パフォーマンス計測を行う
 */
import { expect } from 'vitest';

/**
 * パフォーマンス計測結果
 */
export interface PerformanceResult {
	// 計測時間（ミリ秒）
	duration: number;
	// イテレーション数
	iterations: number;
	// 1イテレーションあたりの平均時間
	avgPerIteration: number;
	// 最小時間
	min: number;
	// 最大時間
	max: number;
	// 標準偏差
	stdDev: number;
}

/**
 * 非機能要件の閾値定義
 *
 * 各操作の許容時間（ミリ秒）を定義
 * これを下回ると失敗となる
 */
export const PERFORMANCE_THRESHOLDS = {
	// LayoutEngine
	layout: {
		// 100ノードのレイアウト計算（キャッシュなし）
		small: 50, // ms
		// 500ノードのレイアウト計算（キャッシュなし）
		medium: 200, // ms
		// 1000ノードのレイアウト計算（キャッシュなし）
		large: 500, // ms
		// キャッシュヒット時（構造変更なし）
		cacheHit: 1 // ms
	},

	// SpatialIndex
	spatialIndex: {
		// 1000ノードの挿入
		insert1000: 50, // ms
		// 5000ノードの挿入
		insert5000: 200, // ms
		// ビューポートクエリ（1000ノード中）
		query: 5, // ms
		// ポイントクエリ
		pointQuery: 1 // ms
	},

	// EdgeFactory
	edgeFactory: {
		// 1000エッジの作成
		create1000: 50, // ms
		// ノードに関連するエッジの取得（インデックス使用）
		getEdgesForNode: 1, // ms
		// エッジの削除
		remove: 1 // ms
	},

	// 差分更新（FactorioViewer のロジック部分）
	diffUpdate: {
		// タスクハッシュ計算（1000タスク）
		computeHash: 10, // ms
		// 変更検出
		detectChanges: 5 // ms
	}
} as const;

/**
 * パフォーマンスを計測
 *
 * @param fn 計測対象の関数
 * @param iterations イテレーション数
 * @param warmupIterations ウォームアップイテレーション数
 * @returns 計測結果
 */
export function measurePerformance(
	fn: () => void,
	iterations = 100,
	warmupIterations = 10
): PerformanceResult {
	// ウォームアップ（JIT コンパイル等を安定させる）
	for (let i = 0; i < warmupIterations; i++) {
		fn();
	}

	const times: number[] = [];

	// 本計測
	for (let i = 0; i < iterations; i++) {
		const start = performance.now();
		fn();
		const end = performance.now();
		times.push(end - start);
	}

	const total = times.reduce((a, b) => a + b, 0);
	const avg = total / iterations;
	const min = Math.min(...times);
	const max = Math.max(...times);

	// 標準偏差
	const variance = times.reduce((sum, t) => sum + Math.pow(t - avg, 2), 0) / iterations;
	const stdDev = Math.sqrt(variance);

	return {
		duration: total,
		iterations,
		avgPerIteration: avg,
		min,
		max,
		stdDev
	};
}

/**
 * 単一実行のパフォーマンスを計測
 */
export function measureOnce(fn: () => void): number {
	const start = performance.now();
	fn();
	return performance.now() - start;
}

/**
 * パフォーマンス閾値をアサート
 *
 * @param result 計測結果
 * @param threshold 閾値（ミリ秒）
 * @param label テスト名（エラーメッセージ用）
 */
export function assertPerformance(
	result: PerformanceResult,
	threshold: number,
	label: string
): void {
	expect(
		result.avgPerIteration,
		`${label}: 平均実行時間 ${result.avgPerIteration.toFixed(3)}ms が閾値 ${threshold}ms を超過`
	).toBeLessThanOrEqual(threshold);
}

/**
 * 単一実行のパフォーマンス閾値をアサート
 */
export function assertPerformanceOnce(duration: number, threshold: number, label: string): void {
	expect(
		duration,
		`${label}: 実行時間 ${duration.toFixed(3)}ms が閾値 ${threshold}ms を超過`
	).toBeLessThanOrEqual(threshold);
}

/**
 * テスト用のタスクデータを生成
 */
export interface MockTaskItem {
	id: string;
	title: string;
	status: 'pending' | 'in_progress' | 'completed' | 'blocked';
	progress?: number;
	priority?: 'critical' | 'high' | 'medium' | 'low';
	assignee?: string;
	dependencies: string[];
}

/**
 * ランダムなタスクリストを生成
 *
 * @param count タスク数
 * @param maxDependencies 最大依存数
 */
export function generateMockTasks(count: number, maxDependencies = 3): MockTaskItem[] {
	const statuses: MockTaskItem['status'][] = ['pending', 'in_progress', 'completed', 'blocked'];
	const priorities: NonNullable<MockTaskItem['priority']>[] = [
		'critical',
		'high',
		'medium',
		'low'
	];

	const tasks: MockTaskItem[] = [];

	for (let i = 0; i < count; i++) {
		const id = `task-${i}`;

		// 自分より前のタスクから依存を選択
		const dependencies: string[] = [];
		const numDeps = Math.min(i, Math.floor(Math.random() * (maxDependencies + 1)));
		for (let j = 0; j < numDeps; j++) {
			const depIndex = Math.floor(Math.random() * i);
			const depId = `task-${depIndex}`;
			if (!dependencies.includes(depId)) {
				dependencies.push(depId);
			}
		}

		tasks.push({
			id,
			title: `Task ${i}`,
			status: statuses[Math.floor(Math.random() * statuses.length)],
			progress: Math.floor(Math.random() * 101),
			priority: priorities[Math.floor(Math.random() * priorities.length)],
			assignee: `user-${Math.floor(Math.random() * 5)}`,
			dependencies
		});
	}

	return tasks;
}

/**
 * パフォーマンス結果をフォーマットして表示
 */
export function formatPerformanceResult(label: string, result: PerformanceResult): string {
	return `${label}:
    平均: ${result.avgPerIteration.toFixed(3)}ms
    最小: ${result.min.toFixed(3)}ms
    最大: ${result.max.toFixed(3)}ms
    標準偏差: ${result.stdDev.toFixed(3)}ms
    イテレーション: ${result.iterations}回`;
}
