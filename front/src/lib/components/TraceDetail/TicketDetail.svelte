<script lang="ts">
	import {
		ComposedModal,
		ModalHeader,
		ModalBody,
		Portal,
		Button,
		NotificationQueue,
		Modal,
		ModalFooter
	} from 'carbon-components-svelte';
	import type { Ticket, Trace } from '$lib/types/apiResponse';
	import { GetTraces, CancelTicket } from '$lib/api';
	import TraceTimeline from './TraceTimeline.svelte';
	import { sampleTrace } from '$lib/testData/ticket';
	import ModalButton from './ModalButton.svelte';
	import UpViewButton from './UpViewButton.svelte';
	import TraceUpdateView from './TraceUpdateView.svelte';
	import { TicketModal } from '$lib/states/ticketDetails.svelte';
	import { DatePicker, DatePickerInput } from 'carbon-components-svelte';
	import { RFC3339 } from '$lib/types/RFC3339';
	import type { ChangeAppointmentRes } from '$lib/types/apiResponse';
	import type { ChangeAppointmentReq } from '$lib/types/apiRequest';
	import { ChangeAppointment } from '$lib/api';
	import { Loading } from 'carbon-components-svelte';

	let view: 'trace' | 'cancel' | 'update' | 'appoint' = $state('trace');

	// TODO：现在不能不传递参数，为了简化使用考虑默认从全局状态获取参数；
	let {
		t = TicketModal.NowTicket,
		open = $bindable(TicketModal.Opened),
		src = TicketModal.SRC,
		onTicketChanged
	}: {
		t?: Ticket | null;
		open?: boolean;
		src?: 'user' | 'operator';
		onTicketChanged?: () => void | Promise<void>;
	} = $props();

	let loading = $state(false);
	let traces: Trace[] = $state([]);

	let q: NotificationQueue = $state<NotificationQueue>(null);
	let q1: NotificationQueue = $state<NotificationQueue>(null);

	// 当模态框打开且有工单ID时，获取记录
	$effect(() => {
		if (open && t != null) {
			fetchTraces(t.tid);
		}
	});

	async function fetchTraces(tid: number) {
		loading = true;
		try {
			const res = await GetTraces(tid.toString());
			traces = res.traces.reverse();
			// traces = sampleTrace; // 使用测试数据
		} catch (e: any) {
			open = false;
			console.error('Failed to load traces', e);
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '加载记录失败',
				subtitle: errMsg,
				timeout: 5000
			});
		} finally {
			loading = false;
		}
	}

	async function cancel(tid: number) {
		try {
			const res = await CancelTicket(tid.toString());
			if (!res.success) {
				throw new Error(res.msg || '取消工单失败');
			}
			q.add({
				kind: 'success',
				title: '工单已取消',
				timeout: 3000
			});
		} catch (e: any) {
			console.error('Failed to cancel ticket', e);
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '取消工单失败',
				subtitle: errMsg,
				timeout: 5000
			});
		} finally {
			open = false;
			view = 'trace';
			await onTicketChanged?.();
		}
	}

	function getTid(t: Ticket | null): string {
		return t ? t.tid.toString() : '...';
	}

	let isUpReady = $state(false);

	let appointment = $state<RFC3339>('' as RFC3339);

	function onAppointDateChange(event: CustomEvent) {
		const { dateStr } = event.detail;
		if (!dateStr) {
			appointment = undefined;
			return;
		}
		const date = new Date(dateStr);
		date.setHours(16, 30, 0, 0); // Set time to 16:30:00
		appointment = RFC3339(date);
	}

	let reappointLock = $state(false);

	async function reappoint() {
		if (reappointLock) return;
		reappointLock = true;
		if (appointment === undefined && appointment === '') {
			q1.add({
				kind: 'error',
				title: '请选择新的预约时间',
				timeout: 5000
			});
			reappointLock = false;
			return;
		}
		try {
			const r: ChangeAppointmentReq = {
				tid: t!.tid,
				new_appointment: appointment
			};
			const res: ChangeAppointmentRes = await ChangeAppointment(r);
			if (!res.success) {
				throw new Error(res.msg || '更改预约失败');
			}
			q.add({
				kind: 'success',
				title: '预约已更改',
				timeout: 3000
			});
		} catch (e: any) {
			console.error('Failed to change appointment', e);
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q1.add({
				kind: 'error',
				title: '更改预约失败',
				subtitle: errMsg,
				timeout: 5000
			});
			return;
		} finally {
			reappointLock = false;
		}
		open = false;
		view = 'trace';
		await onTicketChanged?.();
	}
</script>

<Portal>
	{#if view === 'trace'}
		<ComposedModal size="sm" bind:open class="mobile-floating-modal">
			<ModalHeader title="🗃️No.{getTid(t)}-记录" />
			<ModalBody>
				<TraceTimeline t={traces} {loading} {src} />
			</ModalBody>
			<ModalButton {src} {t} {loading} bind:appointment bind:view bind:open />
		</ComposedModal>
	{/if}
	{#if view === 'cancel'}
		<Modal
			size="sm"
			class="mobile-floating-modal"
			danger
			bind:open
			modalHeading="确定取消工单吗？"
			primaryButtonText="确定取消 "
			secondaryButtonText="算了"
			on:close={() => ((open = false), (view = 'trace'))}
			on:click:button--secondary={() => ((open = false), (view = 'trace'))}
			on:submit={() => cancel(t!.tid)}
		>
			<p>该操作不可逆，若要重新开启工单，您可以在稍后提交一个新的工单。</p>
			<br />
			<br />
		</Modal>
	{/if}
	{#if view === 'update'}
		<!-- 使用moobile-floating-model在这里貌似会显示不好，所以不用了 -->
		<ComposedModal on:close={() => (view = 'trace')} bind:open>
			<NotificationQueue bind:this={q1} />
			<ModalHeader title="  🖋️请更新No.{getTid(t)}的状态" />
			<ModalBody>
				<TraceUpdateView
					{t}
					bind:view
					bind:open
					bind:isUpReady
					{q}
					{q1}
					onUpdated={onTicketChanged}
				/>
			</ModalBody>
			<UpViewButton bind:view bind:isUpReady />
		</ComposedModal>
	{/if}
	{#if view === 'appoint'}
		<ComposedModal on:close={() => (view = 'trace')} bind:open>
			<NotificationQueue bind:this={q1} />
			<ModalHeader title="  🖋️更改No.{getTid(t)}的预约上门时间" />
			<ModalBody>
				<p>
					您可以在此更改本报修的预约时间，我们会努力在您指定日期下午的4：30~6：00上门为您维修网络。
				</p>
				<br />
				<DatePicker datePickerType="single" on:change={onAppointDateChange}>
					<DatePickerInput
						id="appointed_at"
						labelText="预约我们上门维修的日期"
						placeholder="当天下午4:30~6:00您需要在宿舍"
					/>
				</DatePicker>
			</ModalBody>
			<ModalFooter>
				<Button kind="secondary" on:click={() => (view = 'trace')}>返回</Button>
				<Button kind="primary" on:click={() => reappoint()} bind:disabled={reappointLock}
					>提交</Button
				>
			</ModalFooter>
		</ComposedModal>
	{/if}
</Portal>

<NotificationQueue bind:this={q} />

<Loading bind:active={reappointLock} />

<style>
	:global(.mobile-floating-modal.bx--modal) {
		@media (max-width: 672px) {
			display: flex !important;
			align-items: center !important;
			justify-content: center !important;
			/* 确保背景色存在 (Carbon默认有，但为了保险起见) */
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
