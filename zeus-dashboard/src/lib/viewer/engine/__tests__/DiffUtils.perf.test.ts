/**
 * DiffUtils パフォーマンステスト
 *
 * 非機能要件:
 * - タスクハッシュ計算（1000タスク）: 10ms 以内
 * - 変更検出: 5ms 以内
 */
import { describe, it, expect } from 'vitest';
import {
	createInitialState,
	computeTasksHash,
	computeDependencyHash,
	detectTaskChanges,
	identifyTaskDiff,
	computeVisibilityDiff
} from '../DiffUtils';
import {
	measurePerformance,
	measureOnce,
	assertPerformance,
	generateMockTasks,
	PERFORMANCE_THRESHOLDS,
	formatPerformanceResult,
	type MockGraphNode
} from './performance-helper';

describe('DiffUtils パフォーマンステスト', () => {
	describe('ハッシュ計算', () => {
		it('タスクハッシュ計算が閾値以内で完了する（1000タスク）', () => {
			const tasks = generateMockTasks(1000);
			const threshold = PERFORMANCE_THRESHOLDS.diffUpdate.computeHash;

			const result = measurePerformance(
				() => {
					computeTasksHash(tasks);
				},
				100,
				20
			);

			console.log(formatPerformanceResult('タスクハッシュ計算 (1000)', result));
			assertPerformance(result, threshold, 'タスクハッシュ計算');
		});

		it('依存関係ハッシュ計算が閾値以内で完了する（1000タスク）', () => {
			const tasks = generateMockTasks(1000);
			const threshold = PERFORMANCE_THRESHOLDS.diffUpdate.computeHash;

			const result = measurePerformance(
				() => {
					computeDependencyHash(tasks);
				},
				100,
				20
			);

			console.log(formatPerformanceResult('依存関係ハッシュ計算 (1000)', result));
			assertPerformance(result, threshold, '依存関係ハッシュ計算');
		});

		it('ハッシュが同一タスクで一貫している', () => {
			const tasks = generateMockTasks(100);

			const hash1 = computeTasksHash(tasks);
			const hash2 = computeTasksHash(tasks);

			expect(hash1).toBe(hash2);
		});

		it('ハッシュがタスク変更を検出する', () => {
			const tasks = generateMockTasks(100);
			// 初期状態を固定
			tasks[0].status = 'pending';
			tasks[0].progress = 0;

			const hash1 = computeTasksHash(tasks);

			// ステータス変更
			tasks[0].status = 'completed';
			tasks[0].progress = 100;
			const hash2 = computeTasksHash(tasks);

			expect(hash1).not.toBe(hash2);
		});
	});

	describe('変更検出', () => {
		it('変更検出が閾値以内で完了する（1000タスク）', () => {
			const tasks = generateMockTasks(1000);
			let state = createInitialState();

			// 初回実行
			const firstResult = detectTaskChanges(tasks, state);
			state = firstResult.newState;

			const threshold = PERFORMANCE_THRESHOLDS.diffUpdate.detectChanges;

			// 2回目以降（変更なし）
			const result = measurePerformance(
				() => {
					detectTaskChanges(tasks, state);
				},
				100,
				20
			);

			console.log(formatPerformanceResult('変更検出 (変更なし)', result));
			assertPerformance(result, threshold, '変更検出');
		});

		it('変更なしを正しく検出する', () => {
			const tasks = generateMockTasks(100);
			let state = createInitialState();

			// 初回
			const first = detectTaskChanges(tasks, state);
			expect(first.changeType).toBe('structure'); // 初回は常に structure
			state = first.newState;

			// 2回目（変更なし）
			const second = detectTaskChanges(tasks, state);
			expect(second.changeType).toBe('none');
		});

		it('データ変更を正しく検出する', () => {
			const tasks = generateMockTasks(100);
			let state = createInitialState();

			// 初期状態を明示的に設定
			tasks[0].status = 'pending';
			tasks[0].progress = 0;

			// 初回
			const first = detectTaskChanges(tasks, state);
			state = first.newState;

			// ステータスのみ変更
			tasks[0].status = 'completed';
			tasks[0].progress = 100;

			const second = detectTaskChanges(tasks, state);
			expect(second.changeType).toBe('data');
		});

		it('構造変更を正しく検出する - タスク追加', () => {
			const tasks = generateMockTasks(100);
			let state = createInitialState();

			const first = detectTaskChanges(tasks, state);
			state = first.newState;

			// タスク追加
			tasks.push({
				id: 'new-task',
				title: 'New Task',
				node_type: 'activity',
				status: 'pending',
				progress: 0,
				priority: 'medium',
				assignee: 'user-0',
				dependencies: []
			});

			const second = detectTaskChanges(tasks, state);
			expect(second.changeType).toBe('structure');
		});

		it('構造変更を正しく検出する - タスク削除', () => {
			const tasks = generateMockTasks(100);
			let state = createInitialState();

			const first = detectTaskChanges(tasks, state);
			state = first.newState;

			// タスク削除
			tasks.pop();

			const second = detectTaskChanges(tasks, state);
			expect(second.changeType).toBe('structure');
		});

		it('構造変更を正しく検出する - 依存関係変更', () => {
			const tasks = generateMockTasks(100);
			let state = createInitialState();

			const first = detectTaskChanges(tasks, state);
			state = first.newState;

			// 依存関係変更
			tasks[50].dependencies = ['task-0', 'task-1'];

			const second = detectTaskChanges(tasks, state);
			expect(second.changeType).toBe('structure');
		});
	});

	describe('タスク差分特定', () => {
		it('追加・削除・変更なしを正しく分類する', () => {
			const previousTasks = generateMockTasks(5);
			const previousIds = new Set(previousTasks.map((t) => t.id));

			// task-0 削除、task-5 追加
			const newTasks: MockGraphNode[] = [
				previousTasks[1],
				previousTasks[2],
				previousTasks[3],
				previousTasks[4],
				{
					id: 'task-5',
					title: 'Task 5',
					node_type: 'activity',
					status: 'pending',
					progress: 0,
					priority: 'medium',
					assignee: 'user-0',
					dependencies: []
				}
			];

			const diff = identifyTaskDiff(newTasks, previousIds);

			expect(diff.added.length).toBe(1);
			expect(diff.added[0].id).toBe('task-5');
			expect(diff.removed.length).toBe(1);
			expect(diff.removed[0]).toBe('task-0');
			expect(diff.unchanged.length).toBe(4);
		});

		it('差分特定が高速に動作する（1000タスク）', () => {
			const previousTasks = generateMockTasks(1000);
			const previousIds = new Set(previousTasks.map((t) => t.id));

			// 10タスク削除、10タスク追加
			const newTasks = [
				...previousTasks.slice(10),
				...generateMockTasks(10).map((t, i) => ({ ...t, id: `new-${i}` }))
			];

			const duration = measureOnce(() => {
				identifyTaskDiff(newTasks, previousIds);
			});

			console.log(`差分特定 (1000タスク): ${duration.toFixed(3)}ms`);
			expect(duration).toBeLessThanOrEqual(10);
		});
	});

	describe('可視性差分', () => {
		it('可視性変更を正しく検出する', () => {
			const previousVisible = new Set(['a', 'b', 'c', 'd']);
			const currentVisible = new Set(['b', 'c', 'e', 'f']);

			const diff = computeVisibilityDiff(currentVisible, previousVisible);

			expect(diff.becameVisible.sort()).toEqual(['e', 'f']);
			expect(diff.becameHidden.sort()).toEqual(['a', 'd']);
		});

		it('可視性差分計算が高速（1000ノード）', () => {
			const previousVisible = new Set<string>();
			const currentVisible = new Set<string>();

			// 500ノードが共通、250ノードが消え、250ノードが新出
			for (let i = 0; i < 750; i++) {
				previousVisible.add(`node-${i}`);
			}
			for (let i = 250; i < 1000; i++) {
				currentVisible.add(`node-${i}`);
			}

			const result = measurePerformance(
				() => {
					computeVisibilityDiff(currentVisible, previousVisible);
				},
				100,
				20
			);

			console.log(formatPerformanceResult('可視性差分計算 (1000)', result));
			expect(result.avgPerIteration).toBeLessThanOrEqual(5);
		});
	});

	describe('スケーラビリティ', () => {
		it('タスク数に対して線形にスケールする', () => {
			const sizes = [100, 500, 1000, 2000, 5000];
			const times: { size: number; avgTime: number }[] = [];

			for (const size of sizes) {
				const tasks = generateMockTasks(size);

				const result = measurePerformance(
					() => {
						computeTasksHash(tasks);
					},
					20,
					5
				);

				times.push({ size, avgTime: result.avgPerIteration });
			}

			console.log('\nハッシュ計算スケーラビリティ:');
			for (const { size, avgTime } of times) {
				const perTask = (avgTime / size) * 1000000; // ns/task
				console.log(`  ${size}タスク: ${avgTime.toFixed(2)}ms (${perTask.toFixed(1)}ns/タスク)`);
			}

			// 線形スケールを確認
			const firstPerTask = times[0].avgTime / times[0].size;
			const lastPerTask = times[times.length - 1].avgTime / times[times.length - 1].size;
			const ratio = lastPerTask / firstPerTask;

			console.log(`  スケール比率: ${ratio.toFixed(2)}x`);
			expect(ratio, 'タスク数増加に対して線形スケール').toBeLessThan(2);
		});
	});
});
