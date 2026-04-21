<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	import { onMount } from 'svelte';
	import { Guard } from '$lib/jwt';
	import { IsAdmin } from '$lib/types/enum';

	onMount(() => Guard(IsAdmin));
</script>

<h1>管理后台</h1>
<br />
<hr />
<br />
<p>正在开发中</p>
<a href="/admin/new_assignment">录入工单</a>
<br />
<a href="/admin/add_ticket">增添报修</a>
<br />
<a href="/admin/filter_users">查找用户</a>