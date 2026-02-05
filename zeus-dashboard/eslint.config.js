import svelte from 'eslint-plugin-svelte';
import svelteParser from 'svelte-eslint-parser';
import tsParser from '@typescript-eslint/parser';

// Svelte 5 のルーン
const svelteGlobals = {
	$state: 'readonly',
	$derived: 'readonly',
	$effect: 'readonly',
	$props: 'readonly',
	$bindable: 'readonly'
};

// ブラウザ環境のグローバル変数
const browserGlobals = {
	window: 'readonly',
	document: 'readonly',
	console: 'readonly',
	setTimeout: 'readonly',
	clearTimeout: 'readonly',
	setInterval: 'readonly',
	clearInterval: 'readonly',
	requestAnimationFrame: 'readonly',
	cancelAnimationFrame: 'readonly',
	HTMLElement: 'readonly',
	HTMLCanvasElement: 'readonly',
	HTMLDivElement: 'readonly',
	HTMLInputElement: 'readonly',
	HTMLButtonElement: 'readonly',
	HTMLUListElement: 'readonly',
	Node: 'readonly',
	alert: 'readonly',
	MouseEvent: 'readonly',
	WheelEvent: 'readonly',
	KeyboardEvent: 'readonly',
	PointerEvent: 'readonly',
	Event: 'readonly',
	EventSource: 'readonly',
	MessageEvent: 'readonly',
	CustomEvent: 'readonly',
	ResizeObserver: 'readonly',
	Performance: 'readonly',
	Window: 'readonly',
	fetch: 'readonly',
	URL: 'readonly',
	URLSearchParams: 'readonly',
	Blob: 'readonly',
	performance: 'readonly',
	crypto: 'readonly',
	location: 'readonly',
	navigator: 'readonly'
};

export default [
	{
		ignores: ['build/', '.svelte-kit/', 'node_modules/', 'storybook-static/', 'static/', '*.config.js', '*.config.ts']
	},
	// Svelte ファイルのみ対象（.ts は oxlint が担当）
	{
		files: ['**/*.svelte'],
		languageOptions: {
			parser: svelteParser,
			parserOptions: {
				parser: tsParser,
				ecmaVersion: 2022,
				sourceType: 'module'
			},
			globals: {
				...browserGlobals,
				...svelteGlobals
			}
		},
		plugins: {
			svelte
		},
		rules: {
			...svelte.configs.recommended.rules,
			'no-unused-vars': 'off' // TypeScript の型は oxlint/tsc に任せる
		}
	}
];
