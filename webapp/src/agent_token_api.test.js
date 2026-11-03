const test = require('node:test');
const assert = require('node:assert/strict');
const {readTokenStatus, rotateAgentToken, revokeAgentToken} = require('./agent_token_api');

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
        assert.equal(options.body, '{}');
        return {ok: true, json: async () => ({token: 'tmt_secret'})};
    });
    assert.equal(token, 'tmt_secret');
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
