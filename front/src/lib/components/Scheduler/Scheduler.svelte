<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Button,
		Checkbox,
		DatePicker,
		DatePickerInput,
		Select,
		SelectItem,
		MultiSelect,
		TextInput,
		TextArea,
		InlineLoading,
		InlineNotification,
		Tile,
		Tag,
		Accordion,
		AccordionItem,
		Table,
		TableHead,
		TableBody,
		TableRow,
		TableHeader,
		TableCell
	} from 'carbon-components-svelte';
	import { Mandarin } from 'flatpickr/dist/l10n/zh.js';
	import { GuardAndContinue } from '$lib/jwt';
	import { IsOperator, ZoneMap } from '$lib/types/enum';
	import {
		schedulerApi,
		schedulerRoles,
		memberColor,
		chinaToday,
		calendarDate,
		type SchedulerOverview,
		type SchedulerMember,
		type Assignment,
		type Swap,
		type AuditEntry
	} from '$lib/scheduler';

	let { manage = false, hall = false }: { manage?: boolean; hall?: boolean } = $props();
	const zones = Object.entries(ZoneMap).filter(([key]) => key !== 'all');
	const weekLabels = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
	const statuses: Record<string, string> = {
		open: '待组员接单',
		pending: '等待组长确认',
		approved: '已换班',
		rejected: '已拒绝',
		canceled: '已撤销',
		stale: '原排班已变化，请重新申请'
	};
	let selectedDate = $state(chinaToday());
	let month = $state(chinaToday().slice(0, 7));
	let overview = $state<SchedulerOverview | null>(null);
	let loading = $state(false);
	let busy = $state(false);
	let error = $state('');
	let notice = $state('');
	let noticeKind = $state<'success' | 'info'>('success');
	let staff = $state<SchedulerMember[]>([]);
	let staffLoadedDate = $state('');
	let leaderWid = $state('');
	let selectedWids = $state<string[]>([]);
	let assignments = $state<Assignment[]>([]);
	let reason = $state('');
	let overrideUnavailable = $state(false);
	let extraWid = $state('');
	let extraZone = $state('');
	let freeDays = $state<number[]>([]);
	let exception = $state('');
	let swapDateA = $state(chinaToday());
	let swapDateB = $state('');
	let swapReason = $state('');
	let swaps = $state<Swap[]>([]);
	let audit = $state<AuditEntry[]>([]);
	let calendar = $state.raw<any>(null);
	let requestSequence = 0;
	let staffSequence = 0;
	let selectedDay = $derived(overview?.days.find((day) => day.date === selectedDate));
	let canArrange = $derived(
		!!overview &&
			(overview.manager ||
				(overview.me.access === 'group-leader' && selectedDay?.leader.wid === overview.me.wid))
	);
	let canManage = $derived(
		!!overview && (overview.manager || overview.me.access === 'group-leader')
	);
	let weekdayLabel = $derived(weekLabels[(new Date(`${selectedDate}T12:00:00`).getDay() + 6) % 7]);
	let groupItems = $derived(
		staff.map((p) => ({
			id: p.wid,
			text: `${p.name}（${p.wid}）· ${schedulerRoles[p.access] || p.access}`
		}))
	);
	let myDays = $derived(
		overview?.days.filter(
			(day) => day.published && day.members.some((m) => m.wid === overview?.me.wid)
		) || []
	);
	let calendarOptions = {
		inline: true,
		static: true,
		onReady: (_dates: Date[], _text: string, instance: any) => {
			calendar = instance;
		},
		onDayCreate: (_dates: Date[], _text: string, _instance: any, element: any) => {
			const date = calendarDate(element.dateObj);
			if (myDays.some((day) => day.date === date)) {
				element.classList.add('scheduler-duty-date');
				element.setAttribute('aria-label', `${date}，需要值班`);
				element.title = '需要值班 16:30–18:00';
			}
		},
		onMonthChange: (_dates: Date[], _text: string, instance: any) => {
			const next = `${instance.currentYear}-${String(instance.currentMonth + 1).padStart(2, '0')}`;
			if (next !== month) {
				month = next;
				void load();
			}
		},
		onYearChange: (_dates: Date[], _text: string, instance: any) => {
			const next = `${instance.currentYear}-${String(instance.currentMonth + 1).padStart(2, '0')}`;
			if (next !== month) {
				month = next;
				void load();
			}
		}
	};
	$effect(() => {
		overview?.days;
		calendar?.redraw();
	});
	$effect(() => {
		const day = selectedDay;
		leaderWid = day?.leader.wid || '';
		selectedWids = day?.members.filter((m) => !m.extra).map((m) => m.wid) || [];
		assignments =
			day?.members
				.filter((m) => !m.extra)
				.map((m) => ({ wid: m.wid, zone: m.zone, area_leader: m.area_leader })) || [];
		exception = String(overview?.exceptions.find((e) => e.date === selectedDate)?.available ?? '');
		audit = [];
	});
	$effect(() => {
		selectedDate;
		overview?.manager;
		selectedDay?.leader.wid;
		if (manage && (overview?.manager || canArrange)) void loadStaff(selectedDate);
		else {
			staffSequence++;
			staff = [];
		}
	});
	function message(e: any): string {
		return e.response?.data?.msg || e.message || '操作失败，请重试';
	}
	async function load() {
		const sequence = ++requestSequence;
		loading = true;
		error = '';
		try {
			const [data, requests] = await Promise.all([
				schedulerApi.overview(month),
				schedulerApi.swaps(month)
			]);
			if (sequence !== requestSequence) return;
			overview = data;
			if (manage && !data.manager && data.me.access !== 'group-leader') {
				void goto('/op/scheduler/');
				return;
			}
			swaps = requests;
			freeDays = [...data.free_days];
		} catch (e) {
			if (sequence === requestSequence) {
				error = message(e);
				overview = null;
				swaps = [];
			}
		} finally {
			if (sequence === requestSequence) loading = false;
		}
	}
	async function loadStaff(date: string) {
		const sequence = ++staffSequence;
		try {
			const data = await schedulerApi.staff(date);
			if (sequence === staffSequence) {
				staff = data;
				staffLoadedDate = date;
			}
		} catch (e) {
			if (sequence === staffSequence) {
				staff = [];
				error = message(e);
			}
		}
	}
	async function perform(fn: () => Promise<any>, success: string) {
		if (busy) return;
		busy = true;
		error = '';
		notice = '';
		try {
			await fn();
			notice = success;
			noticeKind = 'success';
			await load();
		} catch (e) {
			error = message(e);
		} finally {
			busy = false;
		}
	}
	function payload() {
		return { date: selectedDate, revision: selectedDay?.revision || 0, reason };
	}
	async function change(action: string, extra: object = {}) {
		await perform(
			() => schedulerApi.change(action, { ...payload(), ...extra }),
			action === 'publish' ? '排班已发布' : '已保存'
		);
	}
	async function suggest(mode: 'reuse' | 'random') {
		if (busy) return;
		busy = true;
		error = '';
		notice = '';
		try {
			const data = await schedulerApi.suggest({ ...payload(), mode });
			assignments = data.assignments;
			noticeKind = 'info';
			notice = `${mode === 'reuse' ? '已沿用' : '已随机生成'} ${data.previous_date} 的片区安排。请检查后点击“保存片区安排”；尚未提交。`;
		} catch (e) {
			error = message(e);
		} finally {
			busy = false;
		}
	}
	function selected(event: CustomEvent) {
		const date = event.detail.dateStr;
		if (typeof date !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(date)) return;
		selectedDate = date;
		if (month !== date.slice(0, 7)) {
			month = date.slice(0, 7);
			void load();
		}
	}
	function toggleFree(index: number, checked: boolean) {
		freeDays = checked
			? [...new Set([...freeDays, index])]
			: freeDays.filter((day) => day !== index);
	}
	async function saveAvailability() {
		await perform(
			() =>
				schedulerApi.availability({
					free_days: freeDays.sort(),
					date: selectedDate,
					available: exception === '' ? null : exception === 'true'
				}),
			'空闲时间已保存；已有班次不自动变更'
		);
	}
	async function review(request: Swap, action: string) {
		await perform(
			() => schedulerApi.reviewSwap({ id: request.id, action }),
			action === 'approve'
				? '已确认；双方组长均确认后生效'
				: action === 'claim'
					? '已接单，等待双方组长确认'
					: '申请已处理'
		);
	}
	async function showAudit() {
		try {
			audit = await schedulerApi.audit(selectedDate);
		} catch (e) {
			error = message(e);
		}
	}
	function zoneLabel(zone: string) {
		return zones.find(([key]) => key === zone)?.[1] || '尚未安排';
	}
	onMount(() => {
		const requestedDate = new URLSearchParams(window.location.search).get('date');
		if (
			requestedDate &&
			/^\d{4}-\d{2}-\d{2}$/.test(requestedDate) &&
			calendarDate(new Date(`${requestedDate}T12:00:00`)) === requestedDate
		) {
			selectedDate = requestedDate;
			month = requestedDate.slice(0, 7);
			calendar?.setDate(requestedDate, false);
		}
		if (GuardAndContinue(IsOperator)) void load();
	});
