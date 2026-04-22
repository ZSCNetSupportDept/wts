<script lang="ts">
	import type { PageProps } from './$types';
	import { TextInput, Button, MultiSelect } from 'carbon-components-svelte';
	import type { FilterUsersReq } from '$lib/types/apiRequest';
	import { BlockMap, ISPMap } from '$lib/types/enum';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Guard } from '$lib/jwt';
	import { IsAdmin, IsAccessIn } from '$lib/types/enum';



	let { data }: PageProps = $props();

	const blocks = Object.entries(BlockMap).map(([id, text]) => ({ id, text }));
	const isps = Object.entries(ISPMap).map(([id, text]) => ({ id, text }));

	let r = $state({} as FilterUsersReq);

	function handleSubmit() {
		const params = new URLSearchParams();

		for (const [key, value] of Object.entries(r)) {
			if (!value) continue; // 跳过空值

			if (Array.isArray(value)) {
				// 如果是数组，使用 append 追加多个同名参数
				value.forEach((v) => params.append(key, v));
			} else {
				params.set(key, value as string);
			}
		}

		goto(`/admin/filter_users/result?${params.toString()}`);
	}

	onMount(() => Guard(IsAccessIn('dev')));
</script>

<h1>用户查找</h1>
<br />
<hr />
<br />
<p>根据一系列条件筛选已经在报修系统中注册过的用户，条件之间遵循“与”逻辑。</p>
<br />
<strong
	>警告：该功能仅供确定运营商指派工单所属用户使用，不要用来干坏事，你的每一次查询在后台都有记录。</strong
>
<br />
<br />

<TextInput id="name" labelText="用户的姓名" placeholder="可模糊搜索" bind:value={r.name} />
<br />
<MultiSelect
	id="block"
	labelText="用户的宿舍楼"
	label="选择宿舍楼"
	items={blocks}
	bind:selectedIds={r.block}
/>
<br />
<MultiSelect
	id="isp"
	labelText="用户的运营商"
	label="选择运营商"
	items={isps}
	bind:selectedIds={r.isp}
/>
<br />
<TextInput id="room" labelText="用户的宿舍号" placeholder="可模糊搜索" bind:value={r.room} />
<br />
<TextInput
	id="phone"
	labelText="用户的联系电话"
	placeholder="可模糊搜索"
	bind:value={r.phone}
/>
<br />
<TextInput
	id="account"
	labelText="用户的校园网账号"
	placeholder="可模糊搜索"
	bind:value={r.account}
/>
<br />
<Button on:click={handleSubmit}>
	提交
</Button>
