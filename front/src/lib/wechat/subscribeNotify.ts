import type { SubscribeConfigRes } from '$lib/types/apiResponse';
import { markSubscribeAsked } from '$lib/states/wechatSubscribeStatus';

// 构建微信 H5 订阅授权链接。
// redirectPath 为授权完成后回跳的站内路径（如 /repair），scene 仅用于统计来源。
export function buildSubscribeURL(cfg: SubscribeConfigRes, redirectPath: string, scene: number): string {
        const redirect = encodeURIComponent(`${window.location.origin}${redirectPath}`);
        return (
                `https://mp.weixin.qq.com/mp/subscribemsg?action=get_confirm` +
                `&appid=${cfg.appid}&scene=${scene}&template_id=${cfg.template_id}` +
                `&redirect_url=${redirect}&reserved=#wechat_redirect`
        );
}

// 跳转微信授权页（整页跳转）。

export function AskForWechatNotifySubscription(cfg: SubscribeConfigRes, redirectPath: string, scene: number) {
        markSubscribeAsked();
        window.location.href = buildSubscribeURL(cfg, redirectPath, scene);
}
