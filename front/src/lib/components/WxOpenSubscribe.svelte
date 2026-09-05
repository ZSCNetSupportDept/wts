<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { GetJsApiConfig } from '$lib/api';
	import { markSubscribeAsked, hasAskedSubscribe } from '$lib/states/wechatSubscribeStatus';

	// 微信 JSSDK 签名校验用的是「页面初次加载时的入口 URL」（realAuthUrl），
	// SPA 前端路由后 location.href 会变，无法从运行时可靠获取。
	// app.html 在 <head> 最顶部、一切 SPA 逻辑执行前，已把真实入口 URL 缓存到 window.__ENTRY_URL__。
	// 此处直接读取；若异常缺失（理论上不会），回退到当前 location.href。
	const SIGN_URL =
		(window as any).__ENTRY_URL__ || window.location.href.split('#')[0];

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

			// 用 app.html 捕获的真实入口 URL 签名，与微信 realAuthUrl 严格一致。
			const url = SIGN_URL;
			console.log('[WxOpenSubscribe] 签名 URL:', url, '| 当前 href:', window.location.href);
			const res = await GetJsApiConfig(url);
			if (!res.success) {
				throw new Error(res.msg || '获取 JS-SDK 配置失败');
			}

			wx.config({
				debug: true,
				appId: res.appid,
				timestamp: res.timestamp,
				nonceStr: res.nonce_str,
				signature: res.signature,
				jsApiList: [],
				openTagList: ['wx-open-subscribe']
			});

			wx.ready(async () => {
				ready = true;
				// ready 置位后 Svelte 的 DOM 更新在微任务中批量刷新，
				// 必须等下一个 tick，<span bind:this={container}> 才会真正挂载。
				await tick();
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
	function insertOpenTag(retries = 5) {
		// bind:this 赋值可能晚于 tick() 完成，未就绪时退避重试，消除竞态
		if (!container || !templateId) {
			if (retries > 0) {
				setTimeout(() => insertOpenTag(retries - 1), 60);
			} else {
				errorMsg = '容器挂载超时';
			}
			return;
		}

		const wrapper = document.createElement('wx-open-subscribe');
		wrapper.setAttribute('template', templateId);
		wrapper.id = 'wx-open-subscribe-btn';
		// 关键：必须给开放标签本身设置明确宽高。
		// 手机微信的开放标签是原生组件同层渲染，按标签的布局尺寸绘制，
		// 若标签本身尺寸为 0（默认 inline 无内容），移动端会渲染为空白。
		wrapper.style.display = 'inline-block';
		wrapper.style.width = '180px';
		wrapper.style.height = '40px';

		// 样式插槽（官方示例要求内容用 <style> 标签包裹）
		const styleScript = document.createElement('script');
		styleScript.type = 'text/wxtag-template';
		styleScript.slot = 'style';
		styleScript.textContent = `
			<style>
			.subscribe-btn {
				display: block;
				width: 100%;
				height: 100%;
				background-color: #07c160;
				color: #fff;
				border: none;
				border-radius: 4px;
				font-size: 14px;
				cursor: pointer;
			}
			</style>
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
