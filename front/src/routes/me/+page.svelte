<script lang="ts">
	import RetroCard from '$lib/components/RetroCard.svelte';
	import type { UserProfile } from '$lib/types/apiResponse';
	import {
		Accordion,
		AccordionItem,
		Toggle,
		Button,
		StructuredList,
		StructuredListBody,
		StructuredListCell,
		StructuredListRow,
		Modal,
		NotificationQueue,
		ButtonSet
	} from 'carbon-components-svelte';
	import BlockRoom from '$lib/components/Ticket/BlockRoom.svelte';
	import { ViewProfile } from '$lib/api';
	import { CheckAndGetJWT, Guard } from '$lib/jwt';
	import { onMount } from 'svelte';
	import { IsAdmin, IsOperator, IsUser } from '$lib/types/enum';
	import WtsISP from '$lib/components/Ticket/WtsISP.svelte';
	import Renew from 'carbon-icons-svelte/lib/Renew.svelte';
	import { TheLastPage } from '$lib/states/theLastPage.svelte';
	import { goto } from '$app/navigation';

	let pending = $state(true);
	let info = $state({} as UserProfile);
	let q: NotificationQueue;

	onMount(() => (Guard(IsUser), fetchUser()));

	async function fetchUser() {
		try {
			const wx = CheckAndGetJWT('parsed').openid;
			if (!wx) {
				throw new Error('未找到用户信息，请重新登录');
			}
			const res = await ViewProfile(wx);
			console.log(res);
			pending = false;
			if (!res.success) {
				throw new Error(res.msg || '获取用户信息失败');
			}
			info = res.profile;
		} catch (e: any) {
			pending = false;
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '获取用户信息失败',
				subtitle: errMsg + '，请重试',
				timeout: 5000
			});
		}
	}
</script>

<h1>我</h1>
<br />
<hr />
<br />
<div class="profile">
	{#await pending}
		<p>加载中...</p>
	{:then}
		<RetroCard>
			<span style="display: flex; align-items: center;">
				<h2 style="margin-right: 0.5rem;">个人信息</h2>
				<Renew
					onclick={() => {
						(TheLastPage.Write('/me'), goto('/login'));
					}}
					style="cursor: pointer;"
				/>
			</span>
			<StructuredList style="margin-bottom: 1rem;">
				<StructuredListBody>
					<StructuredListRow>
						<StructuredListCell noWrap>姓名</StructuredListCell>
						<StructuredListCell>{info.name}</StructuredListCell>
					</StructuredListRow>
					<StructuredListRow>
						<StructuredListCell noWrap>学号</StructuredListCell>
						<StructuredListCell>{info.sid}</StructuredListCell>
					</StructuredListRow>
					<StructuredListRow>
						<StructuredListCell noWrap>联系电话</StructuredListCell>
						<StructuredListCell>{info.phone}</StructuredListCell>
					</StructuredListRow>
					<StructuredListRow>
						<StructuredListCell noWrap>宿舍地址</StructuredListCell>
						<StructuredListCell><BlockRoom b={info.block} r={info.room} /></StructuredListCell>
					</StructuredListRow>
					<StructuredListRow>
						<StructuredListCell noWrap>账号</StructuredListCell>
						<StructuredListCell>{info.account}</StructuredListCell>
					</StructuredListRow>
					<StructuredListRow>
						<StructuredListCell noWrap>运营商</StructuredListCell>
						<StructuredListCell><WtsISP i={info.isp} /></StructuredListCell>
					</StructuredListRow>
				</StructuredListBody>
			</StructuredList>
			<p style="font-size:small;margin-bottom:1rem;">
				该信息用于我们提供上门服务，如有变化请修改：
			</p>
			<Button on:click={() => (window.location.href = '/me/update')}>修改信息</Button>
		</RetroCard>
	{/await}

	{#if IsOperator(info.access)}
		<br />
		<br />
		<br />
		<RetroCard style="margin-bottom:1rem;">
			<h2>网维操作</h2>
			<p>在这里进入网维后台系统</p>
			<br />
			<ButtonSet stacked>
				<Button on:click={() => (window.location.href = '/op')}>进入后台</Button>
				{#if IsAdmin(info.access)}
					<Button kind="ghost" on:click={() => (window.location.href = '/admin')}
						>进入管理后台</Button
					>
				{/if}
			</ButtonSet>
		</RetroCard>
	{/if}
</div>

<br />
<br />
<br />
<RetroCard>
	<Accordion size="xl">
		<AccordionItem title="设置">
			<div>
				<!-- <Theme bind:theme persist persistKey="__carbon-theme" /> -->
				<Toggle labelText="深色模式(暂时用不了)" disabled />
			</div>
		</AccordionItem>
		<AccordionItem title="联系我们">
			<p>如果您对网维的服务或本系统有任何意见或建议，请尽管联系我们！我们非常重视您的建议。</p>
			<br />
			<p>ZSC学生网络支撑QQ群:123123123</p>
			<br />
			<p>科长QQ:</p>
			<br />
			<p>科长微信/电话：</p>
			<br />
		</AccordionItem>
		<AccordionItem title="关于">
			<a href="/about">关于网络维护科</a>
			<br />
			<br />
			<hr />
			<br />
			<br />
			<p>作者：paakaauuxx</p>
			<br />
			<p>前端框架：SvelteKit</p>
			<br />
			<p>UI框架：Carbon Components Svelte</p>
			<br />
			<p>后端框架：Go(Echo)+PostgreSQL(sqlc)</p>
			<br />
			<p>ZSCNetworkSupport 版权所有</p>
			<br />
			<p>Under AGPLv3,<a href="https://github.com/ZSCNetSupportDept/wts">源代码</a></p>
		</AccordionItem>
	</Accordion>
</RetroCard>

<NotificationQueue bind:this={q} />
