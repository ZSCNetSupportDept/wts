<script lang="ts">
	import Report from 'carbon-icons-svelte/lib/Report.svelte';
	import Help from 'carbon-icons-svelte/lib/Help.svelte';
	import Settings from 'carbon-icons-svelte/lib/Settings.svelte';
	import Home from 'carbon-icons-svelte/lib/Home.svelte';
	import Return from 'carbon-icons-svelte/lib/Return.svelte';
	import SearchAdvanced from 'carbon-icons-svelte/lib/SearchAdvanced.svelte';
	import CarbonForIbmDotcom from 'carbon-icons-svelte/lib/CarbonForIbmDotcom.svelte';
	import EventSchedule from 'carbon-icons-svelte/lib/EventSchedule.svelte';
	import { page } from '$app/state';

	let isAdminView = $derived(page.url.pathname.startsWith('/admin'));
	let isOperatorView = $derived(page.url.pathname.startsWith('/op'));
	let isUserView = $derived(!isAdminView && !isOperatorView);
</script>

{#if isUserView}
	<!-- 底部 Dock 导航栏 -->
	<nav class="bottom-nav">
		<a href="/" class="nav-item" class:active={page.url.pathname === '/'}>
			<Home />
			<span>主页</span>
		</a>

		<a href="/repair" class="nav-item" class:active={page.url.pathname === '/repair'}>
			<Report />
			<span>我的报修</span>
		</a>
		<a href="/help" class="nav-item" class:active={page.url.pathname === '/help'}>
			<Help />
			<span>网络攻略</span>
		</a>
		<a href="/me" class="nav-item" class:active={page.url.pathname === '/me'}>
			<Settings />
			<span>我</span>
		</a>
	</nav>
{/if}

{#if isOperatorView}
	<!-- 底部 Dock 导航栏 -->
	<nav class="bottom-nav">
		<a href="/op" class="nav-item" class:active={page.url.pathname === '/op'}>
			<CarbonForIbmDotcom />
			<span>后台中心</span>
		</a>

		<a
			href="/op/ticket_search"
			class="nav-item"
			class:active={page.url.pathname === '/op/ticket_search'}
		>
			<SearchAdvanced />
			<span>检索工单</span>
		</a>
		<a href="/op/scheduler" class="nav-item" class:active={page.url.pathname === '/op/scheduler'}>
			<EventSchedule />
			<span>我的排班</span>
		</a>
		<a href="/" class="nav-item" class:active={page.url.pathname === '/'}>
			<Return />
			<span>返回首页</span>
		</a>
	</nav>
{/if}

{#if isAdminView}
	<!-- 底部 Dock 导航栏 -->
	<nav class="bottom-nav">
		<a href="/admin" class="nav-item" class:active={page.url.pathname === '/admin'}>
			<CarbonForIbmDotcom />
			<span>管理中心</span>
		</a>

		<a
			href="/admin/add_ticket"
			class="nav-item"
			class:active={page.url.pathname === '/admin/add_ticket'}
		>
			<SearchAdvanced />
			<span>增添工单</span>
		</a>
		<a
			href="/admin/scheduler"
			class="nav-item"
			class:active={page.url.pathname === '/admin/scheduler'}
		>
			<EventSchedule />
			<span>成员排班</span>
		</a>
		<a href="/" class="nav-item" class:active={page.url.pathname === '/'}>
			<Return />
			<span>返回首页</span>
		</a>
	</nav>
{/if}

<style>
	.bottom-nav {
		position: fixed;
		bottom: 20px;
		left: 50%;
		transform: translateX(-50%);

		width: 90%;
		max-width: 450px;
		display: flex;
		justify-content: space-around;
		align-items: center;
		padding: 12px 0;

		background-color: #262626; /* Carbon Gray 90 */
		border-radius: 30px;
		box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
		z-index: 100;
	}

	.nav-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;

		color: #f4f4f4; /* Carbon Gray 10 */
		text-decoration: none;
		font-size: 12px;

		padding: 4px 12px;
		border-radius: 15px;
		transition: background-color 0.2s ease-in-out;
	}

	.nav-item:hover,
	.nav-item:focus {
		background-color: #393939; /* Carbon Gray 80 */
		outline: none;
	}

	.nav-item.active {
		color: #0f62fe; /* IBM Blue 60 */
	}
</style>
