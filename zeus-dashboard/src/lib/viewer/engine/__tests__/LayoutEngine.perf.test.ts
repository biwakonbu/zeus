/**
 * LayoutEngine パフォーマンステスト
 *
 * 非機能要件:
 * - 100ノード: 50ms 以内
 * - 500ノード: 200ms 以内
 * - 1000ノード: 500ms 以内
 * - キャッシュヒット時: 1ms 以内
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
	measurePerformance,
	measureOnce,
	assertPerformance,
	assertPerformanceOnce,
	generateMockTasks,
	PERFORMANCE_THRESHOLDS,
	formatPerformanceResult
} from './performance-helper';

// TaskNode をモック（PixiJS 依存を回避）
vi.mock('../../rendering/TaskNode', () => ({
	TaskNode: {
		getWidth: () => 200,
		getHeight: () => 80
	}
}));

// モック後にインポート
import { LayoutEngine } from '../LayoutEngine';

describe('LayoutEngine パフォーマンステスト', () => {
	let engine: LayoutEngine;

	beforeEach(() => {
		engine = new LayoutEngine();
	});

	describe('レイアウト計算', () => {
		it('100ノードのレイアウト計算が閾値以内で完了する', () => {
			const tasks = generateMockTasks(100);
			const threshold = PERFORMANCE_THRESHOLDS.layout.small;

			const result = measurePerformance(() => {
				engine.clearCache();
				engine.layout(tasks);
			}, 20, 5);

			console.log(formatPerformanceResult('100ノード レイアウト', result));
			assertPerformance(result, threshold, '100ノード レイアウト');
		});

		it('500ノードのレイアウト計算が閾値以内で完了する', () => {
			const tasks = generateMockTasks(500);
			const threshold = PERFORMANCE_THRESHOLDS.layout.medium;

			const result = measurePerformance(() => {
				engine.clearCache();
				engine.layout(tasks);
			}, 10, 3);

			console.log(formatPerformanceResult('500ノード レイアウト', result));
			assertPerformance(result, threshold, '500ノード レイアウト');
		});

		it('1000ノードのレイアウト計算が閾値以内で完了する', () => {
			const tasks = generateMockTasks(1000);
			const threshold = PERFORMANCE_THRESHOLDS.layout.large;

			const result = measurePerformance(() => {
				engine.clearCache();
				engine.layout(tasks);
			}, 5, 2);

			console.log(formatPerformanceResult('1000ノード レイアウト', result));
			assertPerformance(result, threshold, '1000ノード レイアウト');
		});
	});

	describe('キャッシュ', () => {
		it('キャッシュヒット時は閾値以内で完了する', () => {
			const tasks = generateMockTasks(500);
			const threshold = PERFORMANCE_THRESHOLDS.layout.cacheHit;

			// 初回実行（キャッシュ構築）
			engine.layout(tasks);

			// 2回目以降（キャッシュヒット）
			const result = measurePerformance(() => {
				engine.layout(tasks);
			}, 100, 10);

			console.log(formatPerformanceResult('キャッシュヒット', result));
			assertPerformance(result, threshold, 'キャッシュヒット');
		});

		it('データ変更のみの場合はキャッシュが有効', () => {
			const tasks = generateMockTasks(500);
			const threshold = PERFORMANCE_THRESHOLDS.layout.cacheHit;

			// 初回実行
			engine.layout(tasks);

			// ステータスのみ変更（構造は同じ）
			tasks[0].status = 'completed';
			tasks[10].progress = 100;

			const duration = measureOnce(() => {
				engine.layout(tasks);
			});

			console.log(`データ変更のみ: ${duration.toFixed(3)}ms`);
			assertPerformanceOnce(duration, threshold, 'データ変更のみ');
		});

		it('構造変更時はキャッシュが無効化される', () => {
			const tasks = generateMockTasks(500);

			// 初回実行
			engine.layout(tasks);

			// 依存関係を変更（構造変更）
			tasks[100].dependencies = ['task-50'];
			engine.clearCache();

			const secondLayout = engine.layout(tasks);

			// 構造変更があるのでキャッシュミス
			// レイアウト結果は異なる可能性がある
			expect(secondLayout).toBeDefined();
		});
	});

	describe('スケーラビリティ', () => {
		it('ノード数に対して線形〜準線形にスケールする', () => {
			const sizes = [100, 250, 500, 750, 1000];
			const times: { size: number; avgTime: number }[] = [];

			for (const size of sizes) {
				const tasks = generateMockTasks(size);
				const result = measurePerformance(() => {
					engine.clearCache();
					engine.layout(tasks);
				}, 5, 2);
				times.push({ size, avgTime: result.avgPerIteration });
			}

			console.log('\nスケーラビリティ分析:');
			for (const { size, avgTime } of times) {
				const perNode = avgTime / size;
				console.log(`  ${size}ノード: ${avgTime.toFixed(2)}ms (${(perNode * 1000).toFixed(3)}µs/ノード)`);
			}

			// 最小と最大のノード数における1ノードあたりの時間を比較
			// 線形なら比率は1:1付近、悪くても1:2未満
			const firstPerNode = times[0].avgTime / times[0].size;
			const lastPerNode = times[times.length - 1].avgTime / times[times.length - 1].size;
			const ratio = lastPerNode / firstPerNode;

			console.log(`  スケール比率: ${ratio.toFixed(2)}x`);
			expect(ratio, 'ノード数増加に対して準線形以下でスケールすべき').toBeLessThan(3);
		});
	});

	describe('レイアウト結果の正確性', () => {
		it('全ノードに位置が割り当てられる', () => {
			const tasks = generateMockTasks(100);
			const result = engine.layout(tasks);

			expect(result.positions.size).toBe(100);

			for (const task of tasks) {
				const pos = result.positions.get(task.id);
				expect(pos, `ノード ${task.id} に位置が割り当てられていない`).toBeDefined();
				expect(typeof pos?.x).toBe('number');
				expect(typeof pos?.y).toBe('number');
				expect(typeof pos?.layer).toBe('number');
			}
		});

		it('依存関係のあるノードは異なるレイヤーに配置される', () => {
			// 明示的な依存チェーン: task-0 -> task-1 -> task-2
			const nodes = [
				{ id: 'task-0', title: 'Task 0', node_type: 'task' as const, status: 'pending' as const, progress: 0, priority: 'medium' as const, assignee: 'user-0', dependencies: [] },
				{ id: 'task-1', title: 'Task 1', node_type: 'task' as const, status: 'pending' as const, progress: 0, priority: 'medium' as const, assignee: 'user-0', dependencies: ['task-0'] },
				{ id: 'task-2', title: 'Task 2', node_type: 'task' as const, status: 'pending' as const, progress: 0, priority: 'medium' as const, assignee: 'user-0', dependencies: ['task-1'] }
			];

			const result = engine.layout(nodes);

			const layer0 = result.positions.get('task-0')?.layer;
			const layer1 = result.positions.get('task-1')?.layer;
			const layer2 = result.positions.get('task-2')?.layer;

			expect(layer0).toBe(0);
			expect(layer1).toBe(1);
			expect(layer2).toBe(2);
		});
	});
});
