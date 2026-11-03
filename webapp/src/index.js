import React from 'react';

const {isTwinPost} = require('./badge');
const {readTokenStatus, rotateAgentToken, revokeAgentToken, csrfHeaders} = require('./agent_token_api');
const pluginID = 'com.darksoftwarefactory.twinmate';
const apiURL = `/plugins/${pluginID}/api/config`;
const provenanceURL = `/plugins/${pluginID}/api/provenance`;

function endpointOrigin(value) {
    try { return new URL(value).origin; } catch { return ''; }
}

const styles = {
    backdrop: {position: 'fixed', inset: 0, background: 'rgba(0,0,0,.55)', zIndex: 9999, display: 'flex', alignItems: 'center', justifyContent: 'center'},
    panel: {width: 'min(520px, calc(100vw - 32px))', maxHeight: 'calc(100vh - 40px)', overflowY: 'auto', borderRadius: 12, background: 'var(--center-channel-bg, #fff)', color: 'var(--center-channel-color, #222)', padding: 24, boxShadow: '0 18px 50px #0004'},
    field: {display: 'block', width: '100%', boxSizing: 'border-box', border: '1px solid var(--center-channel-color-20, #ccc)', background: 'var(--center-channel-bg, #fff)', color: 'inherit', borderRadius: 6, padding: 10, marginTop: 6},
    label: {display: 'block', marginTop: 16, fontWeight: 600},
    secondary: {opacity: .75, fontSize: 13, lineHeight: 1.5},
    button: {padding: '9px 18px', borderRadius: 6, border: 0, background: 'var(--button-bg, #166de0)', color: 'var(--button-color, #fff)', cursor: 'pointer'},
};

function TwinmateSettingsRoot() {
    const [open, setOpen] = React.useState(false);
    const [endpoint, setEndpoint] = React.useState('');
    const [token, setToken] = React.useState('');
    const [enabled, setEnabled] = React.useState(false);
    const [hasToken, setHasToken] = React.useState(false);
    const [savedEndpoint, setSavedEndpoint] = React.useState('');
    const [removeToken, setRemoveToken] = React.useState(false);
    const [status, setStatus] = React.useState('');
    const [hasAgentToken, setHasAgentToken] = React.useState(false);
    const [issuedAgentToken, setIssuedAgentToken] = React.useState('');
    const [agentTokenStatus, setAgentTokenStatus] = React.useState('');

    React.useEffect(() => {
        const openSettings = () => { setOpen(true); setStatus(''); setIssuedAgentToken(''); setAgentTokenStatus(''); };
        window.addEventListener('twinmate:open', openSettings);
        return () => window.removeEventListener('twinmate:open', openSettings);
    }, []);

    React.useEffect(() => {
        if (!open) { return; }
        let alive = true;
        fetch(apiURL, {credentials: 'same-origin'}).then(async (response) => {
            if (!response.ok) { throw new Error('Unable to load settings'); }
            return response.json();
        }).then((data) => {
            if (alive) { setEndpoint(data.endpoint || ''); setSavedEndpoint(data.endpoint || ''); setToken(''); setRemoveToken(false); setEnabled(Boolean(data.enabled)); setHasToken(Boolean(data.has_token)); }
        }).catch((error) => { if (alive) { setStatus(error.message); } });
        return () => { alive = false; };
    }, [open]);

    React.useEffect(() => {
        if (!open) { return; }
        let alive = true;
        readTokenStatus().then((data) => { if (alive) { setHasAgentToken(Boolean(data.has_token)); } })
            .catch((error) => { if (alive) { setAgentTokenStatus(error.message); } });
        return () => { alive = false; };
    }, [open]);

    async function issueAgentToken() {
        if (hasAgentToken && !window.confirm('Rotate your agent token? The existing OpenClaw/MCP connection will stop working.')) { return; }
        setAgentTokenStatus('Generating token…');
        try {
            const generated = await rotateAgentToken();
            setIssuedAgentToken(generated);
            setHasAgentToken(true);
            setAgentTokenStatus('Copy this token now. Twinmate will never display it again.');
        } catch (error) { setAgentTokenStatus(error.message); }
    }

    async function removeAgentToken() {
        if (!window.confirm('Revoke access for your configured OpenClaw/MCP agent?')) { return; }
        try {
            await revokeAgentToken();
            setIssuedAgentToken('');
            setHasAgentToken(false);
            setAgentTokenStatus('Access token revoked.');
        } catch (error) { setAgentTokenStatus(error.message); }
    }

    async function save(event) {
        event.preventDefault();
        setStatus('Saving…');
        try {
            const response = await fetch(apiURL, {method: 'PUT', credentials: 'same-origin', headers: {'Content-Type': 'application/json', ...csrfHeaders()}, body: JSON.stringify({endpoint, token, enabled, remove_token: removeToken})});
            if (!response.ok) { throw new Error('Unable to save settings. Check the HTTPS endpoint.'); }
            setHasToken(!removeToken && (Boolean(token) || (hasToken && endpointOrigin(savedEndpoint) === endpointOrigin(endpoint))));
            setSavedEndpoint(endpoint);
            setRemoveToken(false);
            setToken('');
            setStatus('Saved. In any DM use /twinmate on to activate your twin there.');
        } catch (error) { setStatus(error.message); }
    }

    if (!open) { return null; }
    return React.createElement('div', {style: styles.backdrop, onClick: () => { setIssuedAgentToken(''); setOpen(false); }},
        React.createElement('div', {style: styles.panel, role: 'dialog', 'aria-modal': true, 'aria-label': 'Twinmate settings', onClick: (e) => e.stopPropagation()},
            React.createElement('div', {style: {display: 'flex', justifyContent: 'space-between', alignItems: 'center'}},
                React.createElement('h2', {style: {margin: 0}}, '🤖 Twinmate'),
                React.createElement('button', {type: 'button', onClick: () => { setIssuedAgentToken(''); setOpen(false); }, 'aria-label': 'Close'}, '×')),
            React.createElement('h3', null, 'OpenClaw / MCP agent access'),
            React.createElement('p', {style: styles.secondary}, 'Give your local agent a revocable Twinmate key to read Mattermost discussions by link and publish clearly marked AI messages in channels where you have permission. Do not share it or commit it to source control.'),
            React.createElement('div', {style: {display: 'flex', gap: 8, flexWrap: 'wrap'}},
                React.createElement('button', {type: 'button', style: styles.button, onClick: issueAgentToken}, hasAgentToken ? 'Rotate agent key' : 'Generate agent key'),
                hasAgentToken && React.createElement('button', {type: 'button', style: {...styles.button, background: '#a33434'}, onClick: removeAgentToken}, 'Revoke key')),
            issuedAgentToken && React.createElement('div', null,
                React.createElement('label', {style: styles.label}, 'One-time Twinmate agent key'),
                React.createElement('input', {style: styles.field, type: 'text', value: issuedAgentToken, readOnly: true, onFocus: (e) => e.target.select()}),
                React.createElement('button', {type: 'button', style: {...styles.button, marginTop: 8}, onClick: () => navigator.clipboard.writeText(issuedAgentToken)}, 'Copy token')),
            React.createElement('p', {style: styles.secondary, role: 'status'}, agentTokenStatus),
            React.createElement('hr', {style: {marginTop: 20, opacity: .2}}),
            React.createElement('h3', null, 'Automatic DM replies (optional)'),
            React.createElement('p', {style: styles.secondary}, 'Connect a reachable HTTPS agent to answer new DMs automatically. This is separate from the local agent access key above. Only DMs you explicitly enable can send messages to this endpoint.'),
            React.createElement('form', {onSubmit: save},
                React.createElement('label', {style: styles.label, htmlFor: 'twinmate-endpoint'}, 'Agent HTTPS endpoint'),
                React.createElement('input', {id: 'twinmate-endpoint', style: styles.field, type: 'url', placeholder: 'https://agent.example.com/events', value: endpoint, onChange: (e) => setEndpoint(e.target.value)}),
                React.createElement('label', {style: styles.label, htmlFor: 'twinmate-token'}, 'Agent bearer token'),
                React.createElement('input', {id: 'twinmate-token', style: styles.field, type: 'password', autoComplete: 'new-password', disabled: removeToken, placeholder: hasToken ? 'Leave blank to keep it on this host' : 'Optional', value: token, onChange: (e) => setToken(e.target.value)}),
                hasToken && React.createElement('label', {style: {...styles.label, fontWeight: 400}},
                    React.createElement('input', {type: 'checkbox', checked: removeToken, onChange: (e) => { setRemoveToken(e.target.checked); setToken(''); }}), ' Remove the saved token'),
                hasToken && savedEndpoint && endpointOrigin(savedEndpoint) !== endpointOrigin(endpoint) && React.createElement('p', {style: styles.secondary}, 'Changing the agent host clears the previously saved token. Enter a new token if the new endpoint needs one.'),
                React.createElement('label', {style: {...styles.label, fontWeight: 400}},
                    React.createElement('input', {type: 'checkbox', checked: enabled, onChange: (e) => setEnabled(e.target.checked)}), ' Enable my personal Twinmate connection'),
                React.createElement('p', {style: styles.secondary}, 'To enable in one conversation, send /twinmate on inside that DM. Use /twinmate off to stop.'),
                React.createElement('button', {style: styles.button, type: 'submit'}, 'Save connection'),
                React.createElement('p', {role: 'status', style: styles.secondary}, status))));
}

