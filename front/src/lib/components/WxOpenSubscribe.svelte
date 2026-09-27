<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { configureWxOpenSubscribe } from '$lib/wechat/wxOpenTagSdk';

	type SubscribeDetail = Record<string, unknown> & {
		errMsg?: string;
		errCode?: string;
		subscribeDetails?: unknown;
	};

	type UnavailableDetail = {
		reason: string;
	};

	let {
		templateId = '',
		label = '订阅报修进度通知',
		width = '10rem',
		onSuccess,
		onError,
		onUnavailable
	}: {
		templateId?: string;
		label?: string;
		width?: string;
		onSuccess?: (detail: SubscribeDetail) => void;
		onError?: (detail: SubscribeDetail) => void;
		onUnavailable?: (detail: UnavailableDetail) => void;
	} = $props();

	let ready = $state(false);
	let unavailable = $state(false);
	let container: HTMLSpanElement | undefined = $state(undefined);
	let settled = false;
	let active = false;

	function escapeHtml(value: string): string {
		return value.replace(/[&<>"']/g, (character) => {
			const entities: Record<string, string> = {
				'&': '&amp;',
				'<': '&lt;',
				'>': '&gt;',
				'"': '&quot;',
				"'": '&#39;'
			};
			return entities[character];
		});
	}

	function markUnavailable(reason: string) {
		if (!active || unavailable) {
			return;
		}
		unavailable = true;
		onUnavailable?.({ reason });
	}

	function settle(kind: 'success' | 'error', detail: SubscribeDetail) {
		if (settled) {
			return;
		}
		settled = true;
		if (kind === 'success') {
			onSuccess?.(detail);
			return;
		}
		onError?.(detail);
	}

	function insertOpenTag() {
		if (!container || !templateId) {
			markUnavailable('开放标签容器未就绪');
			return;
		}

		const wrapper = document.createElement('wx-open-subscribe');
		wrapper.setAttribute('template', templateId);
		wrapper.style.display = 'inline-block';
		wrapper.style.width = width;
		wrapper.style.height = '48px';

		const styleScript = document.createElement('script');
		styleScript.type = 'text/wxtag-template';
		styleScript.slot = 'style';
		styleScript.textContent = `
			<style>
			.subscribe-btn {
				display: block;
				box-sizing: border-box;
				width: 100%;
				height: 100%;
				padding: 0 16px;
				background-color: #0f62fe;
				color: #ffffff;
				border: 0;
				border-radius: 0;
				font-family: 'IBM Plex Sans', 'Helvetica Neue', Arial, sans-serif;
				font-size: 14px;
				line-height: 18px;
				text-align: left;
				cursor: pointer;
			}
			.subscribe-btn:active {
				background-color: #002d9c;
			}
			</style>
		`;

		const buttonScript = document.createElement('script');
		buttonScript.type = 'text/wxtag-template';
		buttonScript.textContent = `<button class="subscribe-btn">${escapeHtml(label)}</button>`;

		wrapper.append(styleScript, buttonScript);
		wrapper.addEventListener('success', (event) => {
			settle('success', ((event as CustomEvent).detail || {}) as SubscribeDetail);
		});
		wrapper.addEventListener('error', (event) => {
			settle('error', ((event as unknown as CustomEvent).detail || {}) as SubscribeDetail);
		});

		container.replaceChildren(wrapper);
	}

	async function initWxConfig() {
		if (!templateId) {
			markUnavailable('未配置订阅消息模板');
			return;
		}
		if (!navigator.userAgent.toLowerCase().includes('micromessenger')) {
			markUnavailable('请在微信中打开');
			return;
		}

		try {
			await configureWxOpenSubscribe();
			if (!active) {
				return;
			}
			ready = true;
			await tick();
			if (active) {
				insertOpenTag();
			}
		} catch (error) {
			const message = error instanceof Error ? error.message : '初始化微信订阅提醒失败';
			markUnavailable(message);
		}
	}

	onMount(() => {
		active = true;
		const handler = (event: Event) => {
			const detail = (event as CustomEvent<{ errMsg?: string }>).detail;
			markUnavailable(detail?.errMsg || '开放标签不可用');
		};
		document.addEventListener('WeixinOpenTagsError', handler);
		void initWxConfig();

		return () => {
			active = false;
			document.removeEventListener('WeixinOpenTagsError', handler);
		};
	});
</script>

{#if ready && templateId && !unavailable}
	<span bind:this={container} style="display: inline-block; width: {width}; height: 48px;"></span>
{:else}
	<button class="subscribe-placeholder" type="button" disabled style="width: {width};">
		{label}
	</button>
{/if}

<style>
	.subscribe-placeholder {
		box-sizing: border-box;
		height: 48px;
		padding: 0 16px;
		background: #8d8d8d;
		color: #ffffff;
		border: 0;
		border-radius: 0;
		font-family: 'IBM Plex Sans', 'Helvetica Neue', Arial, sans-serif;
		font-size: 14px;
		line-height: 18px;
		text-align: left;
		opacity: 0.6;
	}
</style>
