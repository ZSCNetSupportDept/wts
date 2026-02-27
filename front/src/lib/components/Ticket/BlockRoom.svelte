<script lang="ts">
	import { BlockMap } from '$lib/types/enum';
	import type { WtsBlock } from '$lib/types/enum';
	let { b, r }: { b: WtsBlock; r: string } = $props();
</script>

<span>
	{BlockMap[b]}-{r}
</span>
