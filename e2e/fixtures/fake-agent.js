// Runs only in the disposable E2E Docker network. No external LLM is called.
const fs = require('node:fs');
const http = require('node:http');
const https = require('node:https');

const events = [];
https.createServer({
    key: fs.readFileSync('/opt/twinmate-test/key.pem'),
    cert: fs.readFileSync('/opt/twinmate-test/cert.pem'),
}, (req, res) => {
    if (req.url !== '/events' || req.method !== 'POST') {res.writeHead(404).end(); return;}
    let raw = '';
    req.on('data', (chunk) => {raw += chunk;});
    req.on('end', () => {
        const event = JSON.parse(raw);
        events.push({event, bearer: req.headers.authorization, eventID: req.headers['x-twinmate-event-id']});
        res.writeHead(200, {'Content-Type': 'application/json'});
        res.end(JSON.stringify({reply: `Automated answer for: ${event.text}`}));
    });
}).listen(8443, '0.0.0.0', () => console.log('Twinmate test agent ready'));

http.createServer((req, res) => {
    res.writeHead(200, {'Content-Type': 'application/json'});
    res.end(JSON.stringify({events}));
}).listen(8080, '0.0.0.0');
