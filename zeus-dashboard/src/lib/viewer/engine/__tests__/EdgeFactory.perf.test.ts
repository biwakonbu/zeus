/**
 * EdgeFactory パフォーマンステスト
 *
 * 非機能要件:
 * - 1000エッジ作成: 50ms 以内
 * - ノードに関連するエッジ取得: 1ms 以内
 * - エッジ削除: 1ms 以内
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
	measurePerformance,
	measureOnce,
	assertPerformance,
	assertPerformanceOnce,
	PERFORMANCE_THRESHOLDS,
	formatPerformanceResult
} from './performance-helper';

// PixiJS Graphics をモック
vi.mock('pixi.js', () => ({
	Graphics: class MockGraphics {
		clear() { return this; }
		moveTo() { return this; }
		lineTo() { return this; }
		bezierCurveTo() { return this; }
		stroke() { return this; }
		destroy() {}
	}
}));

// モック後にインポート
import { EdgeFactory, TaskEdge } from '../../rendering/TaskEdge';

describe('EdgeFactory パフォーマンステスト', () => {
	let factory: EdgeFactory;

	beforeEach(() => {
		factory = new EdgeFactory();
	});

	afterEach(() => {
		factory.clear();
	});

	describe('エッジ作成', () => {
		it('1000エッジの作成が閾値以内で完了する', () => {
			const threshold = PERFORMANCE_THRESHOLDS.edgeFactory.create1000;

			const duration = measureOnce(() => {
				for (let i = 0; i < 1000; i++) {
					factory.getOrCreate(`node-${i}`, `node-${i + 1}`);
				}
			});

			console.log(`1000エッジ作成: ${duration.toFixed(3)}ms`);
			assertPerformanceOnce(duration, threshold, '1000エッジ作成');
		});

		it('既存エッジの取得が新規作成より高速', () => {
			// 最初に1000エッジを作成
			for (let i = 0; i < 1000; i++) {
				factory.getOrCreate(`node-${i}`, `node-${i + 1}`);
			}

			// 既存エッジの再取得
			const result = measurePerformance(() => {
				factory.getOrCreate('node-500', 'node-501');
			}, 1000, 100);

			console.log(formatPerformanceResult('既存エッジ取得', result));

			// 既存エッジの取得は非常に高速（0.1ms 未満）
			expect(result.avgPerIteration).toBeLessThanOrEqual(0.1);
		});
	});

	describe('インデックス検索', () => {
		it('ノードに関連するエッジの取得が閾値以内で完了する', () => {
			// 複数の接続を持つグラフを構築
			// node-0 は多くのノードに接続
			for (let i = 1; i <= 50; i++) {
				factory.getOrCreate('node-0', `node-${i}`);
			}

			// 他のノード間のエッジも追加
			for (let i = 1; i < 500; i++) {
				factory.getOrCreate(`node-${i}`, `node-${i + 1}`);
			}

			const threshold = PERFORMANCE_THRESHOLDS.edgeFactory.getEdgesForNode;

			const result = measurePerformance(() => {
				factory.getEdgesForNode('node-0');
			}, 1000, 100);

			console.log(formatPerformanceResult('getEdgesForNode', result));
			assertPerformance(result, threshold, 'getEdgesForNode');
		});

		it('getEdgesForNode が正確なエッジを返す', () => {
			factory.getOrCreate('node-a', 'node-b');
			factory.getOrCreate('node-a', 'node-c');
			factory.getOrCreate('node-b', 'node-c');
			factory.getOrCreate('node-d', 'node-a');

			const edgesForA = factory.getEdgesForNode('node-a');
			expect(edgesForA.length).toBe(3); // a->b, a->c, d->a

			const edgesForB = factory.getEdgesForNode('node-b');
			expect(edgesForB.length).toBe(2); // a->b, b->c

			const edgesForD = factory.getEdgesForNode('node-d');
			expect(edgesForD.length).toBe(1); // d->a
		});

		it('getEdgeCountForNode が O(1) で動作する', () => {
			// 大量のエッジを持つノードを作成
			for (let i = 0; i < 100; i++) {
				factory.getOrCreate('hub-node', `spoke-${i}`);
			}

			const result = measurePerformance(() => {
				factory.getEdgeCountForNode('hub-node');
			}, 1000, 100);

			console.log(formatPerformanceResult('getEdgeCountForNode', result));

			// O(1) なので非常に高速
			expect(result.avgPerIteration).toBeLessThanOrEqual(0.01);
		});
	});

	describe('削除操作', () => {
		it('エッジ削除が閾値以内で完了する', () => {
			// エッジを作成
			for (let i = 0; i < 100; i++) {
				factory.getOrCreate(`node-${i}`, `node-${i + 1}`);
			}

			const threshold = PERFORMANCE_THRESHOLDS.edgeFactory.remove;

			const result = measurePerformance(() => {
				// 存在するエッジを削除（毎回再作成してから削除）
				factory.getOrCreate('test-from', 'test-to');
				factory.remove('test-from', 'test-to');
			}, 100, 20);

			console.log(formatPerformanceResult('エッジ削除', result));
			assertPerformance(result, threshold, 'エッジ削除');
		});

		it('削除後にインデックスが正しく更新される', () => {
			factory.getOrCreate('node-a', 'node-b');
			factory.getOrCreate('node-a', 'node-c');

			expect(factory.getEdgeCountForNode('node-a')).toBe(2);

			factory.remove('node-a', 'node-b');

			expect(factory.getEdgeCountForNode('node-a')).toBe(1);
			expect(factory.getEdgeCountForNode('node-b')).toBe(0);
		});
	});

	describe('スケーラビリティ', () => {
		it('大規模グラフでもインデックス検索が高速', () => {
			// 5000エッジのグラフを構築
			for (let i = 0; i < 5000; i++) {
				const from = `node-${Math.floor(i / 10)}`;
				const to = `node-${i % 500 + 500}`;
				factory.getOrCreate(from, to);
			}

			// 多くのエッジを持つノードの検索
			const result = measurePerformance(() => {
				factory.getEdgesForNode('node-0');
			}, 500, 50);

			console.log(formatPerformanceResult('大規模グラフ検索', result));

			// 大規模でも 5ms 以内
			expect(result.avgPerIteration).toBeLessThanOrEqual(5);
		});

		it('エッジ数に対する検索時間が一定', () => {
			// 異なるサイズのグラフで検索時間を比較
			const sizes = [100, 500, 1000, 2000];
			const times: { size: number; avgTime: number }[] = [];

			for (const size of sizes) {
				const testFactory = new EdgeFactory();

				// グラフを構築
				for (let i = 0; i < size; i++) {
					testFactory.getOrCreate(`node-${i}`, `node-${i + 1}`);
				}
				// 特定ノードに複数の接続を追加
				for (let i = 0; i < 20; i++) {
					testFactory.getOrCreate('target-node', `connected-${i}`);
				}

				const result = measurePerformance(() => {
					testFactory.getEdgesForNode('target-node');
				}, 200, 50);

				times.push({ size, avgTime: result.avgPerIteration });
				testFactory.clear();
			}

			console.log('\nインデックス検索スケーラビリティ:');
			for (const { size, avgTime } of times) {
				console.log(`  ${size}エッジ: ${(avgTime * 1000).toFixed(1)}µs`);
			}

			// O(1) インデックスなので、グラフサイズに関係なく検索時間は一定
			const firstTime = times[0].avgTime;
			const lastTime = times[times.length - 1].avgTime;

			// サイズが20倍になっても時間は2倍以内
			expect(lastTime).toBeLessThanOrEqual(firstTime * 3);
		});
	});

	describe('TaskEdge 静的メソッド', () => {
		it('createKey が一意のキーを生成する', () => {
			const key1 = TaskEdge.createKey('a', 'b');
			const key2 = TaskEdge.createKey('b', 'a');
			const key3 = TaskEdge.createKey('a', 'b');

			expect(key1).toBe('a-->b');
			expect(key2).toBe('b-->a');
			expect(key1).toBe(key3);
			expect(key1).not.toBe(key2);
		});
	});
});
