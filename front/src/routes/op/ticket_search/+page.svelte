<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	import { CheckAndGetJWT, Guard } from '$lib/jwt';
	import { IsAdmin, IsOperator, PriorityMap, CategoryMap, ISPMap } from '$lib/types/enum';
	import { onMount } from 'svelte';
	import {
		RadioButtonGroup,
		RadioButton,
		Checkbox,
		CheckboxGroup,
		DatePicker,
		DatePickerInput,
		TimePicker,
		Grid,
		Row,
		Column,
		NumberInput,
		Button,
		NotificationQueue,
		Toggle
	} from 'carbon-components-svelte';
	import type { FilterTicketsReq } from '$lib/types/apiRequest';
	import { ZoneMap, ZoneToBlock, StatusMap, type WtsZone } from '$lib/types/enum';
	import { RFC3339 } from '$lib/types/RFC3339';
	import { startOfDay, endOfDay } from 'date-fns';
	import { criteria } from '$lib/states/ticketCriteriaSearch.svelte';
	import { goto } from '$app/navigation';
	import type { WtsStatus, WtsPriority, WtsCategory, WtsISP } from '$lib/types/enum';

	onMount(() => Guard(IsOperator));

	let req = $state(criteria.r as FilterTicketsReq);

	let zoneSelected: WtsZone[] = $state(criteria._blocks_in_zone ?? []);

	let order: 'priority' | 'newest' | 'oldest' = $state(criteria._order ?? 'priority');
	let floor: number | null = $state(criteria._floor ?? null);
	let viewTodayScheduled = $state(criteria._view_today_scheduled ?? false);

	let isScheduledSelected = $state(req.status?.includes('scheduled') ?? false);

	// $effect(() => {
	// 	$inspect(req);
	// 	$inspect(zoneSelected);
	// 	$inspect(order);
	// 	$inspect(floor);
	// });

	let onDateChange = (which: 'newer' | 'older') => (event: CustomEvent) => {
		const { dateStr } = event.detail;
		if (dateStr) {
			const date = new Date(dateStr);
			const adjustedDate = which === 'newer' ? startOfDay(date) : endOfDay(date);
			const rfcDate = RFC3339(adjustedDate);
			if (which === 'newer') {
				req.newer_than = rfcDate;
			} else {
				req.older_than = rfcDate;
			}
		}
	};

	let q: NotificationQueue;

	function search() {
		req.block = zoneSelected.flatMap((zone) => ZoneToBlock[zone as WtsZone]);
		criteria.r = $state.snapshot(req);
		criteria._blocks_in_zone = $state.snapshot(zoneSelected) as WtsZone[];
		criteria._order = $state.snapshot(order);
		criteria._floor = $state.snapshot(floor);
		criteria._view_today_scheduled = $state.snapshot(viewTodayScheduled);
		console.log(criteria);
		setTimeout(() => goto('/op/tickets'), 500);
	}

	const allZones = Object.keys(ZoneMap) as WtsZone[];
	const allStatuses = IsAdmin(CheckAndGetJWT('parsed').access)
		? (Object.keys(StatusMap) as WtsStatus[])
		: (Object.keys(StatusMap).filter(
				(status) => status !== 'solved' && status !== 'canceled'
			) as WtsStatus[]);
	const allPriorities = Object.keys(PriorityMap) as WtsPriority[];
	const allCategories = Object.keys(CategoryMap) as WtsCategory[];
	const allISPs = Object.keys(ISPMap) as WtsISP[];

	const zoneOptions = [
		'FX',
		'BM',
		'DM',
		'QT',
		'XHAB',
		'XHCD',
		'ZH',
		'other'
	] as const satisfies readonly WtsZone[];

	const statusOptionsAdmin = [
		'fresh',
		'scheduled',
		'delay',
		'escalated',
		'solved',
		'canceled'
	] as const satisfies readonly WtsStatus[];
	const statusOptionsUser = [
		'fresh',
		'scheduled',
		'delay',
		'escalated'
	] as const satisfies readonly WtsStatus[];
	const statusOptions: readonly WtsStatus[] = IsAdmin(CheckAndGetJWT('parsed').access)
		? statusOptionsAdmin
		: statusOptionsUser;

	const priorityOptions = [
		'highest',
		'assigned',
		'mainline',
		'normal',
		'in-passing',
		'least'
	] as const satisfies readonly WtsPriority[];

	const categoryOptions = [
		'first-install',
		'low-speed',
		'ip-or-device',
		'client-or-account',
		'others'
	] as const satisfies readonly WtsCategory[];

	const ispOptions = ['telecom', 'unicom', 'mobile', 'others'] as const satisfies readonly WtsISP[];

	function allSelected(selected: readonly string[] | null | undefined, options: readonly string[]) {
		if (!selected) return false;
		return options.length > 0 && options.every((o) => selected.includes(o));
	}

	function uniq<T>(arr: T[]) {
		return Array.from(new Set(arr));
	}

	function sameArray<T>(a: readonly T[], b: readonly T[]) {
		if (a.length !== b.length) return false;
		for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
		return true;
	}

	$effect(() => {
		req.status ??= [];
		req.priority ??= [];
		req.category ??= [];
		req.isp ??= [];

		const nextZone = uniq(zoneSelected).filter((z) =>
			(zoneOptions as readonly WtsZone[]).includes(z)
		);
		if (!sameArray(zoneSelected, nextZone)) zoneSelected = nextZone;

		const nextStatus = uniq(req.status).filter((s) =>
			(statusOptions as readonly WtsStatus[]).includes(s)
		);
		if (!sameArray(req.status, nextStatus)) req.status = nextStatus;

		const nextPriority = uniq(req.priority).filter((p) =>
			(priorityOptions as readonly WtsPriority[]).includes(p)
		);
		if (!sameArray(req.priority, nextPriority)) req.priority = nextPriority;

		const nextCategory = uniq(req.category).filter((c) =>
			(categoryOptions as readonly WtsCategory[]).includes(c)
		);
		if (!sameArray(req.category, nextCategory)) req.category = nextCategory;

		const nextIsp = uniq(req.isp).filter((i) => (ispOptions as readonly WtsISP[]).includes(i));
		if (!sameArray(req.isp, nextIsp)) req.isp = nextIsp;
	});

	$effect(() =>{
		isScheduledSelected = req.status?.includes('scheduled') ?? false;
	})
