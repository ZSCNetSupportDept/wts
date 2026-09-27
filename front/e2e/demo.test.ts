import { expect, test } from '@playwright/test';

test('home page has expected h1', async ({ page }) => {
	await page.goto('/');
	await expect(page.locator('h1')).toBeVisible();
});

test('notification lander renders the URL summary', async ({ page }) => {
	await page.goto('/wx_notify_lander/?tid=42&status=solved&message=%E7%BD%91%E7%BA%BF%E5%B7%B2%E6%9B%B4%E6%8D%A2');

	await expect(page.getByText('工单 No.42')).toBeVisible();
	await expect(page.getByText('已解决')).toBeVisible();
	await expect(page.getByText('网线已更换')).toBeVisible();
	await expect(page.getByRole('button', { name: '查看工单详情' })).toBeVisible();
});
