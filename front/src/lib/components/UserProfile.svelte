<script lang="ts">
	import { goto } from '$app/navigation';
	import RetroCard from './RetroCard.svelte';
	import type { UserProfile } from '$lib/types/apiResponse';
	import BlockRoom from './Ticket/BlockRoom.svelte';
	import WtsISP from './Ticket/WtsISP.svelte';
	import { AccessMap } from '$lib/types/enum';

	let { u }: { u: UserProfile } = $props();

	const handleClick = () => goto(`/admin/user_tickets?${new URLSearchParams({ wx: u.wx, name: u.name }).toString()}`);

	const handleKeydown = (event: KeyboardEvent) => {
		if (event.key !== 'Enter' && event.key !== ' ') {
			return;
		}

		event.preventDefault();
		handleClick();
	};
</script>

<RetroCard
	style="cursor: pointer; outline: none;"
	role="button"
	tabindex="0"
	onclick={handleClick}
	onkeydown={handleKeydown}
>

	<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
		<strong style="flex-shrink: 0;width: 7em;">姓名</strong>
		<div style="font-size: 15px;">{u.name}</div>
	</div>

	<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
		<strong style="flex-shrink: 0;width: 7em;">学号</strong>
		<p style="font-size: 15px;">{u.sid}</p>
	</div>

	<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
		<strong style="flex-shrink: 0;width: 7em;">宿舍</strong>
		<p style="font-size: 15px;"><BlockRoom b={u.block} r={u.room} /></p>
	</div>

	<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
		<strong style="flex-shrink: 0;width: 7em;">联系电话</strong>
		<p style="font-size: 15px;">{u.phone}</p>
	</div>

	<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
		<strong style="flex-shrink: 0;width: 7em;">校园网账号</strong>
		<p style="font-size: 15px;">{u.account}，<WtsISP i={u.isp} /></p>
	</div>

	<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
		<strong style="flex-shrink: 0;width: 7em;">权限</strong>
		<p style="font-size: 15px;">{AccessMap[u.access]}</p>
	</div>

	<div class="flex items-baseline" style="margin-top: 12.5px; font-size: 15.5px;">
		<strong style="flex-shrink: 0;width: 7em;">微信OpenID</strong>
		<p style="font-size: 15px;">{u.wx}</p>
	</div>
</RetroCard>


