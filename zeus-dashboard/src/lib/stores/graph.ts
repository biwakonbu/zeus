import { writable, derived } from 'svelte/store';
import type { GraphResponse, GraphStats } from '$lib/types/api';
import { fetchGraph } from '$lib/api/client';

// グラフデータのストア
export const graphData = writable<GraphResponse | null>(null);

// グラフの読み込み状態
export const graphLoading = writable<boolean>(false);

// グラフのエラー
export const graphError = writable<string | null>(null);

// Mermaid コードの派生ストア
export const mermaidCode = derived(graphData, ($data): string => {
	return $data?.mermaid ?? '';
});

// グラフ統計の派生ストア
export const graphStats = derived(graphData, ($data): GraphStats | null => {
	return $data?.stats ?? null;
});

// 循環検出の派生ストア
export const cycles = derived(graphData, ($data): string[][] => {
	return $data?.cycles ?? [];
});

// 孤立ノードの派生ストア
export const isolated = derived(graphData, ($data): string[] => {
	return $data?.isolated ?? [];
});

// 循環あり判定
export const hasCycles = derived(cycles, ($cycles): boolean => {
	return $cycles.length > 0;
});

// 孤立ノードあり判定
export const hasIsolated = derived(isolated, ($isolated): boolean => {
	return $isolated.length > 0;
});

// グラフを更新
export async function refreshGraph(): Promise<void> {
	graphLoading.set(true);
	graphError.set(null);

	try {
		const data = await fetchGraph();
		graphData.set(data);
	} catch (err) {
		const message = err instanceof Error ? err.message : 'Unknown error';
		graphError.set(message);
	} finally {
		graphLoading.set(false);
	}
}

// グラフを直接設定（SSE 用）
export function setGraph(data: GraphResponse): void {
	graphData.set(data);
	graphError.set(null);
}
