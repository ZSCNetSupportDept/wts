<script lang="ts">
        import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { Button, NotificationQueue } from 'carbon-components-svelte';
	import { Return } from 'carbon-icons-svelte';
	import { GetTicket } from '$lib/api';
	import type { Ticket, GetTicketRes } from '$lib/types/apiResponse';
	import OperatorTicket from '$lib/components/Ticket/OperatorTicket.svelte';
	import TicketDetail from '$lib/components/TraceDetail/TicketDetail.svelte';
	import { TicketModal } from '$lib/states/ticketDetails.svelte';
	import { Guard } from '$lib/jwt';
	import { IsAdmin } from '$lib/types/enum';
	let tickets = $state([] as Ticket[]);
	let name = page.url.searchParams.get('name')
	let ok = $state(false);

	onMount(getTicket);

	async function getTicket() {
		ok = false;
		let wx = page.url.searchParams.get('wx');
		if (!wx) {
			q.add({
				title: '错误',
				subtitle: '缺少微信号参数',
				kind: 'error',
				timeout: 5000
			});
			return;
		}
		ok = false;
		try {
			let res: GetTicketRes = await GetTicket(wx);
			if (!res.success) {
				throw new Error(res.msg || '获取用户失败');
			}
			tickets = res.tickets || [];
			ok = true;
		} catch (e: any) {
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '获取用户失败',
				subtitle: errMsg,
				timeout: 3000
			});
			return;
		}
	}

	let q: NotificationQueue;

	onMount(() => Guard(IsAdmin));
</script>

<h1>用户报修</h1>
<br />
<hr />
<br />
<p>用户：{name}的所有报修</p>
<br />
<div
	style="display: flex; justify-content: flex-end; transform: translate(-17px,0px); margin-bottom: 15px;"
>
	<Button href="/admin/filter_users">返回<Return /></Button>
</div>

{#if !ok}
	<p>处理中，请稍等...</p>
{/if}
{#if tickets.length > 0}
	{#each tickets as t}
		<OperatorTicket {t} />
	{/each}
{:else}
	<span>该用户没有提交过报修工单</span>
{/if}

<TicketDetail t={TicketModal.NowTicket} bind:open={TicketModal.Opened} src={TicketModal.SRC} />

<NotificationQueue bind:this={q} />
