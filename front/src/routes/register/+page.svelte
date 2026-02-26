<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	import {
		TextInput,
		RadioButtonGroup,
		RadioButton,
		Select,
		SelectItem,
		SelectItemGroup,
		Button,
		Checkbox,
		ComposedModal,
		ModalBody,
		ModalFooter,
		ModalHeader,
		NotificationQueue,
		Loading
	} from 'carbon-components-svelte';
	import type { RegisterReq } from '$lib/types/apiRequest';
	import { invalidState } from '$lib/types/invalidState.svelte';
	import { Register } from '$lib/api';
	import { TheLastPage } from '$lib/states/theLastPage.svelte';
	import { onMount } from 'svelte';
	import { IsUnregistered } from '$lib/types/enum';
	import { Guard } from '$lib/jwt';

	let q: NotificationQueue;

	let notLoading: boolean = $state(true);

	onMount(() => Guard(IsUnregistered));

	//注册请求体状态变量
	let req = $state({
		block: '0'
	} as unknown as RegisterReq);

	//模态框状态变量
	let checked = $state(false);
	let open = $state(false);

	//无效状态变量

	let name = new invalidState();
	let sid = new invalidState();
	let account = new invalidState();
	let phone = new invalidState();
	let block = new invalidState();
	let room = new invalidState();

	//检查输入合法性
	function check(): boolean {
		let ok = false;
		// 重置所有无效状态
		name.reset();
		sid.reset();
		account.reset();
		phone.reset();
		block.reset();
		room.reset();
		// 首先，姓名学号不能超过15个字符且不能为空
		name.assert(req.name.length > 0, '姓名不能为空');
		name.assert(req.name.length <= 15, '姓名不能超过15个字符');
		sid.assert(req.sid.length > 0, '学号不能为空');
		sid.assert(req.sid.length <= 15, '学号不能超过15个字符');
		// 然后，校园网账号和手机号是中国大陆的11位手机号码
		const phoneRegex = /^1[3-9]\d{9}$/;
		account.assert(
			req.isp === 'others' || phoneRegex.test(req.account),
			'校园网账号应为有效的11位手机号'
		);
		account.assert(req.isp !== 'others' || req.account.length > 0, '请输入您的校园网账号');
		account.assert(req.isp !== 'others' || req.account.length <= 15, '校园网账号不能超过15个字符');
		phone.assert(phoneRegex.test(req.phone), '联系电话应为有效的11位手机号');
		// 接着，宿舍楼不能为空且房间号不能超过5个字符且不能为空
		block.assert(req.block !== '0', '请选择宿舍楼');
		room.assert(req.room.length > 0, '房间号不能为空');
		room.assert(
			req.block === 'other' || /^[0-9]{1,4}$/.test(req.room),
			'请填写一个5位以内的纯数字...'
		);

		//最后，总结断言结果
		if (
			name.notOK ||
			sid.notOK ||
			account.notOK ||
			phone.notOK ||
			block.notOK ||
			room.notOK ||
			req.isp === undefined
		) {
			ok = false;
		} else {
			ok = true;
		}

		return ok;
	}

	// 在不合法时跳转到对应的地方以便用户修改
	function jump() {
		if (name.notOK) {
			document.getElementById('name')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (sid.notOK) {
			document.getElementById('sid')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (account.notOK) {
			document.getElementById('account')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (phone.notOK) {
			document.getElementById('phone')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (block.notOK) {
			document.getElementById('block')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		} else if (room.notOK) {
			document.getElementById('room')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
		if (req.isp === undefined) {
			document.getElementById('isp')?.scrollIntoView({ behavior: 'smooth', block: 'center' });
			q.add({
				kind: 'error',
				title: '请选择运营商',
				subtitle: '请选择您校园卡的运营商',
				timeout: 5000
			});
		}
	}

	async function submit() {
		open = false;
		checked = false;
		notLoading = false;

		try {
			const res = await Register(req);
			notLoading = true;
			if (!res.success) {
				throw new Error(res.msg || '注册失败.........');
			}
			q.add({
				kind: 'success',
				title: '注册成功',
				timeout: 3000
			});
			setTimeout(() => {
				TheLastPage.Write('/');
				window.location.href = '/login';
			}, 3900);
		} catch (e: any) {
			notLoading = true;
			console.error('register fail:', e);
			const errMsg = e.response?.data?.msg || e.message || '未知错误';
			q.add({
				kind: 'error',
				title: '注册失败',
				subtitle: errMsg + '，请重试',
				timeout: 5000
			});
		}
	}
</script>

<h1>注册</h1>
<br />
<hr />
<br />
<p style="font-size: 15px;">
	<i
		>您似乎还没有注册。在使用该系统之前，请先注册您的个人信息。这些信息将被用于我们上门为您维修网络问题。</i
	>
</p>
<br />
<TextInput
	id="name"
	labelText="姓名"
	placeholder="请输入您的真实姓名..."
	bind:value={req.name}
	invalid={name.notOK}
	invalidText={name.txt}
/>
<br />
<br />
<TextInput
	id="sid"
	labelText="学号"
	placeholder="请输入您的学号..."
	bind:value={req.sid}
	invalid={sid.notOK}
	invalidText={sid.txt}
/>
<br />
<br />
<RadioButtonGroup
	id="isp"
	legendText="请选择您校园卡的运营商"
	bind:selected={req.isp}
	required={true}
>
	<RadioButton labelText="电信" value="telecom" />
	<RadioButton labelText="联通" value="unicom" />
	<RadioButton labelText="移动" value="mobile" />
	<RadioButton labelText="其它" value="others" />
</RadioButtonGroup>
<br />
<br />
<TextInput
	id="account"
	labelText="校园网账号"
	placeholder="请输入您校园卡的手机号..."
	bind:value={req.account}
	invalid={account.notOK}
	invalidText={account.txt}
/>
<br />
<br />
<hr />
<br />
<br />
<TextInput
	id="phone"
	labelText="电话"
	placeholder="请输入您的联系电话..."
	bind:value={req.phone}
	invalid={phone.notOK}
	invalidText={phone.txt}
/>
<br />
<br />
<Select
	id="block"
	labelText="宿舍楼"
	bind:selected={req.block}
	invalid={block.notOK}
	invalidText={block.txt}
>
	<SelectItem value="0" text="请选择您的所住的宿舍楼..." disabled hidden />
	<SelectItemGroup label="凤翔宿舍区">
		<SelectItem value="1" text="1栋" />
		<SelectItem value="2" text="2栋" />
		<SelectItem value="3" text="3栋" />
		<SelectItem value="4" text="4栋" />
		<SelectItem value="5" text="5栋" />
		<SelectItem value="6" text="6栋" />
	</SelectItemGroup>
	<SelectItemGroup label="北门宿舍区">
		<SelectItem value="7" text="7栋" />
		<SelectItem value="8" text="8栋" />
		<SelectItem value="9" text="9栋" />
		<SelectItem value="10" text="10栋" />
		<SelectItem value="11" text="11栋" />
	</SelectItemGroup>
	<SelectItemGroup label="东门宿舍区">
		<SelectItem value="12" text="12栋" />
		<SelectItem value="13" text="13栋" />
		<SelectItem value="14" text="14栋" />
		<SelectItem value="15" text="15栋" />
		<SelectItem value="20" text="20栋" />
		<SelectItem value="21" text="21栋" />
		<SelectItem value="22" text="22栋" />
	</SelectItemGroup>
	<SelectItemGroup label="歧头山宿舍区">
		<SelectItem value="16" text="16栋" />
		<SelectItem value="17" text="17栋" />
		<SelectItem value="18" text="18栋" />
		<SelectItem value="19" text="19栋" />
	</SelectItemGroup>
	<SelectItemGroup label="香晖苑">
		<SelectItem value="XHA" text="香晖苑-A栋" />
		<SelectItem value="XHB" text="香晖苑-B栋" />
		<SelectItem value="XHC" text="香晖苑-C栋" />
		<SelectItem value="XHD" text="香晖苑-D栋" />
	</SelectItemGroup>
	<SelectItemGroup label="朝晖苑">
		<SelectItem value="ZH" text="朝晖苑" />
	</SelectItemGroup>
	<SelectItemGroup label="其它">
		<SelectItem value="other" text="其它" />
	</SelectItemGroup>
</Select>
<br />
<br />
<TextInput
	id="room"
	labelText="房间号"
	placeholder="请输入您所住的房间..."
	bind:value={req.room}
	invalid={room.notOK}
	invalidText={room.txt}
/>
<br />
<br />
<br />
<Button
	on:click={() => {
		check() ? (open = true) : jump();
	}}>提交注册</Button
>

<ComposedModal
	bind:open
	on:close={() => {
		((open = false), (checked = false));
	}}
	class="mobile-floating-modal"
>
	<ModalHeader title="确认您的信息" />

	<ModalBody hasForm>
		<Checkbox labelText="我确认所填信息准确无误，真实有效，且未盗用他人信息" bind:checked />
		<br />
		<br />
	</ModalBody>

	<ModalFooter>
		<Button kind="secondary" on:click={() => ((open = false), (checked = false))}>取消</Button>
		<Button
			kind="primary"
			disabled={!checked}
			on:click={() => {
				submit();
			}}>确认并提交</Button
		>
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
