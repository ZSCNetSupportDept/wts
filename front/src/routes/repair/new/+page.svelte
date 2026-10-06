<script lang="ts">
	import { CheckAndGetJWT, GuardAndContinue } from '$lib/jwt';
	import type { NewTicketReq } from '$lib/types/apiRequest';
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
	import { RFC3339 } from '$lib/types/RFC3339';
	import { onMount } from 'svelte';
	import { IsUser } from '$lib/types/enum';
	import { SUPPORT_QQ } from '$lib/env/businesses';

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
	import { NewTicket, GetSubscribeConfig } from '$lib/api';
	import { goto } from '$app/navigation';
	import WxOpenSubscribe from '$lib/components/WxOpenSubscribe.svelte';

	let notLoading: boolean = $state(true);

	let q: NotificationQueue;

	let r = $state({} as NewTicketReq);
	let subscribeTemplateId = $state('');
	let showValidation = $state(false);
	let submitting = $state(false);
	let subscribeUnavailable = $state(false);
	let subscribeConfigLoading = $state(true);

	type TicketValidation = {
		occurAt?: string;
		appointedAt?: string;
		description?: string;
		notes?: string;
	};

	function onOccurDateChange(event: CustomEvent) {
		const { dateStr } = event.detail;
		// Carbon DatePicker 初始化日历时会对空值派发一次 change。
		// 这不是用户输入，不能因此显示其它字段的校验错误。
		if (!dateStr) {
			r.occur_at = undefined;
			return;
		}
		showValidation = true;
		r.occur_at = RFC3339(dateStr);
	}

	function onAppointDateChange(event: CustomEvent) {
		const { dateStr } = event.detail;
		if (!dateStr) {
			r.appointed_at = undefined;
			return;
		}
		showValidation = true;
		const date = new Date(dateStr);
		date.setHours(16, 30, 0, 0); // Set time to 16:30:00
		r.appointed_at = RFC3339(date);
	}

	function onCategoryChange(event: CustomEvent<unknown>) {
		// Carbon RadioButtonGroup 初始化订阅 store 时会立即用初始值派发一次 change，
		// 实测该值是 null（而不是 undefined），所以原先 `!== undefined` 的判断拦不住它，
		// 会把 showValidation 提前置为 true、一进页面就显示“请填写故障描述”。
		// 这不是用户输入，直接忽略。
		if (event.detail === undefined || event.detail === null) {
			return;
		}
		showValidation = true;
	}

	function handleSubmit() {
		showValidation = true;
		if (!isValid) {
			jumpInvalid();
			return;
		}
		void submit();
	}

	let occurAt = new invalidState();
	let appointedAt = new invalidState();
	let description = new invalidState();
	let notes = new invalidState();

	function validateTicketDraft(): TicketValidation {
		const errors: TicketValidation = {};
		if (r.occur_at && !IsRFC3339(r.occur_at)) {
			errors.occurAt = '请输入正确的故障发生时间';
		}
		if (r.appointed_at && !IsRFC3339(r.appointed_at)) {
			errors.appointedAt = '请输入正确的预约时间';
		}
		if (!r.description?.trim()) {
			errors.description = '请填写故障描述';
		} else if (r.description.length > 200) {
			errors.description = '字数太多了，请控制在200字以内';
		}
		if (r.notes && r.notes.length > 200) {
			errors.notes = '字数太多了...请控制在200字以内';
		}
		return errors;
	}

	let validation = $derived(validateTicketDraft());
	let isValid = $derived(Object.keys(validation).length === 0);

	// 只写不读：不要在 $effect 里调用 assert() 来更新这些展示状态，
	// 因为 assert() 内部会读取 notOK，而 effect 又在写 notOK，
	// 会让 effect 依赖自己写入的状态、无限更新（effect_update_depth_exceeded）。
	function syncInvalid(state: invalidState, shouldShow: boolean, msg?: string) {
		if (shouldShow && msg) {
			state.notOK = true;
			state.txt = msg;
		} else {
			state.notOK = false;
			state.txt = '';
		}
	}

	$effect(() => {
		const errors = validation;
		const shouldShow = showValidation;
		syncInvalid(occurAt, shouldShow, errors.occurAt);
		syncInvalid(appointedAt, shouldShow, errors.appointedAt);
		syncInvalid(description, shouldShow, errors.description);
		syncInvalid(notes, shouldShow, errors.notes);
	});

	async function submit() {
		if (submitting || !isValid) return;
		submitting = true;
		let created = false;
		const issuerSID = CheckAndGetJWT('parsed')?.sid;
		const request: NewTicketReq = {
			...r,
			issuer_sid: issuerSID || '',
			category: r.category || 'others',
			occur_at: r.occur_at || undefined,
			appointed_at: r.appointed_at || undefined,
			description: r.description.trim(),
			notes: r.notes || undefined
		};
		try {
			notLoading = false;
			const res = await NewTicket(request);
			notLoading = true;
			if (!res.success) {
				throw new Error(res.msg || '提交失败.........');
			}
			created = true;
			q.add({
				kind: 'success',
				title: '提交成功',
				timeout: 1000
			});
			setTimeout(() => goto('/repair'), 1500);
		} catch (e: any) {
			notLoading = true;
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '提交失败',
				subtitle: errMsg,
				timeout: 5000
			});
		} finally {
			if (!created) {
				submitting = false;
			}
		}
	}

	function jumpInvalid() {
		if (validation.occurAt) {
			document.getElementById('occur_at')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (validation.description) {
			document
				.getElementById('description')
				?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (validation.appointedAt) {
			document
				.getElementById('appointed_at')
				?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (validation.notes) {
			document.getElementById('notes')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
	}

	async function fetchSubscribeConfig() {
		//确认用户是否在微信中打开网页
		const ua = navigator.userAgent.toLowerCase();
		if (!ua.includes('micromessenger')) {
			subscribeUnavailable = true;
			subscribeConfigLoading = false;
			return;
		}

		try {
			const cfg = await GetSubscribeConfig();
			if (cfg.success && cfg.template_id) {
				subscribeTemplateId = cfg.template_id;
			} else {
				subscribeUnavailable = true;
			}
		} catch (e: any) {
			subscribeUnavailable = true;
			q.add({
				kind: 'warning',
				title: '获取订阅配置失败',
				subtitle: e.response?.data?.msg || e.message || '未知错误',
				timeout: 3000
			});
		} finally {
			subscribeConfigLoading = false;
		}
	}

	onMount(() => {
		if (GuardAndContinue(IsUser)) {
			void fetchSubscribeConfig();
		}
	});

	function onOpenSubscribeUnavailable() {
		subscribeUnavailable = true;
	}
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
		id="occur_at"
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
	on:change={onCategoryChange}
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
	id="description"
	labelText="故障描述"
	placeholder="请告诉我们你遇到了什么网络问题，越详细越好~"
	bind:value={r.description}
	on:input={() => (showValidation = true)}
	invalid={description.notOK}
	invalidText={description.txt}
/>
<br />
<br />
<DatePicker datePickerType="single" on:change={onAppointDateChange}>
	<DatePickerInput
		id="appointed_at"
		labelText="预约我们上门维修的日期"
		placeholder="当天下午4:30~6:00您需要在宿舍"
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
	id="notes"
	labelText="备注"
	placeholder="其它您需要告诉我们的事情，没有可不填"
	bind:value={r.notes}
	on:input={() => (showValidation = true)}
	invalid={notes.notOK}
	invalidText={notes.txt}
/>
<br />
<br />
<p style="color: gray; font-style: italic;">
	如果报修时有任何疑问，请加入QQ群：{SUPPORT_QQ} 询问与反馈，我们会热情地解答您的问题。
</p>
<br />
{#if !isValid || subscribeUnavailable}
	<Button disabled={submitting} on:click={handleSubmit}>提交</Button>
{:else if submitting || subscribeConfigLoading || !subscribeTemplateId}
	<Button disabled>提交</Button>
{:else}
	<WxOpenSubscribe
		templateId={subscribeTemplateId}
		label="提交"
		onSuccess={submit}
		onError={submit}
		onUnavailable={onOpenSubscribeUnavailable}
	/>
{/if}

<NotificationQueue bind:this={q} />

<Loading active={!notLoading} />
