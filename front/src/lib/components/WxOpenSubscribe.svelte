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
		width,
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

	function carbonButtonWidth(text: string): string {
		// Carbon 默认按钮的左右空间为 15px + 63px，中文正文约为 14px/字。
		const textWidthInRem = [...text].length * 0.875;
		return `${Math.max(6.75, textWidthInRem + 5)}rem`;
	}

	let resolvedWidth = $derived(width || carbonButtonWidth(label));

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
		wrapper.style.width = resolvedWidth;
		wrapper.style.height = '3rem';
		wrapper.style.verticalAlign = 'top';

		const styleScript = document.createElement('script');
		styleScript.type = 'text/wxtag-template';
		styleScript.slot = 'style';
		styleScript.textContent = `
			<style>
			.subscribe-btn {
				position: relative;
				display: inline-flex;
				box-sizing: border-box;
				width: 100%;
				max-width: 20rem;
				min-height: 3rem;
				height: 3rem;
				flex-shrink: 0;
				align-items: center;
				justify-content: space-between;
				padding: calc(.875rem - 3px) 63px calc(.875rem - 3px) 15px;
				margin: 0;
				background-color: #0f62fe;
				color: #ffffff;
				border: 1px solid transparent;
				border-radius: 0;
				font-family: 'IBM Plex Sans', 'Helvetica Neue', Arial, sans-serif;
				font-size: 14px;
				font-weight: 400;
				line-height: 1.28572;
				letter-spacing: 0.16px;
				text-align: left;
				text-decoration: none;
				cursor: pointer;
				outline: none;
				vertical-align: top;
				transition: background 70ms cubic-bezier(0, 0, 0.38, 0.9), border-color 70ms cubic-bezier(0, 0, 0.38, 0.9), box-shadow 70ms cubic-bezier(0, 0, 0.38, 0.9), outline 70ms cubic-bezier(0, 0, 0.38, 0.9);
			}
			.subscribe-btn:hover {
				background-color: #0353e9;
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
	<span bind:this={container} style="display: inline-block; width: {resolvedWidth}; height: 3rem; vertical-align: top;"></span>
{:else}
	<button class="subscribe-placeholder" type="button" disabled style="width: {resolvedWidth};">
		{label}
	</button>
{/if}

<style>
	.subscribe-placeholder {
		position: relative;
		display: inline-flex;
		box-sizing: border-box;
		max-width: 20rem;
		min-height: 3rem;
		height: 3rem;
		align-items: center;
		justify-content: space-between;
		padding: calc(.875rem - 3px) 63px calc(.875rem - 3px) 15px;
		margin: 0;
		background: #c6c6c6;
		color: #8d8d8d;
		border: 1px solid transparent;
		border-radius: 0;
		font-family: 'IBM Plex Sans', 'Helvetica Neue', Arial, sans-serif;
		font-size: 14px;
		font-weight: 400;
		line-height: 1.28572;
		letter-spacing: 0.16px;
		text-align: left;
		cursor: not-allowed;
		vertical-align: top;
	}
</style>
