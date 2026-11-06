import { expect, test, type Page } from '@playwright/test';

const today = new Intl.DateTimeFormat('sv-SE', { timeZone: 'Asia/Shanghai' }).format(new Date());
const people = [
	{ wid: '101', name: '测试组长', access: 'group-leader' },
	{ wid: '201', name: '测试正式成员', access: 'formal-member' },
	{ wid: '301', name: '测试实习甲', access: 'informal-member' },
	{ wid: '302', name: '测试实习乙', access: 'informal-member' },
	{ wid: '900', name: '测试科长', access: 'chief' }
];
const day = {
	date: today,
	leader: people[0],
	published: true,
	revision: 3,
	members: people.slice(0, 4).map((person, index) => ({
		...person,
		zone: index % 2 ? 'BM' : 'FX',
		area_leader: index === 1,
		extra: false,
		available: true
	}))
};

async function setup(page: Page, wid: string, pending: boolean | 'open' = false) {
	const me = people.find((p) => p.wid === wid)!;
	await page.setViewportSize({ width: 390, height: 844 });
	await page.addInitScript(
		({ me }) => {
			const encode = (value: object) => btoa(unescape(encodeURIComponent(JSON.stringify(value))));
			localStorage.setItem(
				'jwt',
				`${encode({ alg: 'HS256' })}.${encode({ ...me, name: me.name, sid: me.wid, openid: me.wid, exp: Date.now() / 1000 + 3600 })}.test-signature`
			);
		},
		{ me }
	);
	await page.route('**/api/**', async (route) => {
		const url = new URL(route.request().url());
		let data: any;
		if (url.pathname === '/api/v3/scheduler')
			data = {
				me,
				manager: me.access === 'chief',
				days: [day],
				free_days: [1, 2, 3, 4, 5],
				exceptions: []
			};
		else if (url.pathname.endsWith('/staff')) data = people.map((p) => ({ ...p, available: true }));
		else if (url.pathname.endsWith('/suggest'))
			data = {
				previous_date: '2026-10-01',
				quotas: { FX: 2, BM: 2 },
				assignments: day.members.map((m) => ({
					wid: m.wid,
					zone: m.zone,
					area_leader: m.area_leader
				}))
			};
		else if (url.pathname.endsWith('/swaps/review')) data = { id: 1, status: 'pending' };
		else if (url.pathname.endsWith('/swaps'))
			data = pending
				? [
						{
							id: 1,
							requester_wid: pending === 'open' ? '401' : '201',
							requester_name: pending === 'open' ? '测试请求人' : '测试正式成员',
							target_wid: pending === 'open' ? '' : '302',
							target_name: pending === 'open' ? '' : '测试实习乙',
							date_a: today,
							date_b: '2026-11-01',
							approval_a: '',
							approval_b: '',
							status: pending === 'open' ? 'open' : 'pending',
							reason: '一次性调整',
							can_review: pending !== 'open' && wid === '101',
							can_claim: pending === 'open' && wid === '201'
						}
					]
				: [];
		else if (url.pathname.endsWith('/audit')) data = [];
		else data = day;
		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({ success: true, data })
		});
	});
}

test('member calendar marks duty dates and hides management controls', async ({ page }) => {
	await setup(page, '201');
	await page.goto('/op/scheduler/');
	await expect(page.getByRole('heading', { name: '我的排班', exact: true })).toBeVisible();
	await expect(page.locator('.scheduler-duty-date').first()).toBeVisible();
	await expect(page.locator('#scheduler-calendar')).toBeHidden();
	await expect(page.locator('.flatpickr-calendar').first()).toBeVisible();
	await expect(page.locator('.scheduler-duty-date').first()).toHaveAttribute(
		'aria-label',
		`${today}，需要值班`
	);
	await expect(page.getByRole('heading', { name: '添加蹭班' })).toHaveCount(0);
	await expect(page.getByRole('heading', { name: '配置每周小组' })).toHaveCount(0);
	await expect(page.getByRole('link', { name: '换班请求大厅', exact: true })).toBeVisible();
	await page.screenshot({ path: 'test-results/scheduler-member-mobile.png', fullPage: true });
	await page.goto('/admin/scheduler/');
	await expect(page).toHaveURL(/\/op\/scheduler\/?$/);
	await expect(page.getByText('当天无管理权限')).toHaveCount(0);
	await expect(page.getByRole('link', { name: '排班管理与确认' })).toHaveCount(0);
	await expect(page.getByRole('button', { name: '保存每周小组' })).toHaveCount(0);
});

