const {request} = require('@playwright/test');

const pluginID = 'com.darksoftwarefactory.twinmate';
const pluginBase = `/plugins/${pluginID}`;

async function getSession(server, person) {
    const credentials = server.users[person];
    if (!credentials) {throw new Error(`Unknown E2E test actor: ${person}`);}
    const context = await request.newContext({baseURL: server.baseURL});
    const login = await context.post('/api/v4/users/login', {data: {login_id: credentials.username, password: credentials.password}});
    if (!login.ok()) {
        const body = await login.text();
        await context.dispose();
        throw new Error(`Cannot log in E2E actor ${person}: HTTP ${login.status()} ${body}`);
    }
    const user = await login.json();
    const token = login.headers().token;
    if (!token) {throw new Error(`Mattermost login returned no session token for ${person}`);}
    const cookies = (await context.storageState()).cookies;
    const csrf = cookies.find((cookie) => cookie.name === 'MMCSRF')?.value;
    const headers = {Authorization: `Bearer ${token}`};
    if (csrf) {headers['X-CSRF-Token'] = csrf;}
    const authorized = async (method, path, options = {}) => context.fetch(path, {
        method,
        ...options,
        headers: {...headers, ...(options.headers || {})},
    });
    return {
        id: user.id,
        username: credentials.username,
        token,
        context,
        get: (path, options) => authorized('GET', path, options),
        post: (path, options) => authorized('POST', path, options),
        put: (path, options) => authorized('PUT', path, options),
        delete: (path, options) => authorized('DELETE', path, options),
        dispose: () => context.dispose(),
    };
}

async function requireOK(response, description) {
    if (!response.ok()) {throw new Error(`${description}: HTTP ${response.status()} ${await response.text()}`);}
    return response.json();
}

module.exports = {getSession, requireOK, pluginBase, pluginID};
