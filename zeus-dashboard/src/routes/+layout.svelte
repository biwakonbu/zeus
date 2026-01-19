<script lang="ts">
	import '../lib/theme/factorio.css';
	import favicon from '$lib/assets/favicon.svg';
	import Header from '$lib/components/layout/Header.svelte';
	import Footer from '$lib/components/layout/Footer.svelte';
	import { connectionState } from '$lib/stores/connection';

	let { children } = $props();
</script>

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

<style>
	.app-container {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	.main-content {
		flex: 1;
		padding: 0;
		width: 100%;
		overflow: hidden;
	}
</style>