</script>

<svelte:head
	><title>{manage ? '管理排班' : hall ? '换班请求大厅' : '我的排班'} · 网维报修</title></svelte:head
>
<h1>{manage ? '管理排班' : hall ? '换班请求大厅' : '我的排班'}</h1>
<p class="intro">
	值班时间 16:30–18:00（北京时间）。{manage
		? '按星期配置小组和片区，从所选日期起沿用一个月。换班与蹭班只影响当天。'
		: '日历中有蓝色标记的日期需要值班。'}
</p>
<div class="toolbar">
	{#if manage || hall || canManage}<Button
			kind="ghost"
			href={manage || hall ? '/op/scheduler/' : '/admin/scheduler/'}
			>{manage || hall ? '查看我的排班' : '排班管理与确认'}</Button
		>{/if}
	{#if !hall}<Button kind="ghost" href="/op/scheduler/swaps/">换班请求大厅</Button>{/if}
	<Button kind="ghost" disabled={loading || busy} on:click={load}>刷新</Button>
</div>
{#if error}<InlineNotification kind="error" title="操作未完成" subtitle={error} hideCloseButton />
{/if}
{#if notice}<InlineNotification
		kind={noticeKind}
		title={noticeKind === 'info' ? '排班预览' : '操作完成'}
		subtitle={notice}
		hideCloseButton
	/>{/if}
{#if loading}<InlineLoading description="正在读取排班…" />{/if}

<div class="calendar">
	<DatePicker
		datePickerType="single"
		dateFormat="Y-m-d"
		locale={Mandarin}
		value={selectedDate}
		flatpickrProps={calendarOptions}
		on:change={selected}
	>
		<DatePickerInput id="scheduler-calendar" labelText="值班日期" hideLabel />
	</DatePicker>
</div>
{#if overview}
	<p class="caption">
		{overview.me.name}（{overview.me.wid}）· {schedulerRoles[overview.me.access]} · 本月 {myDays.length}
		次值班
	</p>
	{#if !manage && !hall}
		<Accordion>
			<AccordionItem title="我的空闲时间">
				<p>每周固定空闲日；未勾选代表尚未填报有空。</p>
				<div class="weekdays">
					{#each weekLabels as label, i}<Checkbox
							labelText={label}
							checked={freeDays.includes(i + 1)}
							disabled={busy}
							on:check={(event) => toggleFree(i + 1, event.detail)}
						/>{/each}
				</div>
				<Select labelText={`${selectedDate} 的临时安排`} bind:selected={exception} disabled={busy}>
					<SelectItem value="" text="沿用每周空闲设置" />
					<SelectItem value="true" text="当天有空" /><SelectItem
						value="false"
						text="当天请假／不可用"
					/>
				</Select>
				<Button disabled={busy} on:click={saveAvailability}>保存空闲时间</Button>
				<p class="caption">修改空闲时间不会取消已发布的班次。需要调整时，请申请换班或联系组长。</p>
			</AccordionItem>
			<AccordionItem title="本月我的值班">
				{#each myDays as day}
					<Button
						kind="ghost"
						on:click={() => {
							selectedDate = day.date;
							calendar?.setDate(day.date, true);
						}}
						>{day.date} · {zoneLabel(
							day.members.find((m) => m.wid === overview?.me.wid)?.zone || ''
						)}</Button
					>
				{:else}<p>本月暂没有已发布的值班安排。</p>{/each}
			</AccordionItem>
		</Accordion>
	{/if}

	{#if !hall}<Tile>
			<h2>{selectedDate} 的排班</h2>
			{#if selectedDay}
				<p>
					值班组长：{selectedDay.leader.name}（{selectedDay.leader.wid}） <Tag
						type={selectedDay.published ? 'green' : 'gray'}
						>{selectedDay.published ? '已发布' : '草稿'}</Tag
					>
				</p>
				<div class="legend" aria-label="成员颜色说明">
					{#each [['组长', '#ed7d31'], ['API / dev', '#9dc3e6'], ['片区负责人', '#ffd966'], ['正式成员', '#92d050'], ['实习成员', '#fff'], ['科长', '#0070c0']] as [label, color]}
						<span style:background-color={color} style:color={label === '科长' ? '#fff' : '#161616'}
							>{label}</span
						>
					{/each}
				</div>
				<div class="table-scroll">
					<Table size="compact"
						><TableHead
							><TableRow
								><TableHeader>成员</TableHeader><TableHeader>片区</TableHeader><TableHeader
									>类型</TableHeader
								></TableRow
							></TableHead
						><TableBody>
							{#each selectedDay.members as member}
								<TableRow
									><TableCell
										><span
											class="member"
											style:background-color={memberColor(member, selectedDay.leader.wid)}
											style:color={member.access === 'chief' ? '#fff' : '#161616'}
											>{member.name}（{member.wid}）</span
										></TableCell
									><TableCell>{zoneLabel(member.zone)}</TableCell><TableCell
										>{schedulerRoles[member.access]}{member.area_leader
											? ' · 片区负责人'
											: ''}{member.extra ? ' · 蹭班' : ''}</TableCell
									></TableRow
								>
							{/each}
						</TableBody></Table
					>
				</div>
			{:else}<p>
					{overview.manager
						? '当天尚未配置小组。可在下面选择组长和成员。'
						: '当天暂无可见的排班安排。'}
				</p>{/if}
		</Tile>

		{#if manage && (overview.manager || canArrange)}
			{#if overview.manager}
				<Tile
					><h2>配置每周小组</h2>
					<p>
						从 {selectedDate} 起一个月内，每个{weekdayLabel}沿用这些组长和成员。单次换班或蹭班日期的修改仅影响当天。
					</p>
					{#if staffLoadedDate !== selectedDate}
						<InlineLoading description="正在读取候选成员…" />
					{:else}
						<Select labelText="值班组长" bind:selected={leaderWid} disabled={busy}>
							<SelectItem value="" text="请选择组长" />
							{#each staff.filter((p) => p.access === 'group-leader') as p}<SelectItem
									value={p.wid}
									text={`${p.name}（${p.wid}）`}
								/>{/each}
						</Select>
						<MultiSelect
							labelText="小组成员"
							label="请选择小组成员"
							items={groupItems}
							bind:selectedIds={selectedWids}
							disabled={busy}
						/>
						<p class="caption">值班组长会自动加入小组。调整小组后需重新安排并发布。</p>
						<Button
							disabled={busy || !leaderWid}
							on:click={() => change('group', { leader_wid: leaderWid, wids: selectedWids })}
							>保存每周小组</Button
						>
					{/if}
				</Tile>
			{/if}
			{#if selectedDay && canArrange}
				<Tile
					><h2>安排成员片区</h2>
					<div class="toolbar">
						<Button kind="secondary" disabled={busy} on:click={() => suggest('reuse')}
							>沿用上次排班</Button
						><Button kind="tertiary" disabled={busy} on:click={() => suggest('random')}
							>随机安排实习成员</Button
						>
					</div>
					<p class="caption">
						随机安排保持上次各片区总人数，正式成员等固定人员需先选好片区并保存。两种方式都需检查并保存。
					</p>
					{#each assignments as assignment, i (assignment.wid)}
						{@const member = selectedDay.members.find((m) => m.wid === assignment.wid)!}
						<div
							class="assignment"
							style:background-color={memberColor(
								{ ...member, area_leader: assignment.area_leader },
								selectedDay.leader.wid
							)}
							style:color={member.access === 'chief' ? '#fff' : '#161616'}
						>
							<p>
								<strong>{member.name}（{member.wid}）</strong> · {schedulerRoles[member.access]}
							</p>
							{#if !member.available}<Tag type="warm-gray">未填报有空或当天不可用</Tag>{/if}
							<Select
								labelText={`${member.name}所属片区`}
								bind:selected={assignments[i].zone}
								disabled={busy}
							>
								<SelectItem value="" text="请选择片区" />{#each zones as [key, label]}<SelectItem
										value={key}
										text={label}
									/>{/each}
							</Select>
							<Checkbox
								labelText="片区负责人"
								bind:checked={assignments[i].area_leader}
								disabled={busy}
							/>
						</div>
					{/each}
					<Button
						disabled={busy || assignments.some((a) => !a.zone)}
						on:click={() => change('assignments', { assignments })}>保存片区安排</Button
					>
					<Checkbox
						labelText="已确认不可用成员的特殊值班安排"
						bind:checked={overrideUnavailable}
						disabled={busy}
					/>
					<Button
						disabled={busy || selectedDay.published}
						on:click={() => change('publish', { override_unavailable: overrideUnavailable })}
						>发布排班</Button
					>
				</Tile>
			{/if}
			<TextArea
				labelText="调整／特殊安排原因"
				bind:value={reason}
				maxCount={1000}
				placeholder="修改已发布排班或确认特殊安排时必填"
				disabled={busy}
			/>
		{/if}

		{#if selectedDay && canArrange}
			<Tile
				><h2>添加蹭班</h2>
				<TextInput labelText="成员工号" bind:value={extraWid} disabled={busy} />
				<Select labelText="蹭班片区" bind:selected={extraZone} disabled={busy}
					><SelectItem value="" text="请选择片区" />{#each zones as [key, label]}<SelectItem
							value={key}
							text={label}
						/>{/each}</Select
				>
				{#if !manage}<TextArea
						labelText="调整原因（已发布班次必填）"
						bind:value={reason}
						maxCount={1000}
					/>{/if}
				<Button
					disabled={busy || !extraWid.trim() || !extraZone}
					on:click={() => change('extra', { wid: extraWid.trim(), zone: extraZone })}
					>添加到当天班次</Button
				>
				{#each selectedDay.members.filter((m) => m.extra) as member}<div class="toolbar">
						<span>{member.name}（{member.wid}）· {zoneLabel(member.zone)}</span><Button
							kind="danger-ghost"
							size="small"
							disabled={busy}
							on:click={() => change('remove_extra', { wid: member.wid })}>移除蹭班</Button
						>
					</div>{/each}
			</Tile>
		{/if}
	{/if}
	{#if hall}
		<Tile
			><h2>申请一次性换班</h2>
			<p>
				提交自己的班次和预期换班日期，由预期日期的组员接单，再交双方组长确认。换班只影响这一次班次。值班组长的更换需由科长/dev/api
				配置。
			</p>
			<Select labelText="我的原值班日期" bind:selected={swapDateA} disabled={busy}>
				<SelectItem value="" text="请选择自己的班次" />
				{#each myDays.filter((d) => d.date >= chinaToday() && d.leader.wid !== overview?.me.wid && d.members.some((m) => m.wid === overview?.me.wid && !m.extra)) as day}<SelectItem
						value={day.date}
						text={`${day.date} · ${zoneLabel(day.members.find((m) => m.wid === overview?.me.wid)?.zone || '')}`}
					/>{/each}
			</Select>
			<DatePicker
				datePickerType="single"
				dateFormat="Y-m-d"
				locale={Mandarin}
				bind:value={swapDateB}
				><DatePickerInput
					id="swap-target-date"
					labelText="预期换班日期"
					placeholder="YYYY-MM-DD"
				/></DatePicker
			>
			<TextArea labelText="换班说明" bind:value={swapReason} maxCount={1000} disabled={busy} />
			<Button
				disabled={busy || !swapDateA || !swapDateB || swapDateA === swapDateB}
				on:click={() =>
					perform(
						() =>
							schedulerApi.requestSwap({
								date_a: swapDateA,
								date_b: swapDateB,
								reason: swapReason
							}),
						'换班请求已发布，等待预期日期的组员接单'
					)}>提交换班申请</Button
			>
		</Tile>
	{/if}
	{#if hall || manage}<section class="swap-requests">
			<h2>{hall ? '本月换班请求大厅' : '本月换班申请'}</h2>
			{#if hall}<p>
					展示你参与的请求和你在预期日期值班时可查看的待接单请求。接单后还需双方组长确认。
				</p>{/if}
			{#each swaps as request}
				<Tile>
					<p>
						<strong>{request.requester_name}（{request.requester_wid}）</strong>
						原班次 {request.date_a} → 预期日期 {request.date_b}
						{#if request.target_wid}<br />接单成员：<strong
								>{request.target_name}（{request.target_wid}）</strong
							>{/if}
					</p>
					<Tag type={request.status === 'approved' ? 'green' : 'gray'}
						>{statuses[request.status] || request.status}</Tag
					>
					{#if request.status !== 'open'}<p>
							原日期组长：{request.approval_a ? '已确认' : '未确认'} · 对方日期组长：{request.approval_b
								? '已确认'
								: '未确认'}
						</p>{/if}
					{#if request.reason}<p>{request.reason}</p>{/if}
					{#if request.status === 'open'}
						{#if request.requester_wid === overview.me.wid}
							<p class="caption">等待预期日期的其他组员接单。</p>
						{:else}
							<Button
								disabled={busy || !request.can_claim}
								on:click={() => review(request, 'claim')}>接单</Button
							>
							{#if !request.can_claim}<p class="caption">
									接单需在预期日期的已发布班次中值班，且不是当天组长；双方不能在交换后的日期重复值班，班次须有效且未过期。
								</p>{/if}
						{/if}
					{/if}
					{#if request.can_review}<div class="toolbar">
							<Button disabled={busy} on:click={() => review(request, 'approve')}>确认换班</Button
							><Button
								kind="danger-tertiary"
								disabled={busy}
								on:click={() => review(request, 'reject')}>拒绝</Button
							>
						</div>{/if}
					{#if (request.status === 'pending' || request.status === 'open') && request.requester_wid === overview.me.wid}<Button
							kind="ghost"
							disabled={busy}
							on:click={() => review(request, 'cancel')}>撤销申请</Button
						>{/if}
				</Tile>
			{:else}<p>本月暂无与你相关的换班申请。</p>{/each}
		</section>{/if}
	{#if manage && selectedDay && canArrange}
		<Accordion
			><AccordionItem title="当天操作记录"
				><Button kind="ghost" on:click={showAudit}>读取记录</Button>
				{#each audit as entry}<p>
						{new Date(entry.created_at).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })} · {entry.actor_wid}
						· {entry.action}<br />{entry.reason}
					</p>{/each}
			</AccordionItem></Accordion
		>
	{/if}
{/if}

<style>
	.calendar :global(.bx--date-picker__icon) {
		display: none;
	}
	.calendar :global(.bx--date-picker__input) {
		display: none;
	}
	h1 {
		font-size: 1.75rem;
		margin-bottom: 1rem;
	}
	h2 {
		font-size: 1.125rem;
		margin-bottom: 0.75rem;
	}
	p {
		line-height: 1.6;
		margin: 0.5rem 0;
	}
	.intro,
	.caption {
		color: var(--cds-text-secondary, #525252);
		font-size: 0.875rem;
	}
	.toolbar,
	.weekdays,
	.legend {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		align-items: center;
		margin: 0.75rem 0;
	}
	.calendar {
		margin: 1rem 0;
		min-height: 330px;
	}
	.legend span {
		padding: 0.25rem 0.5rem;
		border: 1px solid #8d8d8d;
		font-size: 0.75rem;
	}
	.member {
		display: inline-block;
		padding: 0.4rem 0.5rem;
		border: 1px solid #8d8d8d;
	}
	.table-scroll {
		overflow-x: auto;
	}
	.table-scroll :global(table) {
		min-width: 420px;
	}
	.assignment {
		border: 1px solid #8d8d8d;
		padding: 0.75rem;
		margin: 0.75rem 0;
	}
	.swap-requests p {
		overflow-wrap: anywhere;
	}
	:global(.scheduler-duty-date) {
		box-shadow: inset 0 -4px 0 #0f62fe !important;
		font-weight: 600 !important;
	}
	:global(.scheduler-duty-date:not(.selected)) {
		background-color: #edf5ff !important;
	}
	:global(.bx--tile) {
		margin: 1rem 0;
	}
</style>
