<script lang="ts">
	import { CheckAndGetJWT, Guard } from '$lib/jwt';
	import type { NewTicketReq } from '$lib/types/apiRequest';
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
	import { RFC3339 } from '$lib/types/RFC3339';
	import { onMount } from 'svelte';
	import { IsUser } from '$lib/types/enum';

	import {
		DatePicker,
		DatePickerInput,
		RadioButtonGroup,
		RadioButton,
		TextArea,
		Button,
		NotificationQueue,
		Loading
	} from 'carbon-components-svelte';
	import { IsRFC3339 } from '$lib/types/RFC3339';
	import { invalidState } from '$lib/types/invalidState.svelte';
	import { NewTicket } from '$lib/api';
	import { goto } from '$app/navigation';

	let notLoading: boolean = $state(true);

	let q: NotificationQueue;

	let r = $state({} as NewTicketReq);

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
		let issuerSID = CheckAndGetJWT('parsed').sid;
		r.issuer_sid = issuerSID;
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

	onMount(() => Guard(IsUser));
</script>

<h1>提交新报修</h1>
<br />
<hr />
<br />
<p>
	<i
		>请仔细填写这张报修表，在成功提交后，会有网维的工作人员在您预约的时间通过电话联系您或上门维修您的问题。</i
	>
</p>
<br />

<DatePicker datePickerType="single" on:change={onOccurDateChange}>
	<DatePickerInput
		labelText="故障是在什么时候发生的？"
		placeholder="记不清楚可不填"
		invalid={occurAt.notOK}
		invalidText={occurAt.txt}
	/>
</DatePicker>
<br />
<br />
<RadioButtonGroup
	legendText="故障大概是什么问题？（准确填写有助于我们维修）"
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
	placeholder="请告诉我们你遇到了什么网络问题，越详细越好~"
	bind:value={r.description}
	invalid={description.notOK}
	invalidText={description.txt}
/>
<br />
<br />
<DatePicker datePickerType="single" on:change={onAppointDateChange}>
	<DatePickerInput
		labelText="预约我们上门维修的日期"
		placeholder="当天4:30~6:00您本人需要在宿舍"
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
	placeholder="其它您需要告诉我们的事情，没有可不填"
	bind:value={r.notes}
	invalid={notes.notOK}
	invalidText={notes.txt}
/>
<br />
<br />
<Button on:click={handleSubmit}>提交</Button>

<NotificationQueue bind:this={q} />

<Loading active={!notLoading} />
