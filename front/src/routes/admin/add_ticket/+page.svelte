<script lang="ts">
	import { CheckAndGetJWT, Guard } from '$lib/jwt';
	import type { NewTicketReq } from '$lib/types/apiRequest';
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
	import { RFC3339 } from '$lib/types/RFC3339';
	import { onMount } from 'svelte';
	import { IsAdmin, IsUser } from '$lib/types/enum';

	import {
		DatePicker,
		DatePickerInput,
		RadioButtonGroup,
		RadioButton,
		TextArea,
		Button,
		NotificationQueue,
		Loading,
		TextInput,
		Select,
		SelectItem
	} from 'carbon-components-svelte';
	import { IsRFC3339 } from '$lib/types/RFC3339';
	import { invalidState } from '$lib/types/invalidState.svelte';
	import { NewTicket } from '$lib/api';
	import { goto } from '$app/navigation';

	let notLoading: boolean = $state(true);

	let q: NotificationQueue;

	let r = $state({
		priority: 'normal'
	} as NewTicketReq);

	function onOccurDateChange(event: CustomEvent) {
		const { dateStr } = event.detail;
		if (dateStr) {
			r.occur_at = RFC3339(dateStr);
		}
	}

	function onAppointDateChange(event: CustomEvent) {
		const { dateStr } = event.detail;
		if (dateStr) {
			const date = new Date(dateStr);
			date.setHours(16, 30, 0, 0); // Set time to 16:30:00
			r.appointed_at = RFC3339(date);
		}
	}

	function handleSubmit() {
		console.log('提交的表单数据:', r);
		check() ? submit() : jumpInvalid();
	}

	let occurAt = new invalidState();
	let appointedAt = new invalidState();
	let description = new invalidState();
	let notes = new invalidState();

	function check(): boolean {
		notLoading = false;
		let ok = false;
		occurAt.reset();
		appointedAt.reset();
		description.reset();
		notes.reset();

		occurAt.assert(!r.occur_at || IsRFC3339(r.occur_at), '请输入正确的故障发生时间');
		appointedAt.assert(!r.appointed_at || IsRFC3339(r.appointed_at), '请输入正确的预约时间');
		description.assert(r.description && r.description.length > 0, '请填写故障描述');
		description.assert(r.description.length <= 100, '字数太多了，请控制在100字以内');
		notes.assert(!r.notes || r.notes.length <= 100, '字数太多了...请控制在100字以内');
		if (r.category == undefined) {
			r.category = 'others';
		}

		if (!r.occur_at) {
			r.occur_at = undefined;
		}

		if (!r.appointed_at) {
			r.appointed_at = undefined; //防止序列化问题
		}

		notLoading = true;
		if (occurAt.notOK || appointedAt.notOK || description.notOK || notes.notOK) {
			ok = false;
		} else {
			ok = true;
		}
		return ok;
	}

	async function submit() {
		try {
			notLoading = false;
			let res = await NewTicket(r);
			notLoading = true;
			if (!res.success) {
				throw new Error(res.msg || '提交失败.........');
			}
			q.add({
				kind: 'success',
				title: '提交成功',
				timeout: 3000
			});
			setTimeout(() => goto('/repair'), 3900);
		} catch (e: any) {
			notLoading = true;
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '提交失败',
				subtitle: errMsg + '，请重试',
				timeout: 5000
			});
		}
	}

	function jumpInvalid() {
		if (occurAt.notOK) {
			document.getElementById('occur_at')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (appointedAt.notOK) {
			document
				.getElementById('appointed_at')
				?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (description.notOK) {
			document
				.getElementById('description')
				?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (notes.notOK) {
			document.getElementById('notes')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
	}

	onMount(() => Guard(IsAdmin));
</script>

<h1>增添工单</h1>
<br />
<hr />
<br />
<p>为他人增添报修，注意，首先需要知道他人的学号。单独的工单增添正在开发中...</p>
<br />

<TextInput labelText="用户的学号" placeholder="请输入用户学号" bind:value={r.issuer_sid} />
<br />
<br />

<DatePicker datePickerType="single" on:change={onOccurDateChange}>
	<DatePickerInput
		labelText="故障发生的日期"
		placeholder="记不清楚可不填"
		invalid={occurAt.notOK}
		invalidText={occurAt.txt}
	/>
</DatePicker>
<br />
<br />
<RadioButtonGroup
	legendText="故障问题的类型"
	orientation="vertical"
	bind:selected={r.category}
	required={true}
>
	<RadioButton labelText="需要新安装宽带" value="first-install" />
	<RadioButton labelText="IP地址或者网络设备问题" value="ip-or-device" />
	<RadioButton labelText="电脑软件或者账号的问题" value="client-or-account" />
	<RadioButton labelText="网速问题" value="low-speed" />
	<RadioButton labelText="其它问题/不清楚" value="others" />
</RadioButtonGroup>
<br />
<br />
<TextArea
	labelText="故障描述"
	placeholder="描述一下故障的情况"
	bind:value={r.description}
	invalid={description.notOK}
	invalidText={description.txt}
/>
<br />
<br />
<DatePicker datePickerType="single" on:change={onAppointDateChange}>
	<DatePickerInput
		labelText="预约上门维修的日期"
		placeholder="当天4:30~6:00用户需要在宿舍"
		invalid={appointedAt.notOK}
		invalidText={appointedAt.txt}
	/>
</DatePicker>
<br />
<br />
<hr />
<br />
<br />
<TextArea
	labelText="备注"
	placeholder="其它对维修成员有用的信息，注意事项等。"
	bind:value={r.notes}
	invalid={notes.notOK}
	invalidText={notes.txt}
/>
<br />
<br />
<Select
	labelText="工单优先级"
	bind:selected={r.priority}
	helperText="选择工单的优先级类型，更高优先级会在系统中优先显示"
>
	<SelectItem value="highest" text="十万火急" style="color: var(--color-red-600);" />
	<SelectItem value="assigned" text="运营商工单" style="color: #2563eb;" />
	<SelectItem value="mainline" text="主线任务" />
	<SelectItem value="normal" text="一般报修" />
	<SelectItem value="in-passing" text="顺路看看" />
	<SelectItem value="least" text="不紧急" />
</Select>
<br />
<br />
<Button on:click={handleSubmit}>提交</Button>

<NotificationQueue bind:this={q} />

<Loading active={!notLoading} />
