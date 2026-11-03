import React from 'react';

const {isTwinPost} = require('./badge');
const {readTokenStatus, rotateAgentToken, revokeAgentToken, csrfHeaders} = require('./agent_token_api');
const {t, viewerLocale} = require('./i18n');
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

function TwinmateSettingsRoot({locale}) {
    const tr = (key) => t(key, locale);
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
        if (hasAgentToken && !window.confirm(tr('token_rotate_confirm'))) { return; }
        setAgentTokenStatus(tr('token_rotating'));
        try {
            const generated = await rotateAgentToken();
            setIssuedAgentToken(generated);
            setHasAgentToken(true);
            setAgentTokenStatus(tr('token_created'));
        } catch (error) { setAgentTokenStatus(error.message); }
    }

    async function removeAgentToken() {
        if (!window.confirm(tr('token_revoke_confirm'))) { return; }
        try {
            await revokeAgentToken();
            setIssuedAgentToken('');
            setHasAgentToken(false);
            setAgentTokenStatus(tr('token_revoked'));
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
            setStatus(tr('saved'));
        } catch (error) { setStatus(error.message); }
    }

    if (!open) { return null; }
    return React.createElement('div', {style: styles.backdrop, onClick: () => { setIssuedAgentToken(''); setOpen(false); }},
        React.createElement('div', {style: styles.panel, role: 'dialog', 'aria-modal': true, 'aria-label': tr('settings_title'), onClick: (e) => e.stopPropagation()},
            React.createElement('div', {style: {display: 'flex', justifyContent: 'space-between', alignItems: 'center'}},
                React.createElement('h2', {style: {margin: 0}}, '🤖 Twinmate'),
                React.createElement('button', {type: 'button', onClick: () => { setIssuedAgentToken(''); setOpen(false); }, 'aria-label': tr('close')}, '×')),
            React.createElement('h3', null, tr('access_title')),
            React.createElement('p', {style: styles.secondary}, tr('access_intro')),
            React.createElement('div', {style: {display: 'flex', gap: 8, flexWrap: 'wrap'}},
                React.createElement('button', {type: 'button', style: styles.button, onClick: issueAgentToken}, tr(hasAgentToken ? 'rotate_key' : 'generate_key')),
                hasAgentToken && React.createElement('button', {type: 'button', style: {...styles.button, background: '#a33434'}, onClick: removeAgentToken}, tr('revoke_key'))),
            issuedAgentToken && React.createElement('div', null,
                React.createElement('label', {style: styles.label}, tr('one_time_key')),
                React.createElement('input', {style: styles.field, type: 'text', value: issuedAgentToken, readOnly: true, onFocus: (e) => e.target.select()}),
                React.createElement('button', {type: 'button', style: {...styles.button, marginTop: 8}, onClick: () => navigator.clipboard.writeText(issuedAgentToken)}, tr('copy'))),
            React.createElement('p', {style: styles.secondary, role: 'status'}, agentTokenStatus),
            React.createElement('hr', {style: {marginTop: 20, opacity: .2}}),
            React.createElement('h3', null, tr('automatic_title')),
            React.createElement('p', {style: styles.secondary}, tr('automatic_intro')),
            React.createElement('form', {onSubmit: save},
                React.createElement('label', {style: styles.label, htmlFor: 'twinmate-endpoint'}, tr('agent_endpoint')),
                React.createElement('input', {id: 'twinmate-endpoint', style: styles.field, type: 'url', placeholder: 'https://agent.example.com/events', value: endpoint, onChange: (e) => setEndpoint(e.target.value)}),
                React.createElement('label', {style: styles.label, htmlFor: 'twinmate-token'}, tr('bearer_token')),
                React.createElement('input', {id: 'twinmate-token', style: styles.field, type: 'password', autoComplete: 'new-password', disabled: removeToken, placeholder: hasToken ? 'Leave blank to keep it on this host' : 'Optional', value: token, onChange: (e) => setToken(e.target.value)}),
                hasToken && React.createElement('label', {style: {...styles.label, fontWeight: 400}},
                    React.createElement('input', {type: 'checkbox', checked: removeToken, onChange: (e) => { setRemoveToken(e.target.checked); setToken(''); }}), ` ${tr('remove_token')}`),
                hasToken && savedEndpoint && endpointOrigin(savedEndpoint) !== endpointOrigin(endpoint) && React.createElement('p', {style: styles.secondary}, 'Changing the agent host clears the previously saved token. Enter a new token if the new endpoint needs one.'),
                React.createElement('label', {style: {...styles.label, fontWeight: 400}},
                    React.createElement('input', {type: 'checkbox', checked: enabled, onChange: (e) => setEnabled(e.target.checked)}), ` ${tr('enable_connection')}`),
                React.createElement('p', {style: styles.secondary}, tr('dm_help')),
                React.createElement('button', {style: styles.button, type: 'submit'}, tr('save')),
                React.createElement('p', {role: 'status', style: styles.secondary}, status))));
}

function TwinmateBadge({post, locale}) {
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
    return React.createElement('span', {title: t('badge_title',locale), style: {display: 'inline-flex', alignItems: 'center', border: '1px solid #8b5cf6', color: '#8b5cf6', fontWeight: 600, fontSize: 11, padding: '1px 6px', borderRadius: 9, marginLeft: 5}}, `✦ ${t('badge',locale)}`);
}

class TwinmatePlugin {
    initialize(registry, store) {
        registry.registerRootComponent(() => React.createElement(TwinmateSettingsRoot,{locale:viewerLocale(store)}));
        registry.registerMainMenuAction(t('menu',viewerLocale(store)), () => window.dispatchEvent(new Event('twinmate:open')));
        if (typeof registry.registerPostHeaderComponent === 'function') {
            registry.registerPostHeaderComponent((props) => React.createElement(TwinmateBadge,{...props,locale:viewerLocale(store)}));
        }
    }
}

window.registerPlugin(pluginID, new TwinmatePlugin());