test('swap hall offers eligible requests without targeting another member', async ({ page }) => {
	await setup(page, '201', 'open');
	await page.goto('/op/scheduler/swaps/');
	await expect(page.getByRole('heading', { name: '换班请求大厅', exact: true })).toBeVisible();
	await expect(page.getByLabel('预期换班日期')).toBeVisible();
	await expect(page.getByLabel('换班对象工号')).toHaveCount(0);
	await expect(page.getByRole('button', { name: '确认换班', exact: true })).toHaveCount(0);
	const desired = new Intl.DateTimeFormat('sv-SE', { timeZone: 'Asia/Shanghai' }).format(
		new Date(new Date(`${today}T12:00:00+08:00`).getTime() + 86400000)
	);
	await page.getByLabel('我的原值班日期').selectOption(today);
	await page.getByLabel('预期换班日期').fill(desired);
	await page.getByLabel('预期换班日期').press('Tab');
	const create = page.waitForRequest(
		(r) => r.method() === 'POST' && r.url().endsWith('/scheduler/swaps')
	);
	await page.getByRole('button', { name: '提交换班申请' }).click();
	expect((await create).postDataJSON()).toEqual({ date_a: today, date_b: desired, reason: '' });
	const request = page.waitForRequest((r) => r.url().endsWith('/scheduler/swaps/review'));
	await expect(page.getByRole('button', { name: '接单', exact: true })).toBeVisible();
	await page.getByRole('button', { name: '接单', exact: true }).click();
	expect((await request).postDataJSON()).toEqual({ id: 1, action: 'claim' });
	await expect(page.getByText('已接单，等待双方组长确认')).toBeVisible();
	await page.screenshot({ path: 'test-results/scheduler-swap-hall-mobile.png', fullPage: true });
});

test('swap hall explains why a viewer cannot claim a request', async ({ page }) => {
	await setup(page, '900', 'open');
	await page.goto('/op/scheduler/swaps/');
	await expect(page.getByRole('button', { name: '接单', exact: true })).toBeVisible();
	await expect(page.getByRole('button', { name: '接单', exact: true })).toBeDisabled();
	await expect(page.getByText(/接单需在预期日期的已发布班次中值班/)).toBeVisible();
});

test('leader can preview a previous arrangement and confirm a relevant swap', async ({ page }) => {
	await setup(page, '101', true);
	let saves = 0;
	page.on('request', (request) => {
		if (request.url().includes('/scheduler/assignments')) saves++;
	});
	await page.goto('/admin/scheduler/');
	await expect(page.getByRole('heading', { name: '安排成员片区' })).toBeVisible();
	await expect(page.getByRole('button', { name: '保存每周小组' })).toHaveCount(0);
	await page.getByRole('button', { name: '沿用上次排班' }).click();
	await expect(page.getByText(/尚未提交/)).toBeVisible();
	expect(saves).toBe(0);
	await page.getByRole('button', { name: '随机安排实习成员' }).click();
	await expect(page.getByText(/已随机生成/)).toBeVisible();
	expect(saves).toBe(0);
	await expect(page.getByRole('heading', { name: '添加蹭班' })).toBeVisible();
	const request = page.waitForRequest((r) => r.url().endsWith('/scheduler/swaps/review'));
	const confirmation = page.getByRole('button', { name: '确认换班', exact: true });
	await confirmation.scrollIntoViewIfNeeded();
	const cardBox = await page.locator('.swap-requests .bx--tile').first().boundingBox();
	const buttonBox = await confirmation.boundingBox();
	expect(cardBox).not.toBeNull();
	expect(buttonBox).not.toBeNull();
	expect(buttonBox!.y).toBeGreaterThanOrEqual(cardBox!.y);
	expect(buttonBox!.y + buttonBox!.height).toBeLessThanOrEqual(cardBox!.y + cardBox!.height);
	await confirmation.click();
	expect((await request).postDataJSON()).toEqual({ id: 1, action: 'approve' });
	await page.screenshot({ path: 'test-results/scheduler-leader-mobile.png', fullPage: true });
});

test('chief can configure groups and role colors follow the supplied legend', async ({ page }) => {
	await setup(page, '900');
	await page.goto('/admin/scheduler/');
	await expect(page.getByRole('heading', { name: '配置每周小组' })).toBeVisible();
	await expect(page.getByRole('button', { name: '保存每周小组' })).toBeVisible();
	await expect(page.getByLabel('值班组长')).toHaveValue('101');
	const members = page.locator('.bx--multi-select');
	await expect(members.locator('input:not([type="checkbox"])')).toHaveCount(0);
	await members.getByRole('combobox').click();
	await expect(members.getByRole('option').filter({ hasText: '测试科长' })).toBeVisible();
	await members.getByRole('option').filter({ hasText: '测试科长' }).click();
	await expect(members.getByRole('combobox')).toBeFocused();
	await expect(page.locator('.assignment').filter({ hasText: '测试组长' })).toHaveCSS(
		'background-color',
		'rgb(237, 125, 49)'
	);
	await expect(page.locator('.assignment').filter({ hasText: '测试正式成员' })).toHaveCSS(
		'background-color',
		'rgb(255, 217, 102)'
	);
	await expect(page.locator('.assignment').filter({ hasText: '测试实习甲' })).toHaveCSS(
		'background-color',
		'rgb(255, 255, 255)'
	);
	await page.screenshot({ path: 'test-results/scheduler-chief-mobile.png', fullPage: true });
});
