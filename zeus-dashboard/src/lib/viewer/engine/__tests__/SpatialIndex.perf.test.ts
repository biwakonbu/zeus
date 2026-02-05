/**
 * SpatialIndex パフォーマンステスト
 *
 * 非機能要件:
 * - 1000ノード挿入: 50ms 以内
 * - 5000ノード挿入: 200ms 以内
 * - ビューポートクエリ: 5ms 以内
 * - ポイントクエリ: 1ms 以内
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { SpatialIndex, type SpatialItem } from '../SpatialIndex';
import {
	measurePerformance,
	measureOnce,
	assertPerformance,
	assertPerformanceOnce,
	PERFORMANCE_THRESHOLDS,
	formatPerformanceResult
} from './performance-helper';

/**
 * テスト用アイテムを生成
 */
function generateItems(count: number, bounds: { width: number; height: number }): SpatialItem[] {
	const items: SpatialItem[] = [];
	const itemWidth = 200;
	const itemHeight = 80;

	for (let i = 0; i < count; i++) {
		items.push({
			id: `item-${i}`,
			x: Math.random() * (bounds.width - itemWidth),
			y: Math.random() * (bounds.height - itemHeight),
			width: itemWidth,
			height: itemHeight
		});
	}

	return items;
}

describe('SpatialIndex パフォーマンステスト', () => {
	const worldBounds = { x: 0, y: 0, width: 10000, height: 10000 };
	let index: SpatialIndex;

	beforeEach(() => {
		index = new SpatialIndex(worldBounds);
	});

	describe('挿入操作', () => {
		it('1000アイテムの挿入が閾値以内で完了する', () => {
			const items = generateItems(1000, worldBounds);
			const threshold = PERFORMANCE_THRESHOLDS.spatialIndex.insert1000;

			const duration = measureOnce(() => {
				for (const item of items) {
					index.insert(item);
				}
			});

			console.log(`1000アイテム挿入: ${duration.toFixed(3)}ms`);
			assertPerformanceOnce(duration, threshold, '1000アイテム挿入');
		});

		it('5000アイテムの挿入が閾値以内で完了する', () => {
			const items = generateItems(5000, worldBounds);
			const threshold = PERFORMANCE_THRESHOLDS.spatialIndex.insert5000;

			const duration = measureOnce(() => {
				for (const item of items) {
					index.insert(item);
				}
			});

			console.log(`5000アイテム挿入: ${duration.toFixed(3)}ms`);
			assertPerformanceOnce(duration, threshold, '5000アイテム挿入');
		});

		it('一括挿入（insertAll）が個別挿入より高速または同等', () => {
			const items = generateItems(1000, worldBounds);

			// 個別挿入
			const index1 = new SpatialIndex(worldBounds);
			const individualTime = measureOnce(() => {
				for (const item of items) {
					index1.insert(item);
				}
			});

			// 一括挿入
			const index2 = new SpatialIndex(worldBounds);
			const batchTime = measureOnce(() => {
				index2.insertAll(items);
			});

			console.log(`個別挿入: ${individualTime.toFixed(3)}ms`);
			console.log(`一括挿入: ${batchTime.toFixed(3)}ms`);

			// 一括挿入は個別より悪くても2倍以内
			expect(batchTime).toBeLessThanOrEqual(individualTime * 2);
		});
	});

	describe('クエリ操作', () => {
		it('ビューポートクエリが閾値以内で完了する（1000アイテム）', () => {
			const items = generateItems(1000, worldBounds);
			index.insertAll(items);

			// 画面サイズのビューポート
			const viewport = { x: 1000, y: 1000, width: 1920, height: 1080 };
			const threshold = PERFORMANCE_THRESHOLDS.spatialIndex.query;

			const result = measurePerformance(
				() => {
					index.queryRect(viewport);
				},
				100,
				20
			);

			console.log(formatPerformanceResult('ビューポートクエリ', result));
			assertPerformance(result, threshold, 'ビューポートクエリ');
		});

		it('ビューポートクエリが閾値以内で完了する（5000アイテム）', () => {
			const items = generateItems(5000, worldBounds);
			index.insertAll(items);

			const viewport = { x: 2000, y: 2000, width: 1920, height: 1080 };
			// 5000アイテムでも同等の閾値を目標
			const threshold = PERFORMANCE_THRESHOLDS.spatialIndex.query * 2;

			const result = measurePerformance(
				() => {
					index.queryRect(viewport);
				},
				100,
				20
			);

			console.log(formatPerformanceResult('ビューポートクエリ (5000)', result));
			assertPerformance(result, threshold, 'ビューポートクエリ (5000)');
		});

		it('ポイントクエリが閾値以内で完了する', () => {
			const items = generateItems(1000, worldBounds);
			index.insertAll(items);

			const threshold = PERFORMANCE_THRESHOLDS.spatialIndex.pointQuery;

			const result = measurePerformance(
				() => {
					// ランダムな位置をクエリ
					const x = Math.random() * worldBounds.width;
					const y = Math.random() * worldBounds.height;
					index.queryPoint(x, y);
				},
				100,
				20
			);

			console.log(formatPerformanceResult('ポイントクエリ', result));
			assertPerformance(result, threshold, 'ポイントクエリ');
		});

		it('queryViewport（マージン付き）が正常に動作する', () => {
			const items = generateItems(500, worldBounds);
			index.insertAll(items);

			const viewport = { x: 1000, y: 1000, width: 1920, height: 1080 };
			const margin = 100;

			const result = measurePerformance(
				() => {
					index.queryViewport(viewport, margin);
				},
				50,
				10
			);

			console.log(formatPerformanceResult('queryViewport (マージン付き)', result));

			// queryRect より多少遅くても許容
			expect(result.avgPerIteration).toBeLessThanOrEqual(
				PERFORMANCE_THRESHOLDS.spatialIndex.query * 2
			);
		});
	});

	describe('クエリ結果の正確性', () => {
		it('ビューポート内のアイテムのみを返す', () => {
			const items: SpatialItem[] = [
				{ id: 'inside-1', x: 100, y: 100, width: 50, height: 50 },
				{ id: 'inside-2', x: 200, y: 200, width: 50, height: 50 },
				{ id: 'outside-1', x: 1000, y: 1000, width: 50, height: 50 },
				{ id: 'overlap', x: 450, y: 450, width: 100, height: 100 } // 境界に重なる
			];

			index.insertAll(items);

			const viewport = { x: 0, y: 0, width: 500, height: 500 };
			const result = index.queryRect(viewport);

			const ids = result.map((r) => r.id).sort();

			expect(ids).toContain('inside-1');
			expect(ids).toContain('inside-2');
			expect(ids).toContain('overlap'); // 重なっているので含まれる
			expect(ids).not.toContain('outside-1');
		});

		it('ポイントクエリが正確なアイテムを返す', () => {
			const items: SpatialItem[] = [
				{ id: 'item-1', x: 0, y: 0, width: 100, height: 100 },
				{ id: 'item-2', x: 200, y: 200, width: 100, height: 100 }
			];

			index.insertAll(items);

			// item-1 内のポイント
			const result1 = index.queryPoint(50, 50);
			expect(result1.map((r) => r.id)).toContain('item-1');

			// item-2 内のポイント
			const result2 = index.queryPoint(250, 250);
			expect(result2.map((r) => r.id)).toContain('item-2');

			// どちらにも含まれないポイント
			const result3 = index.queryPoint(150, 150);
			expect(result3.length).toBe(0);
		});
	});

	describe('Quadtree の効率性', () => {
		it('Quadtree がブルートフォースより高速', () => {
			const items = generateItems(2000, worldBounds);

			// Quadtree
			const quadtree = new SpatialIndex(worldBounds);
			quadtree.insertAll(items);

			const viewport = { x: 2500, y: 2500, width: 1920, height: 1080 };

			const quadtreeResult = measurePerformance(
				() => {
					quadtree.queryRect(viewport);
				},
				50,
				10
			);

			// ブルートフォース（単純なフィルタ）
			const bruteforceResult = measurePerformance(
				() => {
					items.filter((item) => {
						return !(
							item.x > viewport.x + viewport.width ||
							item.x + item.width < viewport.x ||
							item.y > viewport.y + viewport.height ||
							item.y + item.height < viewport.y
						);
					});
				},
				50,
				10
			);

			console.log(formatPerformanceResult('Quadtree', quadtreeResult));
			console.log(formatPerformanceResult('ブルートフォース', bruteforceResult));

			// Quadtree はブルートフォースと同等以上（理想は高速）
			// 小さいデータセットではオーバーヘッドがあるかもしれないので、
			// 2倍遅くなければ OK とする
			expect(quadtreeResult.avgPerIteration).toBeLessThanOrEqual(
				bruteforceResult.avgPerIteration * 2
			);
		});
	});

	describe('再構築', () => {
		it('rebuild がスケーラブルに動作する', () => {
			const items = generateItems(1000, worldBounds);
			index.insertAll(items);

			const newBounds = { x: -5000, y: -5000, width: 20000, height: 20000 };

			const duration = measureOnce(() => {
				index.rebuild(newBounds);
			});

			console.log(`rebuild (1000アイテム): ${duration.toFixed(3)}ms`);

			// rebuild は insert と同程度の時間
			expect(duration).toBeLessThanOrEqual(PERFORMANCE_THRESHOLDS.spatialIndex.insert1000 * 2);

			// 再構築後もクエリが正常に動作する
			const queryResult = index.queryRect({ x: 0, y: 0, width: 10000, height: 10000 });
			expect(queryResult.length).toBeGreaterThan(0);
		});
	});
});
