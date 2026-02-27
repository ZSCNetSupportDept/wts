<script lang="ts">
	import { AccessMap } from '$lib/types/enum';
	import type { WtsAccess } from '$lib/types/enum';
	let { a }: { a: WtsAccess } = $props();
</script>

<span class="inline-block rounded-[19px] bg-gray-300 px-3 py-1 text-[15px]">
	{AccessMap[a]}
</span>