function TwinmateBadge({post}) {
    const [verified, setVerified] = React.useState(false);
    React.useEffect(() => {
        if (!isTwinPost(post) || !post.id) { setVerified(false); return; }
        let alive = true;
        let retry;
        const check = async (attempt) => {
            try {
                const response = await fetch(`${provenanceURL}?post_id=${encodeURIComponent(post.id)}`, {credentials: 'same-origin'});
                if (!response.ok) { throw new Error('Provenance unavailable'); }
                const result = await response.json();
                if (alive && result.verified) { setVerified(true); return; }
            } catch { /* Body label remains visible if the badge cannot be verified. */ }
            if (alive && attempt === 0) { retry = setTimeout(() => check(1), 650); }
        };
        setVerified(false);
        check(0);
        return () => { alive = false; clearTimeout(retry); };
    }, [post.id, Boolean(post.props?.twinmate?.generated)]);
    if (!verified || !isTwinPost(post)) { return null; }
    return React.createElement('span', {title: 'Sent by a personal AI twin, not typed by the person', style: {display: 'inline-flex', alignItems: 'center', border: '1px solid #8b5cf6', color: '#8b5cf6', fontWeight: 600, fontSize: 11, padding: '1px 6px', borderRadius: 9, marginLeft: 5}}, '✦ AI-двойник');
}

class TwinmatePlugin {
    initialize(registry) {
        registry.registerRootComponent(TwinmateSettingsRoot);
        registry.registerMainMenuAction('Twinmate · My digital twin', () => window.dispatchEvent(new Event('twinmate:open')));
        if (typeof registry.registerPostHeaderComponent === 'function') {
            registry.registerPostHeaderComponent(TwinmateBadge);
        }
    }
}

window.registerPlugin(pluginID, new TwinmatePlugin());
