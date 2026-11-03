const {test, expect} = require('./fixtures');
const {getSession, requireOK} = require('../helpers/mm-api');
const {getTeam, makeChannel, addMember, sendPost, issueKey, agentCall} = require('../helpers/scenarios');

test('real Chromium renders verified AI badge in English and Russian, and opens settings', async ({server, page, context}) => {
    const admin = await getSession(server, 'admin');
    const misha = await getSession(server, 'misha');
    try {
        const team = await getTeam(admin);
        const channel = await makeChannel(admin, team.id);
        await addMember(admin, channel.id, misha.id);
        await sendPost(misha, channel.id, 'Normal human text for comparison');
        const key = await issueKey(misha);
        const sent = await agentCall(server, key, 'POST', '/messages', {channel_id: channel.id, message: 'AI-assisted investigation result'});
        expect(sent.status).toBe(201);

        // Session token supplied by the real Mattermost login endpoint.
        await context.addCookies([{name: 'MMAUTHTOKEN', value: misha.token, url: server.baseURL, httpOnly: true, sameSite: 'Lax'}]);
        await requireOK(await misha.put(`/api/v4/users/${misha.id}/patch`, {data: {locale: 'en'}}), 'set English user locale');
        await page.goto(`${server.baseURL}/tmt-e2e/channels/${channel.name}`);
        await expect(page.getByText('AI-assisted investigation result')).toBeVisible({timeout: 30_000});
        await expect(page.getByText('AI twin', {exact: true})).toBeVisible({timeout: 20_000});
        // The unverified human post must not receive a second badge.
        expect(await page.getByText('AI twin', {exact: true}).count()).toBe(1);

        await page.evaluate(() => window.dispatchEvent(new Event('twinmate:open')));
        await expect(page.getByRole('dialog', {name: 'Twinmate settings'})).toBeVisible();
        await expect(page.getByRole('button', {name: 'Rotate agent key'})).toBeVisible();

        await requireOK(await misha.put(`/api/v4/users/${misha.id}/patch`, {data: {locale: 'ru'}}), 'set Russian user locale');
        await page.reload();
        await expect(page.getByText('ИИ-двойник', {exact: true})).toBeVisible({timeout: 20_000});
        await page.evaluate(() => window.dispatchEvent(new Event('twinmate:open')));
        await expect(page.getByRole('dialog', {name: 'Настройки Twinmate'})).toBeVisible();
    } finally {
        await Promise.all([admin.dispose(), misha.dispose()]);
    }
});
