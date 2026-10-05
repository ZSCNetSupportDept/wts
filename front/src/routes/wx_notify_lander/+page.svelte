<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Button, NotificationQueue } from 'carbon-components-svelte';
	import RetroCard from '$lib/components/RetroCard.svelte';
	import WxOpenSubscribe from '$lib/components/WxOpenSubscribe.svelte';
	import { GetSubscribeConfig } from '$lib/api';
	import { StatusMap, type WtsStatus } from '$lib/types/enum';

	const MAX_MESSAGE_LENGTH = 200;

	let q: NotificationQueue;
	let tid = $state<number | undefined>(undefined);
	let status = $state<WtsStatus | undefined>(undefined);
	let message = $state('');
	let subscribeTemplateId = $state('');
	let subscribeUnavailable = $state(false);
	let subscribeConfigLoading = $state(true);
	let parsedLink = $state(false);
	let validLink = $derived(parsedLink && tid !== undefined && status !== undefined);

	function parseTid(value: string | null): number | undefined {
		if (!value || !/^\d+$/.test(value)) return undefined;
		const parsed = Number(value);
		return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : undefined;
	}

	function parseStatus(value: string | null): WtsStatus | undefined {
		if (!value || !Object.hasOwn(StatusMap, value)) return undefined;
		return value as WtsStatus;
	}

	function openTicket() {
		if (tid === undefined) return;
		void goto(`/repair/?open=${tid}`);
	}

	function onOpenSubscribeUnavailable() {
		subscribeUnavailable = true;
	}

	async function fetchSubscribeConfig() {
		if (!navigator.userAgent.toLowerCase().includes('micromessenger')) {
			subscribeUnavailable = true;
			subscribeConfigLoading = false;
			return;
		}

		try {
			const cfg = await GetSubscribeConfig();
			if (cfg.success && cfg.template_id) {
				subscribeTemplateId = cfg.template_id;
			} else {
				subscribeUnavailable = true;
			}
		} catch (error: any) {
			subscribeUnavailable = true;
			q.add({
				kind: 'warning',
				title: '微信提醒暂不可用',
				subtitle: error.response?.data?.msg || error.message || '未知错误',
				timeout: 3000
			});
		} finally {
			subscribeConfigLoading = false;
		}
	}

	onMount(() => {
		tid = parseTid(page.url.searchParams.get('tid'));
		status = parseStatus(page.url.searchParams.get('status'));
		message = (page.url.searchParams.get('message') || '').slice(0, MAX_MESSAGE_LENGTH);
		parsedLink = true;
		if (!tid || !status) {
			q.add({
				kind: 'error',
				title: '通知链接无效',
				subtitle: '缺少或包含无效的工单信息',
				timeout: 5000
			});
			return;
		}
		void fetchSubscribeConfig();
	});
</script>

<h1>工单状态通知</h1>
<br />
<hr />
<br />

{#if !parsedLink}
	<p>正在读取通知...</p>
{:else if validLink}
	<RetroCard style="padding: 10px;">
		<div class="ticket-summary">
			<p class="ticket-id">📃No.{tid}</p>
			<div class="summary-row">
				<strong>状态</strong>
				<span>{StatusMap[status]}</span>
			</div>
			<div class="summary-row">
				<strong>更新说明</strong>
				<span>{message || '暂无补充说明'}</span>
			</div>
		</div>
	</RetroCard>

	{#if subscribeTemplateId && !subscribeUnavailable}
		<WxOpenSubscribe
			templateId={subscribeTemplateId}
			label="查看工单详情"
			onSuccess={openTicket}
			onError={openTicket}
			onUnavailable={onOpenSubscribeUnavailable}
		/>
	{:else if subscribeConfigLoading}
		<Button disabled>查看工单详情</Button>
	{:else}
		<Button on:click={openTicket}>查看工单详情</Button>
	{/if}
{:else}
	<p>这条通知链接无效，无法定位对应工单。</p>
	<br />
	<Button href="/repair/">查看我的工单</Button>
{/if}

<NotificationQueue bind:this={q} />

<style>
	.ticket-summary {
		cursor: default;
	}

	.ticket-id {
		font-size: 19px;
		font-weight: bold;
	}

	.summary-row {
		display: flex;
		gap: 1rem;
		margin-top: 12px;
		line-height: 1.6;
	}

	.summary-row strong {
		flex: 0 0 7em;
	}

</style>