import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
	plugins: [svelte({ hot: false })],
	test: {
		// テスト環境
		environment: 'node',
		// インクルードパターン
		include: ['src/**/*.{test,spec}.{js,ts}'],
		// 除外パターン
		exclude: ['node_modules', 'build', '.svelte-kit'],
		// グローバル設定
		globals: true,
		// タイムアウト（パフォーマンステスト用に長めに設定）
		testTimeout: 30000,
		// カバレッジ設定
		coverage: {
			provider: 'v8',
			reporter: ['text', 'json', 'html'],
			include: ['src/lib/viewer/engine/**/*.ts', 'src/lib/viewer/rendering/**/*.ts']
		},
		// エイリアス設定
		alias: {
			$lib: '/src/lib'
		}
	}
});