</script>

<h1>报修单检索</h1>
<br />
<hr />
<br />
<p>选择您需要检索报修工单的条件</p>

{#if IsAdmin(CheckAndGetJWT('parsed').access)}
	<br />
	<RadioButtonGroup id="scope" legendText="范围" bind:selected={req.scope} required={true}>
		<RadioButton labelText="只看活跃的" value="active" />
		<RadioButton labelText="所有报修单" value="all" />
	</RadioButtonGroup>
{/if}

<br />
<DatePicker datePickerType="single" on:change={onDateChange('newer')}>
	<DatePickerInput labelText="只看这之后的报修单：" placeholder="从该日开始时起" />
</DatePicker>

<br />
<DatePicker datePickerType="single" on:change={onDateChange('older')}>
	<DatePickerInput labelText="只看这之前的报修单：" placeholder="从该日结束时起" />
</DatePicker>
<!-- TODO:可以选择时间 -->

<br />
<CheckboxGroup legendText="片区" id="block" bind:selected={zoneSelected} required={true}>
	<Grid narrow>
		<Row>
			<Column sm={2} md={2} lg={4}><Checkbox value="FX" labelText={ZoneMap['FX']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="BM" labelText={ZoneMap['BM']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="DM" labelText={ZoneMap['DM']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="QT" labelText={ZoneMap['QT']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="XHAB" labelText={ZoneMap['XHAB']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="XHCD" labelText={ZoneMap['XHCD']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="ZH" labelText={ZoneMap['ZH']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="other" labelText={ZoneMap['other']} /></Column>
		</Row>
	</Grid>
</CheckboxGroup>
<div class="toggle">
	<Toggle
		size="sm"
		toggled={allSelected(zoneSelected, zoneOptions)}
		on:toggle={(e) => {
			const { toggled } = e.detail as { toggled: boolean };
			zoneSelected = toggled ? [...zoneOptions] : [];
		}}
	>
		<span slot="labelA">全不选</span>
		<span slot="labelB">全选</span>
	</Toggle>
</div>

<br />
<br />
<CheckboxGroup legendText="状态" id="status" bind:selected={req.status} required={true}>
	<Grid narrow>
		<Row>
			<Column sm={2} md={2} lg={4}><Checkbox value="fresh" labelText={StatusMap['fresh']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="scheduled" labelText={StatusMap['scheduled']} /></Column
			>
			<Column sm={2} md={2} lg={4}><Checkbox value="delay" labelText={StatusMap['delay']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="escalated" labelText={StatusMap['escalated']} /></Column
			>
			{#if IsAdmin(CheckAndGetJWT('parsed').access)}
				<Column sm={2} md={2} lg={4}
					><Checkbox value="solved" labelText={StatusMap['solved']} /></Column
				>
				<Column sm={2} md={2} lg={4}
					><Checkbox value="canceled" labelText={StatusMap['canceled']} /></Column
				>
			{/if}
		</Row>
	</Grid>
</CheckboxGroup>
<div class="toggle">
	<Toggle
		size="sm"
		toggled={allSelected(req.status, statusOptions)}
		on:toggle={(e) => {
			const { toggled } = e.detail as { toggled: boolean };
			req.status = toggled ? [...statusOptions] : [];
		}}
	>
		<span slot="labelA">全不选</span>
		<span slot="labelB">全选</span>
	</Toggle>
</div>

<br />
<br />
<CheckboxGroup legendText="优先级" id="priority" bind:selected={req.priority} required={true}>
	<Grid narrow>
		<Row>
			<Column sm={2} md={2} lg={4}><Checkbox value="highest" labelText="最高" /></Column>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="assigned" labelText={PriorityMap['assigned']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="mainline" labelText={PriorityMap['mainline']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="normal" labelText={PriorityMap['normal']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="in-passing" labelText={PriorityMap['in-passing']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="least" labelText={PriorityMap['least']} /></Column
			>
		</Row>
	</Grid>
</CheckboxGroup>
<div class="toggle">
	<Toggle
		size="sm"
		toggled={allSelected(req.priority, priorityOptions)}
		on:toggle={(e) => {
			const { toggled } = e.detail as { toggled: boolean };
			req.priority = toggled ? [...priorityOptions] : [];
		}}
	>
		<span slot="labelA">全不选</span>
		<span slot="labelB">全选</span>
	</Toggle>
</div>
<br />
<br />
<CheckboxGroup legendText="类型" id="category" bind:selected={req.category} required={true}>
	<Grid narrow>
		<Row>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="first-install" labelText={CategoryMap['first-install']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="client-or-account" labelText={CategoryMap['client-or-account']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="ip-or-device" labelText={CategoryMap['ip-or-device']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="low-speed" labelText={CategoryMap['low-speed']} /></Column
			>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="others" labelText={CategoryMap['others']} /></Column
			>
		</Row>
	</Grid>
</CheckboxGroup>
<div class="toggle">
	<Toggle
		size="sm"
		toggled={allSelected(req.category, categoryOptions)}
		on:toggle={(e) => {
			const { toggled } = e.detail as { toggled: boolean };
			req.category = toggled ? [...categoryOptions] : [];
		}}
	>
		<span slot="labelA">全不选</span>
		<span slot="labelB">全选</span>
	</Toggle>
</div>

<br />
<br />
<CheckboxGroup legendText="运营商" id="isp" bind:selected={req.isp} required={true}>
	<Grid narrow>
		<Row>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="telecom" labelText={ISPMap['telecom']} /></Column
			>
			<Column sm={2} md={2} lg={4}><Checkbox value="unicom" labelText={ISPMap['unicom']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="mobile" labelText={ISPMap['mobile']} /></Column>
			<Column sm={2} md={2} lg={4}><Checkbox value="others" labelText={ISPMap['others']} /></Column>
			<Column sm={2} md={2} lg={4}
				><Checkbox value="broadnet" labelText={ISPMap['broadnet']} hidden /></Column
			>
			<!--暂时藏起来-->
		</Row>
	</Grid>
</CheckboxGroup>
<div class="toggle">
	<Toggle
		size="sm"
		toggled={allSelected(req.isp, ispOptions)}
		on:toggle={(e) => {
			const { toggled } = e.detail as { toggled: boolean };
			req.isp = toggled ? [...ispOptions] : [];
		}}
	>
		<span slot="labelA">全不选</span>
		<span slot="labelB">全选</span>
	</Toggle>
</div>

<br />
<br />
<hr />
<br />
<RadioButtonGroup id="order" legendText="排序" bind:selected={order} required={true}>
	<RadioButton labelText="优先级从高到低" value="priority" />
	<RadioButton labelText="时间从新到旧" value="newest" />
	<RadioButton labelText="时间从旧到新" value="oldest" />
</RadioButtonGroup>

<br />
<br />
<NumberInput
	min={1}
	max={20}
	step={1}
	bind:value={floor}
	allowEmpty={true}
	allowDecimal={false}
	labelText="只看如下楼层(不填代表查看全部楼层)"
/>

<br />
<br />
<Toggle labelText="只显示预约在今天的预约单" bind:toggled={viewTodayScheduled} disabled={!isScheduledSelected}/>
<br />
<br />
<Button on:click={search}>搜索</Button>

<NotificationQueue bind:this={q} />