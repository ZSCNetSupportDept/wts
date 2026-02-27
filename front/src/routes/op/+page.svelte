<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	import { onMount } from 'svelte';
	import { CheckAndGetJWT, Guard } from '$lib/jwt';
	import { IsOperator, ZoneMap, ZoneToBlock, type WtsBlock, type WtsZone } from '$lib/types/enum';
	import { TicketOverview } from '$lib/api';
	import { NotificationQueue, Tile } from 'carbon-components-svelte';
	import { Radio } from 'carbon-icons-svelte';
	import { criteria, type Criteria } from '$lib/states/ticketCriteriaSearch.svelte';
	import { PriorityMap, type WtsPriority } from '$lib/types/enum';
	import { CategoryMap, type WtsCategory } from '$lib/types/enum';
	import { ISPMap, type WtsISP } from '$lib/types/enum';
	import { goto } from '$app/navigation';

	let name: string = $state('网维成员');

	onMount(() => Guard(IsOperator));

	onMount(() => {
		name = CheckAndGetJWT('parsed').name;
	});
	onMount(() => getTicketOverview());

	let countByBlock: Record<WtsBlock, number> = $state(undefined);
	let countByZone: Record<WtsZone, number> = $state(undefined);

	const zoneDisplayOrder: WtsZone[] = ['FX', 'BM', 'DM', 'QT', 'XHAB', 'XHCD', 'ZH'];

	function zoneTone(count: number | undefined) {
		const value = count ?? 0;
		if (value === 0) return 'none';
		if (value <= 5) return 'green';
		if (value <= 15) return 'yellow';
		return 'red';
	}

	async function getTicketOverview() {
		try {
			let res = await TicketOverview();
			if (!res.success) {
				throw new Error(res.msg || '获取片区总览失败');
			}
			countByBlock = res.count_by_block;
			parseTicketCount();
		} catch (e: any) {
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '获取片区总览失败',
				subtitle: errMsg + '，请重试',
				timeout: 5000
			});
		}
	}

	let q: NotificationQueue;

	function parseTicketCount() {
		const zoneCounts: Record<WtsZone, number> = {} as Record<WtsZone, number>;

		(Object.keys(ZoneToBlock) as WtsZone[]).forEach((zone) => {
			const blocks = ZoneToBlock[zone];
			zoneCounts[zone] = blocks.reduce((acc, block) => {
				return acc + (countByBlock?.[block] ?? 0);
			}, 0);
		});

		countByZone = zoneCounts;
	}

	function search(zone: WtsZone): Criteria {
		return {
			r: {
				scope: 'active',
				issuer: undefined,
				block: ZoneToBlock[zone],
				status: ['fresh', 'scheduled', 'escalated', 'delay'],
				priority: Object.keys(PriorityMap) as WtsPriority[],
				category: Object.keys(CategoryMap) as WtsCategory[],
				isp: Object.keys(ISPMap) as WtsISP[],
				newer_than: undefined,
				older_than: undefined
			},
			_order: 'priority',
			_floor: null,
			_blocks_in_zone: [zone],
			_view_today_scheduled: true
		} as Criteria;
	}

	function jumpSearch(zone: WtsZone) {
		Object.assign(criteria, search(zone));
		goto('/op/tickets');
	}
</script>

<h1>报修操作后台</h1>
<br />
<hr />
<br />
<p>
	你好，{name}!今天修了多少单？
</p>
<br />
<br />
<br />
<br />
<br />
<h2>片区总览</h2>
<p>每个片区的报修单数量</p>
<br />
<div class="zone-tiles">
	{#each zoneDisplayOrder as zone}
		{#if typeof countByZone?.[zone] !== 'undefined'}
			<Tile
				class={`zone-tile zone-${zoneTone(countByZone?.[zone])}`}
				on:click={() => jumpSearch(zone)}
			>
				<span class="zone-name">{ZoneMap[zone]}</span>
				<span class="zone-count">{countByZone?.[zone] ?? 0}</span>
			</Tile>
		{:else}
			<Tile class="zone-tile zone-none" on:click={() => jumpSearch(zone)}>
				<span class="zone-name">{ZoneMap[zone]}</span>
				<span class="zone-count">0</span>
			</Tile>
		{/if}
	{/each}
</div>

<NotificationQueue bind:this={q} />

<style>
	.zone-tiles {
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	:global(.zone-tile) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		border: 1px solid #c6c6c6;
		padding: 0.5rem 0.75rem;
	}

	.zone-name {
		flex: 1;
	}

	.zone-count {
		text-align: right;
		margin-left: auto;
		min-width: 2ch;
		font-variant-numeric: tabular-nums;
		color: var(--cds-text-secondary, #525252);
		font-weight: 600;
	}

	:global(.zone-none) {
		background-color: #ffffff;
	}

	:global(.zone-green) {
		background-color: #d9fbdb; /* Green 10 */
	}

	:global(.zone-yellow) {
		background-color: #fcf4d6; /* Yellow 10 */
	}

	:global(.zone-red) {
		background-color: #fff1f1; /* Red 10 */
	}
</style>
