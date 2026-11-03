const {startServer} = require('./server');
const {clearState, saveState} = require('./state');

module.exports = async function setup() {
    clearState();
    if (process.env.MM_E2E_USE_EXISTING_SERVER) {
        throw new Error('Twinmate E2E deliberately uses a disposable Mattermost. Existing servers are never seeded.');
    }
    const server = await startServer();
    saveState({baseURL: server.baseURL, users: server.users});
    return async () => {
        clearState();
        await server.stop();
    };
};
