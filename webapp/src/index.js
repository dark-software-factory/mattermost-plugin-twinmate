import React from 'react';

const {isTwinPost} = require('./badge');
const pluginID = 'com.darksoftwarefactory.twinmate';
const apiURL = `/plugins/${pluginID}/api/config`;

const styles = {
    backdrop: {position: 'fixed', inset: 0, background: 'rgba(0,0,0,.55)', zIndex: 9999, display: 'flex', alignItems: 'center', justifyContent: 'center'},
    panel: {width: 'min(520px, calc(100vw - 32px))', borderRadius: 12, background: 'var(--center-channel-bg, #fff)', color: 'var(--center-channel-color, #222)', padding: 24, boxShadow: '0 18px 50px #0004'},
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
    const [status, setStatus] = React.useState('');

    React.useEffect(() => {
        const openSettings = () => { setOpen(true); setStatus(''); };
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
            if (alive) { setEndpoint(data.endpoint || ''); setToken(''); setEnabled(Boolean(data.enabled)); setHasToken(Boolean(data.has_token)); }
        }).catch((error) => { if (alive) { setStatus(error.message); } });
        return () => { alive = false; };
    }, [open]);

    async function save(event) {
        event.preventDefault();
        setStatus('Saving…');
        try {
            const response = await fetch(apiURL, {method: 'PUT', credentials: 'same-origin', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({endpoint, token, enabled})});
            if (!response.ok) { throw new Error('Unable to save settings. Check the HTTPS endpoint.'); }
            setHasToken(hasToken || Boolean(token));
            setToken('');
            setStatus('Saved. In any DM use /twinmate on to activate your twin there.');
        } catch (error) { setStatus(error.message); }
    }

    if (!open) { return null; }
    return React.createElement('div', {style: styles.backdrop, onClick: () => setOpen(false)},
        React.createElement('div', {style: styles.panel, role: 'dialog', 'aria-modal': true, 'aria-label': 'Twinmate settings', onClick: (e) => e.stopPropagation()},
            React.createElement('div', {style: {display: 'flex', justifyContent: 'space-between', alignItems: 'center'}},
                React.createElement('h2', {style: {margin: 0}}, '🤖 Twinmate'),
                React.createElement('button', {type: 'button', onClick: () => setOpen(false), 'aria-label': 'Close'}, '×')),
            React.createElement('p', {style: styles.secondary}, 'Connect your personal HTTPS agent. Only DMs you explicitly enable can send new messages to it. Your external agent receives the message text and participant IDs.'),
            React.createElement('form', {onSubmit: save},
                React.createElement('label', {style: styles.label, htmlFor: 'twinmate-endpoint'}, 'Agent HTTPS endpoint'),
                React.createElement('input', {id: 'twinmate-endpoint', style: styles.field, type: 'url', placeholder: 'https://agent.example.com/events', value: endpoint, onChange: (e) => setEndpoint(e.target.value)}),
                React.createElement('label', {style: styles.label, htmlFor: 'twinmate-token'}, 'Agent bearer token'),
                React.createElement('input', {id: 'twinmate-token', style: styles.field, type: 'password', autoComplete: 'new-password', placeholder: hasToken ? 'Token saved — leave blank to keep it' : 'Optional', value: token, onChange: (e) => setToken(e.target.value)}),
                React.createElement('label', {style: {...styles.label, fontWeight: 400}},
                    React.createElement('input', {type: 'checkbox', checked: enabled, onChange: (e) => setEnabled(e.target.checked)}), ' Enable my personal Twinmate connection'),
                React.createElement('p', {style: styles.secondary}, 'To enable in one conversation, send /twinmate on inside that DM. Use /twinmate off to stop.'),
                React.createElement('button', {style: styles.button, type: 'submit'}, 'Save connection'),
                React.createElement('p', {role: 'status', style: styles.secondary}, status))));
}

function TwinmateBadge({post}) {
    if (!isTwinPost(post)) { return null; }
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
