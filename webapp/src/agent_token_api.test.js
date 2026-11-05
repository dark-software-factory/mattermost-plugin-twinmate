const test = require('node:test');
const assert = require('node:assert/strict');
const {readTokenStatus, rotateAgentToken, revokeAgentToken, csrfHeaders} = require('./agent_token_api');

test('CSRF header uses the Mattermost MMCSRF cookie for mutating requests', () => {
    assert.deepEqual(csrfHeaders('session=one; MMCSRF=secure-token%2Bvalue; other=two'), {'X-CSRF-Token': 'secure-token+value'});
    assert.deepEqual(csrfHeaders('other=value'), {});
});

test('token status never exposes a secret', async () => {
    const status = await readTokenStatus(async (url, options) => {
        assert.match(url, /\/api\/agent\/token$/);
        assert.equal(options.credentials, 'same-origin');
        return {ok: true, json: async () => ({has_token: true})};
    });
    assert.deepEqual(status, {has_token: true});
});

test('token rotation requires JSON and retrieves the one-time secret', async () => {
    const token = await rotateAgentToken(async (_, options) => {
        assert.equal(options.method, 'POST');
        assert.equal(options.headers['Content-Type'], 'application/json');
        assert.equal(options.body, '{"allow_send":false,"channel_ids":[]}');
        return {ok: true, json: async () => ({token: 'tmt_secret'})};
    });
    assert.equal(token, 'tmt_secret');
});

test('token issuance sends requested immutable scopes', async () => {
    await rotateAgentToken(async (_, options) => {
        assert.deepEqual(JSON.parse(options.body), {allow_send: true, channel_ids: ['abcdefghijklmnopqrstuvwx12']});
        return {ok: true, json: async () => ({token: 'tmt_scoped'})};
    }, {allow_send: true, channel_ids: ['abcdefghijklmnopqrstuvwx12']});
});

test('session-authenticated rotate includes the real MMCSRF header', async () => {
    const oldDocument = global.document;
    global.document = {cookie: 'MMCSRF=the-session-csrf'};
    try {
        await rotateAgentToken(async (_, options) => {
            assert.equal(options.headers['X-CSRF-Token'], 'the-session-csrf');
            return {ok: true, json: async () => ({token: 'tmt_new'})};
        });
    } finally {
        global.document = oldDocument;
    }
});

test('token revoke uses the session-authenticated endpoint', async () => {
    let called = false;
    await revokeAgentToken(async (_, options) => {
        called = true;
        assert.equal(options.method, 'DELETE');
        assert.equal(options.credentials, 'same-origin');
        return {ok: true};
    });
    assert.equal(called, true);
});
