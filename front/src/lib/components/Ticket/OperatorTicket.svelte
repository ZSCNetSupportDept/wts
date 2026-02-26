<script lang="ts">
	import RetroCard from '../RetroCard.svelte';
	import type { Ticket } from '$lib/types/apiResponse';
	import { FormatDate, FormatTime } from '$lib/types/RFC3339';
	import WtsStatus from './WtsStatus.svelte';
	import WtsPriority from './WtsPriority.svelte';
	import BlockRoom from './BlockRoom.svelte';
	import WtsISP from './WtsISP.svelte';
	import WtsCategory from './WtsCategory.svelte';
	import { TicketModal } from '$lib/states/ticketDetails.svelte';

	let { t }: { t: Ticket } = $props();
</script>

<RetroCard style="padding: 10px;">
	<div
		role="button"
		tabindex="0"
		onclick={() => TicketModal.open(t, 'operator')}
		onkeydown={(e) => e.key === 'Enter' && TicketModal.open(t, 'operator')}
		style="cursor: pointer; outline: none;"
	>
		<div class="flex items-center justify-between">
			<p class="font-bold" style="font-size: 19px;">📃No.{t.tid}</p>
			<WtsPriority p={t.priority} />
		</div>
		{#if t.appointed_at}
			<p style="color: #0f62fe; font-size: 12.5px;">
				<strong>该报修已预约在{FormatDate(t.appointed_at)}</strong>
			</p>
		{/if}
		<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
			<strong style="flex-shrink: 0;width: 7em;">状态</strong>
			<div style="font-size: 15px;"><WtsStatus s={t.status} ap={t.appointed_at}/></div>
		</div>
		<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
			<strong style="flex-shrink: 0;width: 7em;">报修时间</strong>
			<p style="font-size: 15px;">{FormatTime(t.submitted_at)}</p>
		</div>
		<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
			<strong style="flex-shrink: 0;width: 7em;">联系方式</strong>
			<p style="font-size: 15px;">{t.issuer.name} {t.issuer.phone}</p>
		</div>
		<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
			<strong style="flex-shrink: 0;width: 7em;">信息</strong>
			<p style="font-size: 15px;">
				<strong
					><BlockRoom b={t.issuer.block} r={t.issuer.room} />，<WtsISP i={t.issuer.isp} /></strong
				><br />账号：{t.issuer.account}
			</p>
		</div>
		{#if t.category != 'others'}
			<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
				<strong style="flex-shrink: 0;width: 7em;">故障类型</strong>
				<p style="font-size: 15px;"><strong><WtsCategory c={t.category} /></strong></p>
			</div>
		{/if}
		<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
			<strong style="flex-shrink: 0;width: 7em;">描述</strong>
			<p style="font-size: 15px;">
				{t.description}
				{#if t.occur_at}
					<br />发生时间：{FormatDate(t.occur_at)}
				{/if}
			</p>
		</div>
		{#if t.notes}
			<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
				<strong style="flex-shrink: 0;width: 7em;">备注</strong>
				<p style="font-size: 15px;">{t.notes}</p>
			</div>
		{/if}
	</div>
</RetroCard>
