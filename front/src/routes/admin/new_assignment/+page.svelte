<script lang="ts">
	import { CheckAndGetJWT, Guard } from '$lib/jwt';
	import type { NewTicketReq } from '$lib/types/apiRequest';
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
	import { RFC3339 } from '$lib/types/RFC3339';
	import { onMount } from 'svelte';
	import { IsAdmin, ISPMap, IsUser } from '$lib/types/enum';

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
		SelectItem,
		SelectItemGroup,
		Toggle
	} from 'carbon-components-svelte';
	import { IsRFC3339 } from '$lib/types/RFC3339';
	import { invalidState } from '$lib/types/invalidState.svelte';
	import { NewTicket } from '$lib/api';
	import { goto } from '$app/navigation';
	import { invalid } from '@sveltejs/kit';
	import type { WtsISP, WtsBlock } from '$lib/types/enum';
	import { BlockMap } from '$lib/types/enum';

	let notLoading: boolean = $state(true);

	let q: NotificationQueue;

	let r = $state({
		priority: 'assigned'
	} as NewTicketReq);

	let r1 = $state({
		room: undefined as string,
		isp: undefined as WtsISP,
		phone: undefined as string,
		account: undefined as string,
		name: undefined as string
	});

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
	let issuer = new invalidState();
	let room = new invalidState();
	let phone = new invalidState();
	let account = new invalidState();
	let name = new invalidState();

	function check(): boolean {
		notLoading = false;
		let ok = false;
		occurAt.reset();
		appointedAt.reset();
		description.reset();
		notes.reset();
		issuer.reset();
		room.reset();
		phone.reset();
		account.reset();
		name.reset();

		occurAt.assert(!r.occur_at || IsRFC3339(r.occur_at), '请输入正确的故障发生时间');
		appointedAt.assert(!r.appointed_at || IsRFC3339(r.appointed_at), '请输入正确的预约时间');
		notes.assert(!!r.notes && r.notes.length > 0, '请填写工单情况描述');
		notes.assert(!r.notes || r.notes.length <= 200, '字数太多了...请控制在200字以内');
		issuer.assert(!!r.issuer_sid && r.issuer_sid !== '0', '请选择宿舍楼');
		room.assert(!!r1.room && r1.room.length > 0, '请填写房间号');
		room.assert(!r1.room || r1.room.length <= 5, '房间号太长了，请控制在5字以内');
		phone.assert(!r1.phone || /^\d{11}$/.test(r1.phone), '请输入正确的联系电话');
		account.assert(!r1.account || r1.account.length <= 25, '账号太长了，请控制在25字以内');
		name.assert(!r1.name || r1.name.length <= 20, '姓名太长了，请控制在20字以内');
		
		let ispOK = r1.isp !== undefined;

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
		return !(occurAt.notOK || appointedAt.notOK || description.notOK || notes.notOK || issuer.notOK || room.notOK || phone.notOK || account.notOK || name.notOK || !ispOK);
	}

	let issuerToBlock: Record<string, WtsBlock> = {
		gd1: '1',
		gd2: '2',
		gd3: '3',
		gd4: '4',
		gd5: '5',
		gd6: '6',
		gd7: '7',
		gd8: '8',
		gd9: '9',
		gd10: '10',
		gd11: '11',
		gd12: '12',
		gd13: '13',
		gd14: '14',
		gd15: '15',
		gd16: '16',
		gd17: '17',
		gd18: '18',
		gd19: '19',
		gd20: '20',
		gd21: '21',
		gd22: '22',
		gdXHA: 'XHA',
		gdXHB: 'XHB',
		gdXHC: 'XHC',
		gdXHD: 'XHD',
		gdZH: 'ZH',
		gdOther: 'other'
	};

	function processDescription() {
		let blockRoom = BlockMap[issuerToBlock[r.issuer_sid]] + '-' + r1.room;
		let isp = ISPMap[r1.isp];
		let name = r1.name;
		let phone = r1.phone;
		let account = r1.account;

		let parts = [];
		let line1 = [blockRoom, isp].filter(Boolean).join('，');
		if (line1) parts.push(line1);

		if (name && phone) {
			parts.push(`${name} ${phone}`);
		} else if (name) {
			parts.push(`姓名：${name}`);
		} else if (phone) {
			parts.push(`联系电话：${phone}`);
		}

		if (account) parts.push('账号: ' + account);

		r.description = parts.join('\n');
	}

	async function submit() {
		try {
			processDescription();
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
			setTimeout(() => goto('/admin'), 3900);
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
		} else if (issuer.notOK) {
			document.getElementById('block')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (room.notOK) {
			document.getElementById('room')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (phone.notOK) {
			document.getElementById('phone')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (account.notOK) {
			document.getElementById('account')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (name.notOK) {
			document.getElementById('name')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
		if (r1.isp === undefined) {
			document.getElementById('isp')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
			q.add({
				kind: 'error',
				title: '请选择校园卡运营商',
				timeout: 3000
			});
		}
	}

	onMount(() => Guard(IsAdmin));
</script>

<h1>录入运营商工单</h1>
<br />
<hr />
<br />
<p>
	在这里向系统中增添不确定用户信息的运营商工单，您可以先<a href="/admin/filter_users">搜索一下</a>数据库中存在的用户再使用<a href="/admin/add_ticket">一般增添界面</a>。
</p>
<br />
<Select
	id="block"
	labelText="该工单所在的宿舍楼"
	bind:selected={r.issuer_sid}
	invalid={issuer.notOK}
	invalidText={issuer.txt}
>
	<SelectItem value="0" text="请选择宿舍" disabled hidden />
	<SelectItemGroup label="凤翔宿舍区">
		<SelectItem value="gd1" text="1栋" />
		<SelectItem value="gd2" text="2栋" />
		<SelectItem value="gd3" text="3栋" />
		<SelectItem value="gd4" text="4栋" />
		<SelectItem value="gd5" text="5栋" />
		<SelectItem value="gd6" text="6栋" />
	</SelectItemGroup>
	<SelectItemGroup label="北门宿舍区">
		<SelectItem value="gd7" text="7栋" />
		<SelectItem value="gd8" text="8栋" />
		<SelectItem value="gd9" text="9栋" />
		<SelectItem value="gd10" text="10栋" />
		<SelectItem value="gd11" text="11栋" />
	</SelectItemGroup>
	<SelectItemGroup label="东门宿舍区">
		<SelectItem value="gd12" text="12栋" />
		<SelectItem value="gd13" text="13栋" />
		<SelectItem value="gd14" text="14栋" />
		<SelectItem value="gd15" text="15栋" />
		<SelectItem value="gd20" text="20栋" />
		<SelectItem value="gd21" text="21栋" />
		<SelectItem value="gd22" text="22栋" />
	</SelectItemGroup>
	<SelectItemGroup label="歧头山宿舍区">
		<SelectItem value="gd16" text="16栋" />
		<SelectItem value="gd17" text="17栋" />
		<SelectItem value="gd18" text="18栋" />
		<SelectItem value="gd19" text="19栋" />
	</SelectItemGroup>
	<SelectItemGroup label="香晖苑">
		<SelectItem value="gdXHA" text="香晖苑-A栋" />
		<SelectItem value="gdXHB" text="香晖苑-B栋" />
		<SelectItem value="gdXHC" text="香晖苑-C栋" />
		<SelectItem value="gdXHD" text="香晖苑-D栋" />
	</SelectItemGroup>
	<SelectItemGroup label="朝晖苑">
		<SelectItem value="gdZH" text="朝晖苑" />
	</SelectItemGroup>
	<SelectItemGroup label="其它">
		<SelectItem value="gdOther" text="其它" />
	</SelectItemGroup>
</Select>
<br />
<br />
<TextInput
	id="room"
	labelText="用户所住房间号"
	placeholder="请输入报修用户所住的房间..."
	bind:value={r1.room}
	invalid={room.notOK}
	invalidText={room.txt}
/>
<br />
<br />
<RadioButtonGroup id="isp" legendText="请选择用户的运营商" bind:selected={r1.isp} required={true}>
	<RadioButton labelText="电信" value="telecom" />
	<RadioButton labelText="联通" value="unicom" />
	<RadioButton labelText="移动" value="mobile" />
	<RadioButton labelText="其它" value="others" />
</RadioButtonGroup>
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
	labelText="工单情况描述"
	placeholder="具体是何情况，有什么故障，还有对值班人员的其它备注。"
	bind:value={r.notes}
	invalid={notes.notOK}
	invalidText={notes.txt}
/>
<br />
<br />
<hr />
<br />
<br />
<DatePicker datePickerType="single" on:change={onOccurDateChange}>
	<DatePickerInput
		labelText="故障发生的日期（选填）"
		placeholder="请输入日期"
		invalid={occurAt.notOK}
		invalidText={occurAt.txt}
	/>
</DatePicker>
<br />
<br />
<DatePicker datePickerType="single" on:change={onAppointDateChange}>
	<DatePickerInput
		labelText="预约上门维修的日期（选填）"
		placeholder="当天4:30~6:00用户需要在宿舍"
		invalid={appointedAt.notOK}
		invalidText={appointedAt.txt}
	/>
</DatePicker>
<br />
<br />
<TextInput
	id="phone"
	labelText="用户的联系电话（选填）"
	placeholder="请输入报修用户的联系电话"
	bind:value={r1.phone}
	invalid={phone.notOK}
	invalidText={phone.txt}
/>
<br />
<br />
<TextInput
	id="account"
	labelText="用户的账号（选填）"
	placeholder="请输入报修用户的校园网账号"
	bind:value={r1.account}
	invalid={account.notOK}
	invalidText={account.txt}
/>
<br />
<br />
<TextInput
	id="name"
	labelText="用户的姓名（选填）"
	placeholder="请输入报修用户的姓名"
	bind:value={r1.name}
	invalid={name.notOK}
	invalidText={name.txt}
/>
<br/>
<br/>
<Toggle
	labelText="紧急派单!!!!!"
	on:change={() => {
		r.priority = r.priority === 'assigned' ? 'highest' : 'assigned';
	}}
	labelB='On （没事别乱用）'
/>
<br/>
<br/>
<Button on:click={handleSubmit}>提交</Button>

<NotificationQueue bind:this={q} />

<Loading active={!notLoading} />
