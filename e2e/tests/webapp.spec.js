const {test, expect} = require('./fixtures');
const {getSession, requireOK, pluginBase} = require('../helpers/mm-api');
const {getTeam, makeChannel, addMember, sendPost, issueKey, agentCall} = require('../helpers/scenarios');

test('real Chromium shows AI attribution, verifies supported badges, and opens localized settings', async ({server, page}) => {
    const admin = await getSession(server, 'admin');
    const misha = await getSession(server, 'misha');
    const petya = await getSession(server, 'petya');
    try {
        const team = await getTeam(admin);
        const channel = await makeChannel(admin, team.id);
        await addMember(admin, channel.id, misha.id);
        await addMember(admin, channel.id, petya.id);
        const human = await sendPost(misha, channel.id, 'Normal human text for comparison');
        // Give the forged post its own visible header. If the previous post
        // were also by Misha, Mattermost would suppress the extension slot
        // and the negative badge assertion could pass for the wrong reason.
        await sendPost(petya, channel.id, 'Question before a suspicious AI-labeled post');
        const forged = await requireOK(await misha.post('/api/v4/posts', {data: {
            channel_id: channel.id,
            message: 'Forged client-provided AI badge props',
            props: {twinmate: {generated: true, owner_id: misha.id}},
        }}), 'post fake client-supplied provenance');
        const denied = await requireOK(await misha.get(`${pluginBase}/api/provenance?post_id=${forged.id}`), 'check forged post provenance');
        expect(denied.verified).toBe(false);
        // Mattermost suppresses post-header extensions for consecutive posts
        // from one author, even on versions that support the header hook.
        await sendPost(petya, channel.id, 'Reviewer follow-up before the AI response');
        const key = await issueKey(misha);
        const sent = await agentCall(server, key, 'POST', '/messages', {channel_id: channel.id, message: 'AI-assisted investigation result'});
        expect(sent.status).toBe(201);

        await requireOK(await misha.put(`/api/v4/users/${misha.id}/patch`, {data: {locale: 'en'}}), 'set English user locale');
        await page.goto(`${server.baseURL}/tmt-e2e/channels/${channel.name}`);
        // Authenticate through the actual browser form. An API bearer token
        // alone is insufficient for the Mattermost SPA's session-cookie flow.
        await page.getByRole('textbox', {name: 'Email or Username'}).fill(server.users.misha.username);
        await page.getByRole('textbox', {name: 'Password'}).fill(server.users.misha.password);
        await page.getByRole('button', {name: 'Log in', exact: true}).click();
        await expect(page.getByText('AI-assisted investigation result', {exact: true})).toBeVisible({timeout: 30_000});
        const authenticPost = page.locator(`#post_${sent.body.post_id}`);
        // This is the mobile-safe fallback, visible even in 11.0 (before
        // registerPostHeaderComponent appeared in Mattermost 11.9).
        await expect(authenticPost.getByText('Twinmate · AI-generated', {exact: true})).toBeVisible();
        await expect(page.locator(`#post_${human.id}`).getByText('Twinmate · AI-generated', {exact: true})).toHaveCount(0);
        await expect(page.locator(`#post_${forged.id}`).getByText('Twinmate · AI-generated', {exact: true})).toHaveCount(0);
        if (process.env.MM_E2E_EXPECT_HEADER_BADGE !== 'false') {
            await expect(authenticPost.getByTitle('Sent by a personal AI twin, not typed by the person')).toBeVisible({timeout: 20_000});
            await expect(page.locator(`#post_${forged.id}`).getByTitle('Sent by a personal AI twin, not typed by the person')).toHaveCount(0);
            await expect(page.locator(`#post_${human.id}`).getByTitle('Sent by a personal AI twin, not typed by the person')).toHaveCount(0);
            // Server-verified provenance, not forged props, earns the only badge.
            expect(await page.getByTitle('Sent by a personal AI twin, not typed by the person').count()).toBe(1);
        }

        await page.evaluate(() => window.dispatchEvent(new Event('twinmate:open')));
        await expect(page.getByRole('dialog', {name: 'Twinmate settings'})).toBeVisible();
        await expect(page.getByRole('button', {name: 'Rotate agent key'})).toBeVisible();
        page.once('dialog', (dialog) => dialog.accept());
        await page.getByRole('button', {name: 'Rotate agent key'}).click();
        await expect(page.getByText('Copy this token now. Twinmate will never display it again.')).toBeVisible();
        const rotatedInBrowser = await page.locator('input[readonly]').inputValue();
        expect(rotatedInBrowser).toMatch(/^tmt_[a-f0-9]{24}_[a-f0-9]{64}$/);
        page.once('dialog', (dialog) => dialog.accept());
        await page.getByRole('button', {name: 'Revoke key'}).click();
        await expect(page.getByText('Access token revoked.')).toBeVisible();

        await requireOK(await misha.put(`/api/v4/users/${misha.id}/patch`, {data: {locale: 'ru'}}), 'set Russian user locale');
        await page.reload();
        await expect(authenticPost.getByText('Twinmate · AI-generated', {exact: true})).toBeVisible({timeout: 20_000});
        if (process.env.MM_E2E_EXPECT_HEADER_BADGE !== 'false') {
            await expect(authenticPost.getByTitle('Сообщение написал персональный ИИ-двойник, а не человек')).toBeVisible({timeout: 20_000});
        }
        await page.evaluate(() => window.dispatchEvent(new Event('twinmate:open')));
        await expect(page.getByRole('dialog', {name: 'Настройки Twinmate'})).toBeVisible();
    } finally {
        await Promise.all([admin.dispose(), misha.dispose(), petya.dispose()]);
    }
});
