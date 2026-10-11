<script lang="ts">
	import type { Ticket } from '$lib/types/apiResponse';
	import { ModalFooter, Button } from 'carbon-components-svelte';

	let {
		view = $bindable<'trace' | 'cancel' | 'update' | 'appoint'>(),
		isUpReady = $bindable<boolean>()
	}: {
		view: 'trace' | 'cancel' | 'update' | 'appoint';
		isUpReady: boolean;
	} = $props();
</script>

<ModalFooter>
	<Button kind="secondary" on:click={() => (view = 'trace')}>返回上一级</Button>
	<Button kind="primary" on:click={() => (isUpReady = true)}>确认</Button>
</ModalFooter>
