<script lang="ts">
	import { StatusMap } from '$lib/types/enum';
	import type { WtsStatus } from '$lib/types/enum';
	import { DateRFC3339, type RFC3339 } from '$lib/types/RFC3339';
	import { isSameDay } from 'date-fns';
	let { s, ap }: { s: WtsStatus; ap: RFC3339 } = $props();

	const colorMap: Record<WtsStatus, string> = {
		fresh: 'text-red-600',
		scheduled: 'text-blue-600',
		delay: 'text-blue-500',
		escalated: 'text-green-600',
		solved: 'text-green-600',
		canceled: 'text-gray-500'
	};
</script>

{#if s === 'scheduled' && isSameDay(ap, new Date())}
	<span>
		<strong class="text-blue-600">已预约</strong>
		<strong class="text-red-600">(今天)</strong>
	</span>
{:else}
	<span class={colorMap[s]}>
		<strong>{StatusMap[s]}</strong>
	</span>
{/if}

<style>
</style>
