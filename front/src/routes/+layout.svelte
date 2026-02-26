<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import logo from '$lib/assets/logo-256.svg';
	import { Theme } from 'carbon-components-svelte';

	let { children } = $props();

	// import 'carbon-components-svelte/css/g10.css'; //主样式

	import Dock from '$lib/components/Dock.svelte';

	import 'carbon-components-svelte/css/all.css';
	import type { CarbonTheme } from 'carbon-components-svelte/src/Theme/Theme.svelte';

	let theme: CarbonTheme = $state('g10');
</script>

<svelte:head><link rel="icon" href={logo} /></svelte:head>

<div class="app-container">
	<Theme bind:theme />

	<!--页面组件将在这里渲染 -->
	<main class="main-content">
		{@render children()}
	</main>

	<Dock />
</div>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
		box-sizing: border-box;
		/* 使用 Carbon 推荐的字体 */
		font-family: 'IBM Plex Sans', sans-serif;
	}

	.app-container {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
		margin: 0 auto;
		max-width: 500px;
		box-shadow: 0 0 25px rgba(0, 0, 0, 0);
		position: relative;
	}

	.main-content {
		flex-grow: 1;
		padding: 1rem;
		padding-bottom: 110px;
	}

	/* 全局修复 Carbon Select 组件的双箭头问题 */
	/* 不知道为什么会有这个奇怪的问题，大概是Tailwind CSS导致的？ */
	:global(.bx--select-input) {
		-webkit-appearance: none !important;
		-moz-appearance: none !important;
		appearance: none !important;
		background-image: none !important;
	}

	:global(.bx--select-input::-ms-expand) {
		display: none;
	}
</style>
