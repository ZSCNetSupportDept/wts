import { GetJsApiConfig } from '$lib/api';

type WxConfig = {
	debug: boolean;
	appId: string;
	timestamp: number;
	nonceStr: string;
	signature: string;
	jsApiList: string[];
	openTagList: string[];
};

type WxSdk = {
	config: (config: WxConfig) => void;
	ready: (callback: () => void) => void;
	error: (callback: (error: { errMsg?: string }) => void) => void;
};

type WxWindow = Window &
	typeof globalThis & {
		__ENTRY_URL__?: string;
		wx?: WxSdk;
	};

let sdkPromise: Promise<WxSdk> | undefined;
let configPromise: Promise<void> | undefined;

function loadWxSdk(): Promise<WxSdk> {
	const wxWindow = window as WxWindow;
	if (wxWindow.wx) {
		return Promise.resolve(wxWindow.wx);
	}
	if (sdkPromise) {
		return sdkPromise;
	}

	sdkPromise = new Promise<WxSdk>((resolve, reject) => {
		const script = document.createElement('script');
		script.src = 'https://res.wx.qq.com/open/js/jweixin-1.6.0.js';
		script.onload = () => {
			const loadedWindow = window as WxWindow;
			if (loadedWindow.wx) {
				resolve(loadedWindow.wx);
				return;
			}
			reject(new Error('微信 JS-SDK 未初始化'));
		};
		script.onerror = () => reject(new Error('加载微信 JS-SDK 失败'));
		document.head.appendChild(script);
	});

	return sdkPromise;
}

// 微信对 SPA 页面使用首次入口 URL 签名，app.html 在路由初始化前保存了该值。
export async function configureWxOpenSubscribe(): Promise<void> {
	if (configPromise) {
		return configPromise;
	}

	const pending = (async () => {
		const wx = await loadWxSdk();
		const wxWindow = window as WxWindow;
		const signUrl = wxWindow.__ENTRY_URL__ || window.location.href.split('#')[0];
		const res = await GetJsApiConfig(signUrl);
		if (!res.success) {
			throw new Error(res.msg || '获取 JS-SDK 配置失败');
		}

		await new Promise<void>((resolve, reject) => {
			wx.ready(resolve);
			wx.error((error) => reject(new Error(error.errMsg || 'wx.config 失败')));
			wx.config({
				debug: false,
				appId: res.appid,
				timestamp: res.timestamp,
				nonceStr: res.nonce_str,
				signature: res.signature,
				jsApiList: [],
				openTagList: ['wx-open-subscribe']
			});
		});
	})();

	configPromise = pending.catch((error) => {
		configPromise = undefined;
		throw error;
	});
	return configPromise;
}