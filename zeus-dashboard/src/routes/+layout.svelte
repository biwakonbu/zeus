<script lang="ts">
	import '../lib/theme/factorio.css';
	import favicon from '$lib/assets/favicon.svg';
	import Header from '$lib/components/layout/Header.svelte';
	import Footer from '$lib/components/layout/Footer.svelte';
	import { ToastContainer, KeyboardHelp } from '$lib/components/ui';
	import { connectionState } from '$lib/stores/connection';
	import { keyboardStore } from '$lib/stores/keyboard';
	import { onMount } from 'svelte';

	let { children } = $props();

	// キーボードヘルプ表示状態
	let showKeyboardHelp = $state(false);

	// グローバルキーボードショートカット登録
	onMount(() => {
		// ? キーでヘルプ表示
		const unregisterHelp = keyboardStore.register({
			key: '?',
			modifiers: ['shift'],
			description: 'キーボードショートカットを表示',
			category: 'ヘルプ',
			action: () => {
				showKeyboardHelp = true;
			}
		});

		// / キーで検索フォーカス（将来対応）
		const unregisterSearch = keyboardStore.register({
			key: '/',
			description: '検索にフォーカス',
			category: 'ナビゲーション',
			action: () => {
				const searchInput = document.querySelector<HTMLInputElement>('[data-search-input]');
				if (searchInput) {
					searchInput.focus();
				}
			}
		});

		return () => {
			unregisterHelp();
			unregisterSearch();
		};
	});

	// グローバルキーイベント処理
	function handleKeydown(event: KeyboardEvent) {
		// 入力要素にフォーカスがある場合はスキップ
		const target = event.target as HTMLElement;
		if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
			return;
		}

		keyboardStore.handleKeydown(event);
	}

	function closeKeyboardHelp() {
		showKeyboardHelp = false;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Zeus Dashboard</title>
</svelte:head>

<div class="app-container industrial-bg">
	<Header connectionState={$connectionState} />

	<main class="main-content">
		{@render children()}
	</main>

	<Footer />
</div>

<!-- グローバル Toast コンテナ -->
<ToastContainer />

<!-- キーボードショートカットヘルプ -->
{#if showKeyboardHelp}
	<KeyboardHelp onClose={closeKeyboardHelp} />
{/if}

<style>
	.app-container {
		display: flex;
		flex-direction: column;
		height: 100vh;
		overflow: hidden;
	}

	.main-content {
		flex: 1;
		padding: 0;
		width: 100%;
		min-height: 0;
		overflow: hidden;
	}
</style>
