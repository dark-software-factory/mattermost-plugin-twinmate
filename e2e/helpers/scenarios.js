const {randomBytes} = require('node:crypto');
const {request} = require('@playwright/test');
const {requireOK, pluginBase} = require('./mm-api');

const unique = () => `tmt-${randomBytes(5).toString('hex')}`;

async function getTeam(admin) {
    return requireOK(await admin.get('/api/v4/teams/name/tmt-e2e'), 'load E2E team');
}

async function makeChannel(admin, teamID, type = 'O') {
    const name = unique();
    return requireOK(await admin.post('/api/v4/channels', {
        data: {team_id: teamID, name, display_name: `Twinmate ${name}`, type},
    }), 'create E2E channel');
}

async function addMember(admin, channelID, userID) {
    return requireOK(await admin.post(`/api/v4/channels/${channelID}/members`, {data: {user_id: userID}}), 'join E2E channel');
}

async function sendPost(actor, channelID, message, rootID = '') {
    return requireOK(await actor.post('/api/v4/posts', {data: {channel_id: channelID, message, ...(rootID ? {root_id: rootID} : {})}}), 'create E2E post');
}

async function issueKey(owner, options = {allow_send: true}) {
    const response = await owner.post(`${pluginBase}/api/agent/token`, {data: options});
    const issued = await requireOK(response, 'issue per-user Twinmate agent key');
    if (!issued.token?.startsWith('tmt_')) {throw new Error('Mattermost returned no usable Twinmate key');}
    return issued.token;
}

async function agentCall(server, token, method, path, body) {
    const context = await request.newContext({baseURL: server.baseURL});
    try {
        const response = await context.fetch(`${pluginBase}/api/agent${path}`, {
            method,
            headers: {
                ...(token ? {'X-Twinmate-Token': token} : {}),
                ...(body !== undefined ? {'Content-Type': 'application/json'} : {}),
            },
            ...(body !== undefined ? {data: body} : {}),
        });
        const text = await response.text();
        let result;
        try {result = JSON.parse(text);} catch {result = {raw: text};}
        return {status: response.status(), body: result};
    } finally {
        await context.dispose();
    }
}

module.exports = {getTeam, makeChannel, addMember, sendPost, issueKey, agentCall};
