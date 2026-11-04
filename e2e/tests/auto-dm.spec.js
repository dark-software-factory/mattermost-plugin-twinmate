const {test, expect} = require('./fixtures');
const {getSession, requireOK, pluginBase} = require('../helpers/mm-api');
const {sendPost, getTeam, issueKey, agentCall} = require('../helpers/scenarios');

test('real one-to-one DM opt-in invokes the HTTPS agent, publishes a verified reply and stops after opt-out', async ({server, request}, testInfo) => {
    const [misha, petya, outsider] = await Promise.all(['misha', 'petya', 'outsider'].map((name) => getSession(server, name)));
    try {
        const team = await getTeam(misha);
        const dm = await requireOK(await misha.post('/api/v4/channels/direct', {data: [misha.id, petya.id]}), 'create direct message');
        const endpoint = 'https://twinmate-agent:8443/events';
        const saved = await misha.put(`${pluginBase}/api/config`, {
            data: {endpoint, token: 'disposable-e2e-bearer', enabled: true, remove_token: false},
        });
        expect(saved.status(), await saved.text()).toBe(200);

        const status = await requireOK(await misha.post('/api/v4/commands/execute', {
            data: {team_id: team.id, channel_id: dm.id, command: '/twinmate status'},
        }), 'read initial DM opt-in status');
        if (testInfo.retry === 0) {
            // On a fresh server, automatic replies must be disabled by default.
            expect(status.text).toContain('**off**');
        }
        // Playwright retries reuse the global disposable Mattermost server.
        // A previous attempt might have failed after enabling this DM.
        const reset = await requireOK(await misha.post('/api/v4/commands/execute', {
            data: {team_id: team.id, channel_id: dm.id, command: '/twinmate off'},
        }), 'reset DM opt-in before this attempt');
        expect(reset.text).toContain('now off');

        const before = await requireOK(await request.get(`${server.agentStatusURL}/`), 'check fake-agent initial events');
        await sendPost(petya, dm.id, 'Private content before opt-in');
        await new Promise((resolve) => setTimeout(resolve, 1_000));
        const stillDisabled = await requireOK(await request.get(`${server.agentStatusURL}/`), 'check initial DM opt-out');
        expect(stillDisabled.events).toHaveLength(before.events.length);
        const optIn = await misha.post('/api/v4/commands/execute', {
            data: {team_id: team.id, channel_id: dm.id, command: '/twinmate on'},
        });
        expect(optIn.status(), await optIn.text()).toBe(200);
        const visible = await requireOK(await petya.get(`/api/v4/channels/${dm.id}/posts?per_page=30`), 'read DM disclosure');
        expect(Object.values(visible.posts).some((post) => post.user_id === misha.id && post.message.includes('Twinmate notice') && post.message.includes('outside Mattermost'))).toBe(true);

        const trigger = await sendPost(petya, dm.id, 'Please investigate the broken migration');
        const outsiderKey = await issueKey(outsider);
        for (const resource of [`/posts/get?reference=${trigger.id}`, `/posts/list?reference=${trigger.id}&direction=before&limit=10`]) {
            const attempted = await agentCall(server, outsiderKey, 'GET', resource);
            expect(attempted.status).toBe(403);
            expect(JSON.stringify(attempted.body)).not.toContain(trigger.message);
            expect(JSON.stringify(attempted.body)).not.toContain('Private content before opt-in');
        }
        await expect.poll(async () => {
            const payload = await requireOK(await request.get(`${server.agentStatusURL}/`), 'poll fake agent');
            return payload.events.length;
        }, {timeout: 25_000}).toBe(before.events.length + 1);

        const received = await requireOK(await request.get(`${server.agentStatusURL}/`), 'inspect agent event');
        const last = received.events.at(-1);
        expect(last.event).toMatchObject({owner_id: misha.id, sender_id: petya.id, channel_id: dm.id, message_id: trigger.id, text: trigger.message});
        expect(last.eventID).toBe(`${trigger.id}:${misha.id}`);
        expect(last.bearer).toBe('Bearer disposable-e2e-bearer');

        let answer;
        await expect.poll(async () => {
            const posts = await requireOK(await misha.get(`/api/v4/channels/${dm.id}/posts?per_page=30`), 'read DM replies');
            answer = Object.values(posts.posts).find((post) => post.user_id === misha.id && post.props?.twinmate?.source_post_id === trigger.id);
            return Boolean(answer);
        }, {timeout: 25_000}).toBe(true);
        expect(answer.message).toContain('Automated answer for: Please investigate');
        const proof = await requireOK(await petya.get(`${pluginBase}/api/provenance?post_id=${answer.id}`), 'verify auto-DM provenance');
        expect(proof.verified).toBe(true);

        const optOut = await misha.post('/api/v4/commands/execute', {
            data: {team_id: team.id, channel_id: dm.id, command: '/twinmate off'},
        });
        expect(optOut.status(), await optOut.text()).toBe(200);
        await sendPost(petya, dm.id, 'Do not forward this post to my agent');
        // The processor is asynchronous; allow a short observation window.
        await new Promise((resolve) => setTimeout(resolve, 2_500));
        const after = await requireOK(await request.get(`${server.agentStatusURL}/`), 'inspect after opt-out');
        expect(after.events).toHaveLength(before.events.length + 1);
    } finally {
        await Promise.all([misha.dispose(), petya.dispose(), outsider.dispose()]);
    }
});
