import { writable, derived } from 'svelte/store';
import type {
	PredictResponse,
	CompletionPrediction,
	RiskPrediction,
	VelocityReport
} from '$lib/types/api';
import { fetchPredict } from '$lib/api/client';

// 予測データのストア
export const predictionData = writable<PredictResponse | null>(null);

// 予測の読み込み状態
export const predictionLoading = writable<boolean>(false);

// 予測のエラー
export const predictionError = writable<string | null>(null);

// 完了予測の派生ストア
export const completion = derived(predictionData, ($data): CompletionPrediction | null => {
	return $data?.completion ?? null;
});

// リスク予測の派生ストア
export const risk = derived(predictionData, ($data): RiskPrediction | null => {
	return $data?.risk ?? null;
});

// ベロシティレポートの派生ストア
export const velocity = derived(predictionData, ($data): VelocityReport | null => {
	return $data?.velocity ?? null;
});

// 十分なデータがあるかの派生ストア
export const hasSufficientData = derived(completion, ($completion): boolean => {
	return $completion?.has_sufficient_data ?? false;
});

// 予測を更新
export async function refreshPrediction(): Promise<void> {
	predictionLoading.set(true);
	predictionError.set(null);

	try {
		const data = await fetchPredict();
		predictionData.set(data);
	} catch (err) {
		const message = err instanceof Error ? err.message : 'Unknown error';
		predictionError.set(message);
	} finally {
		predictionLoading.set(false);
	}
}

// 予測を直接設定（SSE 用）
export function setPrediction(data: PredictResponse): void {
	predictionData.set(data);
	predictionError.set(null);
}
