<script lang="ts">
	import { CheckAndGetJWT, Guard } from '$lib/jwt';
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
		ComposedModal,
		ModalBody,
		ModalFooter,
		ModalHeader,
		NotificationQueue,
		Loading
	} from 'carbon-components-svelte';
	import { IsRFC3339 } from '$lib/types/RFC3339';
	import { invalidState } from '$lib/types/invalidState.svelte';
	import { NewTicket, GetSubscribeConfig } from '$lib/api';
	import { goto } from '$app/navigation';
	import WxOpenSubscribe from '$lib/components/WxOpenSubscribe.svelte';
	import {
		hasKeepSubscribeChoice,
		markKeepSubscribeChoice
	} from '$lib/states/wechatSubscribeStatus';

	let notLoading: boolean = $state(true);

	let q: NotificationQueue;

	let r = $state({} as NewTicketReq);
	let subscribeTemplateId = $state('');

	// 订阅提示模态框状态
	let subscribeModalOpen = $state(false);
	// WxOpenSubscribe 覆盖层就绪状态
	let subscribeReady = $state(false);

	// 是否处于微信内且可发起订阅授权
	const canAskSubscribe = $derived(subscribeTemplateId !== '');

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
		if (!check()) {
			jumpInvalid();
			return;
		}
		// 微信内且可订阅、且用户尚未点过「好」（未记住选择）时：先弹提示模态框。
		// 模态框「好」按钮上叠有透明开放标签，用户的真实点击会直接触发微信授权。
		if (canAskSubscribe && !hasKeepSubscribeChoice()) {
			subscribeModalOpen = true;
			return;
		}
		// 已记住选择 / 非微信 / 无模板：直接提交。
		// （已记住选择意味着用户之前在微信授权界面勾选了「总是保持上述选择」，
		//  后续由微信侧记住授权，这里无需也无法再次无感触发——合成点击会被微信拦截。）
		proceedSubmit();
	}

	// 微信授权已出结果（成功/拒绝/失败都算）：记录「已做出选择」、关闭模态框并继续提交报修。
	// 该回调由叠在「好」按钮上的透明开放标签被真实点击后触发。
	function onSubscribeSettled() {
		markKeepSubscribeChoice();
		subscribeModalOpen = false;
		proceedSubmit();
	}

	// 模态框「好」按钮自身的点击（与透明开放标签的授权触发并发）：仅作兜底。
	// 正常情况下授权结果回调 onSubscribeSettled 会负责后续；若开放标签未就绪/非微信
	// 环境导致授权未触发（无 success/error 回调），这里保证模态框能关闭并继续提交。
	function onModalConfirm() {
		if (!canAskSubscribe || !subscribeReady) {
			markKeepSubscribeChoice();
			subscribeModalOpen = false;
			proceedSubmit();
		}
		// 已就绪时：授权由真实点击开放标签触发，等待 onSubscribeSettled 回调，这里不重复提交。
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
		description.assert(r.description.length <= 200, '字数太多了，请控制在200字以内');
		notes.assert(!r.notes || r.notes.length <= 200, '字数太多了...请控制在200字以内');
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

	async function proceedSubmit() {
		let issuerSID = CheckAndGetJWT('parsed')?.sid;
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

	async function fetchSubscribeConfig() {
		//确认用户是否在微信中打开网页
		const ua = navigator.userAgent.toLowerCase();
		if (!ua.includes('micromessenger')) return;

		try {
			const cfg = await GetSubscribeConfig();
			if (cfg.success && cfg.template_id) {
				subscribeTemplateId = cfg.template_id;
			}
		} catch (e: any) {
			q.add({
				kind: 'warning',
				title: '获取订阅配置失败',
				subtitle: e.response?.data?.msg || e.message || '未知错误',
				timeout: 3000
			});
		}
	}

	onMount(() => (Guard(IsUser), fetchSubscribeConfig()));
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
	labelText="备注"
	placeholder="其它您需要告诉我们的事情，没有可不填"
	bind:value={r.notes}
	invalid={notes.notOK}
	invalidText={notes.txt}
/>
<br />
<br />
<p style="color: gray; font-style: italic;">
	如果报修时有任何疑问，请加入QQ群：{SUPPORT_QQ} 询问与反馈，我们会热情地解答您的问题。
</p>
<br />
<Button on:click={handleSubmit}>提交</Button>

<!-- 订阅提示模态框：交互按钮（好/返回）与微信授权触发器分离 -->
<ComposedModal
	bind:open={subscribeModalOpen}
	on:close={() => {
		subscribeModalOpen = false;
	}}
	class="mobile-floating-modal"
	preventCloseOnClickOutside
>
	<ModalHeader title="微信提醒报修进度授权" />
	<ModalBody hasForm>
		<p>
			当您的报修提交后，工单状态更新时，我们可以通过微信通知您最新进展。
		</p>
		<br />
		<p>
			如果您需要，点击“好”后会弹出微信授权界面，请在其中勾选“总是保持上述选择”并同意我们的推送。
		</p>
	</ModalBody>
	<ModalFooter>
		<Button
			kind="secondary"
			on:click={() => {
				subscribeModalOpen = false;
			}}>返回</Button
		>
		<!--
			「好」按钮区域：相对定位容器内，可见的 Carbon 按钮在下，
			透明 wx-open-subscribe 开放标签绝对定位铺满叠在上。用户的真实点击
			落在透明标签上 → 微信原生授权被真实手势触发；按钮自身 on:click 作兜底。
		-->
		<span style="position: relative; display: inline-block;">
			<Button kind="primary" on:click={onModalConfirm}>好</Button>
			{#if canAskSubscribe}
				<WxOpenSubscribe
					templateId={subscribeTemplateId}
					scene={0}
					onReady={() => (subscribeReady = true)}
					onSuccess={onSubscribeSettled}
					onError={onSubscribeSettled}
				/>
			{/if}
		</span>
	</ModalFooter>
</ComposedModal>

<NotificationQueue bind:this={q} />

<Loading active={!notLoading} />

<style>
	:global(.mobile-floating-modal.bx--modal) {
		@media (max-width: 672px) {
			display: flex !important;
			align-items: center !important;
			justify-content: center !important;
			background-color: rgba(22, 22, 22, 0.5) !important;
		}
	}

	:global(.mobile-floating-modal .bx--modal-container) {
		@media (max-width: 672px) {
			width: 90% !important;
			max-width: 400px !important;
			height: auto !important;
			max-height: 85vh !important;
			position: relative !important;
			margin: 0 !important;
			top: auto !important;
			left: auto !important;
			transform: none !important;
			box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4) !important;
		}
	}

	:global(.mobile-floating-modal .bx--modal-content) {
		@media (max-width: 672px) {
			max-height: 60vh !important;
			overflow-y: auto !important;
			margin-bottom: 0 !important;
		}
	}
</style>
