<script lang="ts">
        import { PriorityMap } from '$lib/types/enum';
        import type { WtsPriority } from '$lib/types/enum';
        let { p }: { p: WtsPriority } = $props();
</script>

{#if p === 'highest' || p==='assigned'}
<span class="text-red-600" style="font-size: 15px;">
        <strong>{PriorityMap[p]}</strong>
</span>
{/if}

{#if p!== 'highest' && p!=='assigned'}
<span class="text-gray-600" style="font-size: 15px;">
        <strong>{PriorityMap[p]}</strong>
</span>
{/if}