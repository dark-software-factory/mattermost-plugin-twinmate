const {test, expect} = require('./fixtures');
const {getSession, pluginID} = require('../helpers/mm-api');

test('fresh Mattermost installs and activates the exact built Twinmate archive', async ({request, server}) => {
    const ping = await request.get(`${server.baseURL}/api/v4/system/ping`);
    expect(ping.ok()).toBeTruthy();
    const admin = await getSession(server, 'admin');
    try {
        const response = await admin.get('/api/v4/plugins');
        expect(response.status()).toBe(200);
        const plugins = await response.json();
        expect(plugins.active.some((entry) => entry.id === pluginID)).toBeTruthy();
    } finally {
        await admin.dispose();
    }
});
