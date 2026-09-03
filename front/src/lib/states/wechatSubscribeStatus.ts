// 微信订阅授权询问状态（sessionStorage 实现）。
// 无论用户接受还是拒绝，只要跳转过授权页一次，本轮session内就不再重复跳转。
// 用户关闭标签页后重新打开时会再次询问.

const KEY = 'wts_wx_subscribe_asked';

export function hasAskedSubscribe(): boolean {
        try {
                return sessionStorage.getItem(KEY) === '1';
        } catch {
                // sessionStorage 不可用时为避免无限循环跳转，视为已询问过
                return true;
        }
}

// 在跳转到微信授权页之前调用
export function markSubscribeAsked(): void {
        try {
                sessionStorage.setItem(KEY, '1');
        } catch {
        }
}
