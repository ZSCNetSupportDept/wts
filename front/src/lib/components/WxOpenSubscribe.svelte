<script lang="ts">
	import { onMount } from 'svelte';
	import { GetJsApiConfig } from '$lib/api';
	import { markSubscribeAsked, hasAskedSubscribe } from '$lib/states/wechatSubscribeStatus';

	let {
		templateId = '',
		scene = 2,
		onSuccess,
		onError
	}: {
		templateId?: string;
		scene?: number;
		onSuccess?: (detail: { errMsg: string; subscribeDetails: string }) => void;
		onError?: (detail: { errMsg: string; errCode: string }) => void;
	} = $props();

	let ready = $state(false);
	let errorMsg = $state('');
	let container: HTMLSpanElement | undefined = $state(undefined);

	// 加载微信 JS-SDK 脚本
	function loadWxSdk(): Promise<void> {
		return new Promise((resolve, reject) => {
			if ((window as any).wx) {
				resolve();
				return;
			}
			const script = document.createElement('script');
			script.src = 'https://res.wx.qq.com/open/js/jweixin-1.6.0.js';
			script.onload = () => resolve();
			script.onerror = () => reject(new Error('加载微信 JS-SDK 失败'));
			document.head.appendChild(script);
		});
	}

	// 初始化 wx.config
	async function initWxConfig() {
		if (hasAskedSubscribe()) {
			errorMsg = '本次会话已询问过订阅授权';
			return;
		}

		const ua = navigator.userAgent.toLowerCase();
		if (!ua.includes('micromessenger')) {
			errorMsg = '请在微信中打开';
			return;
		}

		try {
			await loadWxSdk();
			const wx = (window as any).wx;

			// 获取当前页面 URL（去除 hash 部分，微信签名要求）
			const url = window.location.href.split('#')[0];
			const res = await GetJsApiConfig(url);
			if (!res.success) {
				throw new Error(res.msg || '获取 JS-SDK 配置失败');
			}

			wx.config({
				debug: false,
				appId: res.appid,
				timestamp: res.timestamp,
				nonceStr: res.nonce_str,
				signature: res.signature,
				jsApiList: [],
				openTagList: ['wx-open-subscribe']
			});

			wx.ready(() => {
				ready = true;
				// 在 DOM 就绪后插入开放标签
				insertOpenTag();
			});

			wx.error((err: any) => {
				errorMsg = `wx.config 失败: ${err.errMsg || '未知错误'}`;
			});
		} catch (e: any) {
			errorMsg = e.message || '初始化失败';
		}
	}

	// 使用 DOM 操作插入 wx-open-subscribe 标签，避免 Svelte 解析 script 标签
	function insertOpenTag() {
		if (!container || !templateId) return;

		const wrapper = document.createElement('wx-open-subscribe');
		wrapper.setAttribute('template', templateId);
		wrapper.id = 'wx-open-subscribe-btn';

		// 样式插槽
		const styleScript = document.createElement('script');
		styleScript.type = 'text/wxtag-template';
		styleScript.slot = 'style';
		styleScript.textContent = `
			.subscribe-btn {
				display: inline-block;
				padding: 8px 16px;
				background-color: #07c160;
				color: #fff;
				border: none;
				border-radius: 4px;
				font-size: 14px;
				cursor: pointer;
			}
		`;

		// 按钮插槽
		const btnScript = document.createElement('script');
		btnScript.type = 'text/wxtag-template';
		btnScript.textContent = '<button class="subscribe-btn">订阅报修进度通知</button>';

		wrapper.appendChild(styleScript);
		wrapper.appendChild(btnScript);

		// 绑定事件
		wrapper.addEventListener('success', (e: Event) => {
			markSubscribeAsked();
			onSuccess?.((e as CustomEvent).detail);
		});
		wrapper.addEventListener('error', (e: Event) => {
			markSubscribeAsked();
			onError?.((e as CustomEvent).detail);
		});

		container.innerHTML = '';
		container.appendChild(wrapper);
	}

	onMount(() => {
		initWxConfig();

		// 监听开放标签错误（低版本微信等场景）
		const handler = (e: Event) => {
			const detail = (e as CustomEvent).detail;
			errorMsg = `开放标签不可用: ${detail.errMsg || '未知原因'}`;
		};
		document.addEventListener('WeixinOpenTagsError', handler);
		return () => document.removeEventListener('WeixinOpenTagsError', handler);
	});
</script>

{#if ready && templateId}
	<span bind:this={container}></span>
{:else if errorMsg}
	<span style="color: #999; font-size: 12px;">{errorMsg}</span>
{:else}
	<span style="color: #999; font-size: 12px;">加载中...</span>
{/if}
