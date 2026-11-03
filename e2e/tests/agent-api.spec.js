const {test, expect} = require('./fixtures');
const {getSession, requireOK, pluginBase} = require('../helpers/mm-api');
const {makeChannel, addMember, sendPost, issueKey, agentCall, getTeam} = require('../helpers/scenarios');

test('real owner key can read, page and publish marked replies while outsiders are denied', async ({server}) => {
    const [admin, misha, petya, outsider] = await Promise.all(
        ['admin', 'misha', 'petya', 'outsider'].map((name) => getSession(server, name)),
    );
    try {
        const team = await getTeam(admin);
        const channel = await makeChannel(admin, team.id, 'P');
        await addMember(admin, channel.id, misha.id);
        await addMember(admin, channel.id, petya.id);

        const first = await sendPost(misha, channel.id, 'investigation begins');
        const middle = await sendPost(petya, channel.id, 'deploy produced an error');
        const last = await sendPost(misha, channel.id, 'checking the logs');

        const ownerKey = await issueKey(misha);
        const outsiderKey = await issueKey(outsider);
        const reference = `${server.baseURL}/tmt-e2e/pl/${middle.id}`;
        const read = await agentCall(server, ownerKey, 'GET', `/posts/get?reference=${encodeURIComponent(reference)}`);
        expect(read.status).toBe(200);
        expect(read.body).toMatchObject({id: middle.id, channel_id: channel.id, user_id: petya.id, message: middle.message});

        const before = await agentCall(server, ownerKey, 'GET', `/posts/list?reference=${middle.id}&direction=before&limit=50`);
        expect(before.status).toBe(200);
        expect(before.body.posts.some((post) => post.id === first.id)).toBeTruthy();
        const after = await agentCall(server, ownerKey, 'GET', `/posts/list?reference=${middle.id}&direction=after&limit=50`);
        expect(after.status).toBe(200);
        expect(after.body.posts.some((post) => post.id === last.id)).toBeTruthy();

        const page = await agentCall(server, ownerKey, 'GET', `/posts/list?reference=${last.id}&direction=before&limit=1`);
        expect(page.status).toBe(200);
        expect(page.body.posts).toHaveLength(1);
        expect(page.body.has_more).toBe(true);
        expect(page.body.next_cursor).toBe(page.body.posts[0].id);

        const forbidden = await agentCall(server, outsiderKey, 'GET', `/posts/get?reference=${middle.id}`);
        expect(forbidden.status).toBe(403);
        expect(JSON.stringify(forbidden.body)).not.toContain(middle.message);

        const sent = await agentCall(server, ownerKey, 'POST', '/messages', {reply_to_post_id: middle.id, message: 'Root cause: race in deployment.'});
        expect(sent.status).toBe(201);
        const created = await requireOK(await misha.get(`/api/v4/posts/${sent.body.post_id}`), 'fetch AI post');
        expect(created).toMatchObject({user_id: misha.id, channel_id: channel.id, root_id: middle.id});
        expect(created.message).toContain('Twinmate · AI-generated');
        expect(created.props.twinmate).toMatchObject({generated: true, owner_id: misha.id});
        const proven = await requireOK(await petya.get(`${pluginBase}/api/provenance?post_id=${created.id}`), 'verify AI provenance');
        expect(proven.verified).toBe(true);

        const illegal = await agentCall(server, outsiderKey, 'POST', '/messages', {channel_id: channel.id, message: 'I should not be here'});
        expect(illegal.status).toBe(403);
        const spoof = await agentCall(server, ownerKey, 'POST', '/messages', {channel_id: channel.id, owner_id: outsider.id, message: 'spoof'});
        expect(spoof.status).toBe(400);
        const wrongLink = await agentCall(server, ownerKey, 'GET', `/posts/get?reference=${encodeURIComponent('https://another.host/tmt-e2e/pl/' + middle.id)}`);
        expect(wrongLink.status).toBe(400);
    } finally {
        await Promise.all([admin, misha, petya, outsider].map((session) => session.dispose()));
    }
});

test('token status, CSRF session, rotation and revocation use real Mattermost auth', async ({server}) => {
    const user = await getSession(server, 'petya');
    try {
        const anonymous = await agentCall(server, '', 'GET', '/posts/get?reference=invalid');
        expect(anonymous.status).toBe(401);
        const status = await requireOK(await user.get(`${pluginBase}/api/agent/token`), 'token status');
        expect(typeof status.has_token).toBe('boolean');

        const first = await issueKey(user);
        const second = await issueKey(user);
        expect(second).not.toBe(first);
        const oldToken = await agentCall(server, first, 'GET', '/posts/get?reference=invalid');
        expect(oldToken.status).toBe(401);

        const revoke = await user.delete(`${pluginBase}/api/agent/token`, {data: {}});
        expect(revoke.status()).toBe(200);
        const revoked = await agentCall(server, second, 'GET', '/posts/get?reference=invalid');
        expect(revoked.status).toBe(401);
        const statusAfter = await requireOK(await user.get(`${pluginBase}/api/agent/token`), 'token status after revocation');
        expect(statusAfter).toEqual({has_token: false});
    } finally {await user.dispose();}
});

test('thread scope returns only replies to the selected root', async ({server}) => {
    const [admin, misha, petya] = await Promise.all(['admin', 'misha', 'petya'].map((name) => getSession(server, name)));
    try {
        const team = await getTeam(admin);
        const channel = await makeChannel(admin, team.id, 'O');
        await addMember(admin, channel.id, misha.id);
        await addMember(admin, channel.id, petya.id);
        const root = await sendPost(misha, channel.id, 'root of incident');
        const reply1 = await sendPost(petya, channel.id, 'reply one', root.id);
        const reply2 = await sendPost(misha, channel.id, 'reply two', root.id);
        const unrelated = await sendPost(petya, channel.id, 'unrelated root');
        const key = await issueKey(misha);
        const result = await agentCall(server, key, 'GET', `/posts/list?reference=${root.id}&direction=after&scope=thread&limit=20`);
        expect(result.status).toBe(200);
        const ids = result.body.posts.map((post) => post.id);
        expect(ids).toContain(reply1.id);
        expect(ids).toContain(reply2.id);
        expect(ids).not.toContain(unrelated.id);
    } finally {await Promise.all([admin, misha, petya].map((session) => session.dispose()));}
});
