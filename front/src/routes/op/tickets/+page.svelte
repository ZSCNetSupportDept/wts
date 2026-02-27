<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	import { CheckAndGetJWT, Guard } from '$lib/jwt';
	import { IsAdmin, IsOperator, PriorityMap, CategoryMap } from '$lib/types/enum';
	import { onMount } from 'svelte';
	import { Button, NotificationQueue } from 'carbon-components-svelte';
	import Return from 'carbon-icons-svelte/lib/Return.svelte';
	import type { FilterTicketsReq } from '$lib/types/apiRequest';
	import type { Ticket } from '$lib/types/apiResponse';
	import { FilterTickets } from '$lib/api';
	import { criteria } from '$lib/states/ticketCriteriaSearch.svelte';
	import OperatorTicket from '$lib/components/Ticket/OperatorTicket.svelte';
	import TicketDetail from '$lib/components/TraceDetail/TicketDetail.svelte';
	import { TicketModal } from '$lib/states/ticketDetails.svelte';
	import type { RFC3339 } from '$lib/types/RFC3339';

	let q: NotificationQueue;
	let tickets = $state([] as Ticket[]);
	let ticketEmpty = $state(false);
	let ok = $state(false);

	onMount(() => (Guard(IsOperator), fetchTickets1()));

	function toMs(rfc3339: RFC3339 | string | undefined) {
		const t = rfc3339 ? Date.parse(rfc3339) : NaN;
		return Number.isFinite(t) ? t : 0;
	}

	/** 从房间号字符串推导楼层：109 -> 1, 1033 -> 10 */
	function getFloorFromRoom(room: string | undefined | null): number | null {
		//console.log('getFloorFromRoom', { room });
		if (!room) return null;

		const digits = String(room).match(/\d+/g)?.join('') ?? '';
		if (digits.length < 3 || digits.length > 4) return null;

		const roomNum = Number.parseInt(digits, 10);
		if (!Number.isFinite(roomNum)) return null;

		const floor = Math.floor(roomNum / 100);
		//console.log('getFloorFromRoom', { room, digits, roomNum, floor });
		return Number.isFinite(floor) ? floor : null;
	}

	function postProcess() {
		if (ticketEmpty) {
			return;
		}
		if (criteria._order === 'newest') {
			tickets = [...tickets].sort((a, b) => toMs(b.submitted_at) - toMs(a.submitted_at));
		}
		if (criteria._order === 'oldest') {
			tickets = [...tickets].sort((a, b) => toMs(a.submitted_at) - toMs(b.submitted_at));
		}
		if (criteria._floor !== null && criteria._floor !== undefined && criteria._floor !== 0) {
			tickets = tickets.filter((t) => getFloorFromRoom(t?.issuer?.room) === criteria._floor);
		}
		if (criteria._view_today_scheduled) {
			const todayStart = new Date().setHours(0, 0, 0, 0);
			const todayEnd = new Date().setHours(23, 59, 59, 999);
			tickets = tickets.filter(
				(t) =>
					t.status !== 'scheduled' ||
					(toMs(t.appointed_at) >= todayStart && toMs(t.appointed_at) <= todayEnd)
			);
		}

		ok = true;
		ticketEmpty = tickets.length === 0;
		return;
	}

	async function fetchTickets1() {
		ok = false;
		try {
			let req: FilterTicketsReq = {} as FilterTicketsReq;
			Object.assign(req, criteria.r);
			let res = await FilterTickets(req);
			if (!res.success) {
				throw new Error(res.msg || '获取工单列表失败');
			}
			tickets = res.tickets;
			if (!tickets) {
				ticketEmpty = true;
			}
			postProcess();
		} catch (e: any) {
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '获取工单列表失败',
				subtitle: errMsg,
				timeout: 3000
			});
			return;
		}
	}

	async function refreshTickets() {
		await fetchTickets1();
	}
</script>

<h1>报修单检索结果</h1>
<br />
<hr />
<br />
<p>按照您提供的条件，获得的检索结果。</p>
<br />
<div
	style="display: flex; justify-content: flex-end; transform: translate(-17px,0px); margin-bottom: 15px;"
>
	<Button href="/op/ticket_search">修改条件<Return /></Button>
</div>
{#if !ok}
	<p>处理中，请稍等...</p>
{/if}
{#if ticketEmpty === false}
	{#each tickets as t}
		<OperatorTicket {t} />
	{/each}
{:else}
	<span>没有找到符合条件的报修单。</span>
{/if}
<TicketDetail
	t={TicketModal.NowTicket}
	bind:open={TicketModal.Opened}
	src={TicketModal.SRC}
	onTicketChanged={refreshTickets}
/>

<NotificationQueue bind:this={q} />
