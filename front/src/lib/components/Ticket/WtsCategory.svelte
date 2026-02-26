<script lang="ts">
        import { CategoryMap } from '$lib/types/enum';
        import type { WtsCategory } from '$lib/types/enum';
        let { c }: { c: WtsCategory } = $props();
</script>

<span>
        {CategoryMap[c]}
</span>