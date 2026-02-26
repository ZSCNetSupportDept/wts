<script lang="ts">
	import type { Ticket } from '$lib/types/apiResponse';
	import { ModalFooter, Button } from 'carbon-components-svelte';
	let {
		src,
		t,
		loading = false,
		view = $bindable<'trace' | 'cancel' | 'update'>(),
		open = $bindable<boolean>()
	}: {
		src: 'user' | 'operator' | null;
		t: Ticket | null;
		loading: boolean;
		view: 'trace' | 'cancel' | 'update';
		open: boolean;
	} = $props();

	function isTicketActive(t: Ticket | null): boolean {
		if (!t) return false;
		return t.status !== 'canceled' && t.status !== 'solved';
	}
</script>

{#if loading}
	<div></div>
{:else if src === 'user' && isTicketActive(t)}
	<!-- 用户工单卡片的下方按钮 -->
	<ModalFooter>
		<Button kind="danger" on:click={() => (view = 'cancel')}>取消工单</Button>
		<Button kind="secondary" on:click={() => ((open = false), (view = 'trace'))}>返回</Button>
	</ModalFooter>
{:else if src === 'operator'}
	<!-- 后台工单记录视图下的下方按钮 -->
	<ModalFooter>
		<Button kind="secondary" on:click={() => ((open = false), (view = 'trace'))}>返回   </Button>
		<Button kind="primary" on:click={() => (view = 'update')}>更新状态</Button>
	</ModalFooter>
{/if}
