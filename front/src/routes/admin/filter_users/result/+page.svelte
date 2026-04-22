<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import type { FilterUsersReq } from '$lib/types/apiRequest';
	import type { FilterUsersRes } from '$lib/types/apiResponse';
	import type { WtsBlock, WtsISP } from '$lib/types/enum';
	import type { UserProfile as UserProfileType } from '$lib/types/apiResponse';
	import { Button, NotificationQueue } from 'carbon-components-svelte';
	import Return from 'carbon-icons-svelte/lib/Return.svelte';
	import { FilterUsers } from '$lib/api';
	import UserProfile from '$lib/components/UserProfile.svelte';
	import { Guard } from '$lib/jwt';
	import { IsAdmin } from '$lib/types/enum';

	let ok = $state(false);
	let users: UserProfileType[] = $state([] as UserProfileType[]);
	let userEmpty = $state(false);

	let r: FilterUsersReq = $state({
		name: page.url.searchParams.get('name') ?? undefined,
		phone: page.url.searchParams.get('phone') ?? undefined,
		block:
			page.url.searchParams.getAll('block').length > 0
				? page.url.searchParams.getAll('block').map((b) => b as WtsBlock)
				: undefined,
		room: page.url.searchParams.get('room') ?? undefined,
		isp:
			page.url.searchParams.getAll('isp').length > 0
				? page.url.searchParams.getAll('isp').map((i) => i as WtsISP)
				: undefined,
		account: page.url.searchParams.get('account') ?? undefined
	} as FilterUsersReq);

	onMount(fetchUsers);

	async function fetchUsers() {
		ok = false;
		try {
			let r1: FilterUsersRes = await FilterUsers(r);
			if (!r1.success) {
				throw new Error(r1.msg || '获取用户失败');
			}
			users = r1.profiles;
			if (!users || users.length === 0) {
				userEmpty = true;
			}
			ok = true;
		} catch (e: any) {
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '获取用户失败',
				subtitle: errMsg,
				timeout: 3000
			});
			return;
		}
	}

	let q: NotificationQueue;

	onMount(() => Guard(IsAdmin));
</script>

<h1>用户查找结果</h1>
<br />
<hr />
<br />
<p>按照您提供的条件，获得的检索结果。</p>
<br />
<div
	style="display: flex; justify-content: flex-end; transform: translate(-17px,0px); margin-bottom: 15px;"
>
	<Button href="/admin/filter_users">修改条件<Return /></Button>
</div>

{#if !ok}
	<p>处理中，请稍等...</p>
{/if}
{#if userEmpty === false}
	{#each users as u}
		<UserProfile {u} />
	{/each}
{:else}
	<span>没有找到符合条件的用户。</span>
{/if}

<NotificationQueue bind:this={q} />
