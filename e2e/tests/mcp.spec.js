const {test, expect} = require('./fixtures');
const {getSession, requireOK} = require('../helpers/mm-api');
const {getTeam, makeChannel, addMember, sendPost, issueKey} = require('../helpers/scenarios');
const {trustedProxy, startMCP} = require('../helpers/mcp');

test('compiled local MCP agent reads Mattermost, pages history and publishes a marked reply over verified TLS', async ({server}) => {
    const admin = await getSession(server, 'admin');
    const owner = await getSession(server, 'misha');
    let proxy;
    let mcp;
    try {
        const team = await getTeam(admin);
        const channel = await makeChannel(admin, team.id);
        await addMember(admin, channel.id, owner.id);
        const first = await sendPost(owner, channel.id, 'root-cause investigation');
        const second = await sendPost(owner, channel.id, 'checking deployment pipeline');
        const key = await issueKey(owner);

        proxy = await trustedProxy(server.baseURL);
        mcp = await startMCP(proxy.url, proxy.cert, key);
        const tools = await mcp.list();
        expect(tools.tools.map((tool) => tool.name)).toEqual(expect.arrayContaining([
            'twinmate_get_post', 'twinmate_list_posts', 'twinmate_send_message',
        ]));

        const read = await mcp.call('twinmate_get_post', {reference: first.id});
        expect(read.isError).not.toBe(true);
        expect(JSON.stringify(read)).toContain('root-cause investigation');

        const history = await mcp.call('twinmate_list_posts', {reference: second.id, direction: 'before', limit: 10});
        expect(history.isError).not.toBe(true);
        expect(JSON.stringify(history)).toContain(first.id);

        const sent = await mcp.call('twinmate_send_message', {reply_to_post_id: first.id, message: 'Fixed the pipeline race.'});
        expect(sent.isError).not.toBe(true);
        const serialized = JSON.stringify(sent);
        const postID = serialized.match(/[a-z0-9]{26}/)?.[0];
        expect(postID, serialized).toBeTruthy();
        const post = await requireOK(await owner.get(`/api/v4/posts/${postID}`), 'read MCP-published post');
        expect(post.user_id).toBe(owner.id);
        expect(post.root_id).toBe(first.id);
        expect(post.message).toContain('Twinmate · AI-generated');
        expect(post.props.twinmate.generated).toBe(true);
    } finally {
        if (mcp) {await mcp.stop();}
        if (proxy) {await proxy.stop();}
        await Promise.all([admin.dispose(), owner.dispose()]);
    }
});
