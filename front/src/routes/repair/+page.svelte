<script lang="ts">
	import OperatorTicket from '$lib/components/Ticket/OperatorTicket.svelte';
	import TicketDetail from '$lib/components/TraceDetail/TicketDetail.svelte';
	import UserTicket from '$lib/components/Ticket/UserTicket.svelte';
	import { sample1, sample2 } from '$lib/testData/ticket';
	import { Button } from 'carbon-components-svelte';
	import { TicketModal } from '$lib/states/ticketDetails.svelte';
	import Contract from 'carbon-pictograms-svelte/lib/Contract.svelte';
	import { onMount } from 'svelte';
	import { IsUser } from '$lib/types/enum';
	import { CheckAndGetJWT, Guard } from '$lib/jwt';
	import type { Ticket } from '$lib/types/apiResponse';
	import { GetTicket } from '$lib/api';
	import { NotificationQueue } from 'carbon-components-svelte';

	let q: NotificationQueue;

	let tickets = $state([] as Ticket[]);

	onMount(() => (Guard(IsUser), fetchTickets()));

	async function fetchTickets() {
		try {
			let res = await GetTicket(CheckAndGetJWT('parsed').openid);
			if (!res.success) {
				throw new Error(res.msg || '获取报修记录失败');
			}
			tickets = res.tickets;
		} catch (e: any) {
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '获取报修记录失败',
				subtitle: errMsg,
				timeout: 3000
			});
			return;
		}
	}

	async function refreshTickets1() {
		await fetchTickets();
	}
</script>

<h1 style="display: flex; align-items: center;">
	<span><Contract /></span>
	<span style="margin-left: 8px;">报修记录</span>
</h1>
<br />
<hr />
<br />
<p>
	这里将显示您提交的所有报修记录，由于各种原因，我们可能只会显示您最近几个报修单。点击单子可展开详情。
</p>
<br />
<div
	style="display: flex; justify-content: flex-end; transform: translate(-17px,0px); margin-bottom: 15px;"
>
	<Button href="/repair/new">提交新报修</Button>
</div>

<!--<OperatorTicket t={sample1} />
<OperatorTicket t={sample2} />-->

<!--<hr />-->

{#each tickets as t}
	<UserTicket {t} />
{/each}

<TicketDetail
	t={TicketModal.NowTicket}
	bind:open={TicketModal.Opened}
	src={TicketModal.SRC}
	onTicketChanged={refreshTickets1}
/>

<NotificationQueue bind:this={q} />
