<script lang="ts">
        import { ISPMap } from '$lib/types/enum';
        import type { WtsISP } from '$lib/types/enum';
        let { i }: { i: WtsISP } = $props();
</script>

<span>
        {ISPMap[i]}
</span>
