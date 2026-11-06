<script lang="ts">
	import { onMount } from 'svelte';
	import { Button, ClickableTile, Tag } from 'carbon-components-svelte';
	import EventSchedule from 'carbon-icons-svelte/lib/EventSchedule.svelte';
	import Report from 'carbon-icons-svelte/lib/Report.svelte';
	import Add from 'carbon-icons-svelte/lib/Add.svelte';
	import Search from 'carbon-icons-svelte/lib/Search.svelte';
	import ArrowRight from 'carbon-icons-svelte/lib/ArrowRight.svelte';
	import { Guard, CheckAndGetJWT } from '$lib/jwt';
	import { IsAdmin } from '$lib/types/enum';
	import { schedulerRoles } from '$lib/scheduler';

	let name = $state('');
	let role = $state('管理层');
	const tools = [
		{ title: '管理排班', description: '配置值班小组、安排片区与添加蹭班。', href: '/admin/scheduler/', icon: EventSchedule, label: '排班管理' },
		{ title: '录入工单', description: '登记工单并安排处理成员。', href: '/admin/new_assignment/', icon: Report, label: '工单管理' },
		{ title: '增添报修', description: '协助用户登记新的报修需求。', href: '/admin/add_ticket/', icon: Add, label: '报修服务' },
		{ title: '查找用户', description: '检索用户信息与相关报修记录。', href: '/admin/filter_users/', icon: Search, label: '用户查询' }
	];
	onMount(() => {
		Guard(IsAdmin);
		const user = CheckAndGetJWT('parsed');
		if (user) { name = user.name; role = schedulerRoles[user.access] || '管理层'; }
	});
</script>

<svelte:head><title>管理层后台 · 网维报修</title></svelte:head>
<section class="management" aria-label="管理层后台">
	<header class="intro">
		<div class="eyebrow">网维 · 管理中心</div>
		<div class="heading"><h1>管理层后台</h1><Tag type="blue">{role}</Tag></div>
		<p>{name ? `${name}，` : ''}从这里开始处理值班、工单和用户事务。</p>
	</header>
	<div class="section-heading"><h2>常用功能</h2><span>按职责进入对应工作区</span></div>
	<div class="tools">
		{#each tools as tool}
			<ClickableTile href={tool.href} class="management-tile">
				<div class="tile-top"><tool.icon size={24} /><span>{tool.label}</span></div>
				<h3>{tool.title}</h3><p>{tool.description}</p>
				<div class="tile-action"><span>进入</span><ArrowRight size={20} /></div>
			</ClickableTile>
		{/each}
	</div>
	<div class="secondary">
		<h2>值班协作</h2><p>查看个人值班安排，或进入换班大厅处理成员请求。</p>
		<div class="links"><Button kind="tertiary" size="small" href="/op/scheduler/">我的排班</Button><Button kind="ghost" size="small" href="/op/scheduler/swaps/">换班请求</Button></div>
	</div>
</section>
<style>
	.management { color: var(--cds-text-01, #161616); }
	.intro { padding: 1rem 0 1.5rem; border-bottom: 1px solid var(--cds-ui-03, #e0e0e0); }
	.eyebrow { color: var(--cds-text-02, #525252); font-size: .75rem; letter-spacing: .08em; margin-bottom: .75rem; }
	.heading { display: flex; align-items: center; justify-content: space-between; gap: .5rem; }
	h1 { font-size: 1.75rem; line-height: 1.3; font-weight: 400; margin: 0; }
	.intro p { margin-top: .75rem; color: var(--cds-text-02, #525252); font-size: .875rem; line-height: 1.6; }
	.section-heading { margin: 1.5rem 0 .75rem; }
	h2 { font-size: 1rem; font-weight: 600; margin: 0; }
	.section-heading span { display: block; font-size: .75rem; color: var(--cds-text-02, #525252); margin-top: .375rem; }
	.tools { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1px; background: var(--cds-ui-03, #e0e0e0); border: 1px solid var(--cds-ui-03, #e0e0e0); }
	.tools :global(.management-tile) { display: flex; flex-direction: column; padding: 1rem; min-height: 190px; color: var(--cds-text-01, #161616); text-decoration: none; background: var(--cds-ui-01, #fff); }
	.tools :global(.management-tile:hover) { background: var(--cds-hover-ui, #e5e5e5); }
	.tile-top { display: flex; align-items: center; justify-content: space-between; gap: .25rem; color: var(--cds-link-01, #0f62fe); }
	.tile-top span { color: var(--cds-text-02, #525252); font-size: .6875rem; }
	h3 { font-size: 1.125rem; font-weight: 600; margin: 1.25rem 0 .5rem; }
	.tools p { color: var(--cds-text-02, #525252); font-size: .8125rem; line-height: 1.5; margin: 0 0 1rem; }
	.tile-action { display: flex; justify-content: space-between; align-items: center; margin-top: auto; color: var(--cds-link-01, #0f62fe); font-size: .8125rem; }
	.secondary { margin-top: 1.5rem; padding: 1rem; border-left: 3px solid var(--cds-interactive-01, #0f62fe); background: var(--cds-ui-01, #fff); }
	.secondary p { color: var(--cds-text-02, #525252); font-size: .875rem; line-height: 1.5; margin: .5rem 0 1rem; }
	.links { display: flex; gap: .5rem; flex-wrap: wrap; }
	@media (max-width: 350px) { .tools { grid-template-columns: 1fr; } .heading { flex-wrap: wrap; } }
</style>
