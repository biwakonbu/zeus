import eslint from '@eslint/js';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsparser from '@typescript-eslint/parser';
import svelte from 'eslint-plugin-svelte';
import svelteParser from 'svelte-eslint-parser';
import prettier from 'eslint-config-prettier';

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

// Svelte 5 のルーン
const svelteGlobals = {
	$state: 'readonly',
	$derived: 'readonly',
	$effect: 'readonly',
	$props: 'readonly',
	$bindable: 'readonly'
};

export default [
	eslint.configs.recommended,
	{
		ignores: ['build/', '.svelte-kit/', 'node_modules/', 'storybook-static/', '*.config.js', '*.config.ts']
	},
	{
		files: ['**/*.ts'],
		languageOptions: {
			parser: tsparser,
			parserOptions: {
				ecmaVersion: 2022,
				sourceType: 'module'
			},
			globals: browserGlobals
		},
		plugins: {
			'@typescript-eslint': tseslint
		},
		rules: {
			...tseslint.configs.recommended.rules,
			'@typescript-eslint/no-unused-vars': 'warn',
			'@typescript-eslint/no-explicit-any': 'warn',
			'no-unused-vars': 'off'
		}
	},
	{
		files: ['**/*.svelte'],
		languageOptions: {
			parser: svelteParser,
			parserOptions: {
				parser: tsparser,
				ecmaVersion: 2022,
				sourceType: 'module'
			},
			globals: {
				...browserGlobals,
				...svelteGlobals
			}
		},
		plugins: {
			svelte,
			'@typescript-eslint': tseslint
		},
		rules: {
			...svelte.configs.recommended.rules,
			'@typescript-eslint/no-unused-vars': 'warn',
			'@typescript-eslint/no-explicit-any': 'warn',
			'no-unused-vars': 'off'
		}
	},
	prettier
];
