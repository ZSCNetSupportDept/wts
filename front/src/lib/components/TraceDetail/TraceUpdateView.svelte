<script lang="ts">
	import {
		Button,
		NotificationQueue,
		Select,
		SelectItem,
		DatePicker,
		DatePickerInput,
		TextArea
	} from 'carbon-components-svelte';
	import type { NewRepairTraceReq } from '$lib/types/apiRequest';
	import type { Ticket } from '$lib/types/apiResponse';
	import { type WtsStatus, type WtsPriority, IsAdmin } from '$lib/types/enum';
	import { FormatDate, RFC3339 } from '$lib/types/RFC3339';
	import { CheckAndGetJWT } from '$lib/jwt';
	import { Category } from 'carbon-icons-svelte';
	import { NewRepairTrace } from '$lib/api';
	import { invalidateAll } from '$app/navigation';
	let r = $state({} as NewRepairTraceReq);

	let access = CheckAndGetJWT('parsed').access || 'user';

	let {
		t,
		view = $bindable<'trace' | 'cancel' | 'update'>(),
		open = $bindable<boolean>(),
		isUpReady = $bindable<boolean>(),
		q,
		q1,
		onUpdated
	}: {
		t: Ticket | null;
		view: 'trace' | 'cancel' | 'update';
		open: boolean;
		isUpReady: boolean;
		q: NotificationQueue; //页面的
		q1: NotificationQueue; //模态框的
		onUpdated?: () => void | Promise<void>;
	} = $props();

	$effect(() => {
		if (isUpReady) {
			checkAndSubmitAndExit();
			isUpReady = false;
		}
	});

	$effect(() => {
		if (t) {
			r.new_category = t.category;
			r.new_priority = t.priority;
		}
	});

	//检查合法性，提交更新，退出模态框
	async function checkAndSubmitAndExit() {
		//检查部分
		try {
			$inspect(r).with(console.log);
			assert(t != null, '工单不能为空');
			assert(r.new_status !== '', '新状态不能为空。');
			r.tid = t.tid;
			assert(
				r.new_status !== 'scheduled' || r.new_appointment != null,
				'新状态是“已预约”时，预约时间不能为空。'
			);
			assert(r.remark != null && r.remark.trim().length > 0, '备注不能为空。');
			assert(r.remark.length <= 500, '备注不能超过500字。');

			if (r.new_category === t.category) {
				r.new_category = undefined;
			}

			if (r.new_priority === t.priority) {
				r.new_priority = undefined;
			}
		} catch (e: any) {
			console.error('检查更新状态失败', e);
			q1.add({
				kind: 'error',
				title: '数据校验失败',
				subtitle: e.message || '未知错误',
				timeout: 5000
			});
			return;
		}
		//提交部分
		try {
			//退出模态框
			open = false;
			view = 'trace';

			let res = await NewRepairTrace(r);
			if (!res.success) {
				throw new Error(res.msg || '未知错误');
			}
			q.add({
				kind: 'success',
				title: '更新状态成功',
				timeout: 3000
			});
		} catch (e: any) {
			console.error('提交更新状态失败', e);
			q.add({
				kind: 'error',
				title: '提交更新状态失败',
				subtitle: e.response?.data?.msg || e.message || '未知错误',
				timeout: 5000
			});
		} finally {
			await onUpdated?.();
		}
	}

	function assert(condition: boolean, expression: string) {
		if (!condition) {
			throw new Error(`无效的输入: ${expression}`);
		}
	}
</script>

<Select labelText="工单新状态" bind:selected={r.new_status} required={true}>
	<SelectItem value="" text="请选择工单的新状态..." disabled={true} hidden={true} />
	<SelectItem value="scheduled" text="已预约" style="color: #2563eb;" />
	<SelectItem value="delay" text="改日修" style="color: #3b82f6;" />
	<SelectItem value="escalated" text="已上报" style="color: #16a34a;" />
	<SelectItem value="solved" text="已解决" style="color: #16a34a;" />
	<SelectItem value="canceled" text="已取消" style="color: var(--color-red-600);" />
	{#if IsAdmin(access)}
		<SelectItem value="fresh" text="待解决（重启报修）" style="color: var(--color-red-600);" />
	{/if}
</Select>

{#if r.new_status === 'scheduled'}
	<br />
	<DatePicker
		datePickerType="single"
		on:change={(e) => (r.new_appointment = RFC3339((e.detail as { dateStr: string }).dateStr))}
	>
		<DatePickerInput labelText="预约的时间" placeholder="x年y月z日" required={true} />
	</DatePicker>
{/if}

<br />
<Select
	labelText="工单新故障类型"
	bind:selected={r.new_category}
	helperText="如果情况变化或信息有误，你可以修改该报修的故障类型"
>
	<SelectItem value="first-install" text="新装" />
	<SelectItem value="low-speed" text="网速慢" />
	<SelectItem value="ip-or-device" text="IP或设备问题" />
	<SelectItem value="client-or-account" text="客户端或账号问题" />
	<SelectItem value="others" text="其它问题" />
</Select>

{#if IsAdmin(access)}
	<br />
	<Select
		labelText="工单新优先级"
		bind:selected={r.new_priority}
		helperText="你可以修改该报修的优先级，更高优先级会在系统中优先显示"
	>
		<SelectItem value="highest" text="最高优先级！" style="color: var(--color-red-600);" />
		<SelectItem value="assigned" text="运营商工单" style="color: #2563eb;" />
		<SelectItem value="mainline" text="主线任务" />
		<SelectItem value="normal" text="一般报修" />
		<SelectItem value="in-passing" text="顺路看看" />
		<SelectItem value="least" text="不紧急" />
	</Select>
{/if}

<br />
<TextArea labelText="备注" placeholder="本次工单更新的备注" bind:value={r.remark} />
