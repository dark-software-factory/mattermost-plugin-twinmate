const tokenURL = '/plugins/com.darksoftwarefactory.twinmate/api/agent/token';

// Mattermost 11's strict plugin CSRF protection requires the cookie's
// one-time-per-session value on any cookie-authenticated write request.
function csrfHeaders(cookieString = typeof document !== 'undefined' ? document.cookie : '') {
    for (const part of cookieString.split(';')) {
        const cookie = part.trim();
        if (cookie.startsWith('MMCSRF=')) {
            return {'X-CSRF-Token': decodeURIComponent(cookie.slice('MMCSRF='.length))};
        }
    }
    return {};
}

async function readTokenStatus(fetcher = fetch) {
    const response = await fetcher(tokenURL, {credentials: 'same-origin'});
    if (!response.ok) { throw new Error('Unable to load agent access status'); }
    return response.json();
}

async function rotateAgentToken(fetcher = fetch) {
    const response = await fetcher(tokenURL, {method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json', ...csrfHeaders()}, body: '{}'});
    if (!response.ok) { throw new Error('Unable to create Twinmate agent token'); }
    const {token} = await response.json();
    if (!token) { throw new Error('Token not returned by Twinmate'); }
    return token;
}

async function revokeAgentToken(fetcher = fetch) {
    const response = await fetcher(tokenURL, {method: 'DELETE', credentials: 'same-origin', headers: {'Content-Type': 'application/json', ...csrfHeaders()}, body: '{}'});
    if (!response.ok) { throw new Error('Unable to revoke Twinmate agent token'); }
}

module.exports = {readTokenStatus, rotateAgentToken, revokeAgentToken, csrfHeaders};
