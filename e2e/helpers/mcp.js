const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const https = require('node:https');
const http = require('node:http');
const readline = require('node:readline');
const {spawn, execFileSync} = require('node:child_process');

// The real local MCP client insists on HTTPS. This disposable, certificate-
// validated reverse proxy forwards to Testcontainers' private HTTP endpoint.
// No TLS verification is bypassed and no production endpoint is touched.
async function trustedProxy(target) {
    const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'twinmate-e2e-tls-'));
    const key = path.join(directory, 'key.pem');
    const cert = path.join(directory, 'cert.pem');
    let server;
    try {
        execFileSync('openssl', ['req', '-x509', '-newkey', 'rsa:2048', '-nodes', '-sha256',
            '-keyout', key, '-out', cert, '-days', '1', '-subj', '/CN=localhost',
            '-addext', 'subjectAltName=DNS:localhost,IP:127.0.0.1'], {stdio: 'ignore'});
        const origin = new URL(target);
        server = https.createServer({key: fs.readFileSync(key), cert: fs.readFileSync(cert)}, (request, response) => {
            const upstream = http.request(new URL(request.url, origin), {
                method: request.method,
                headers: {...request.headers, host: origin.host},
            }, (result) => {
                response.writeHead(result.statusCode, result.headers);
                result.pipe(response);
            });
            upstream.on('error', (error) => {response.writeHead(502); response.end(error.message);});
            request.pipe(upstream);
        });
        await new Promise((resolve, reject) => {
            server.once('error', reject);
            server.listen(0, '127.0.0.1', resolve);
        });
        return {
            url: `https://127.0.0.1:${server.address().port}`,
            cert,
            stop: async () => {
                await new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve()));
                fs.rmSync(directory, {recursive: true, force: true});
            },
        };
    } catch (error) {
        if (server?.listening) {server.close();}
        fs.rmSync(directory, {recursive: true, force: true});
        throw error;
    }
}

async function startMCP(url, cert, token) {
    const binary = path.resolve(__dirname, '../../dist/twinmate-mcp-linux-amd64');
    if (!fs.existsSync(binary)) {throw new Error(`Missing compiled MCP executable: ${binary}`);}
    const child = spawn(binary, [], {
        stdio: ['pipe', 'pipe', 'pipe'],
        env: {...process.env, TWINMATE_MATTERMOST_URL: url, TWINMATE_AGENT_TOKEN: token, SSL_CERT_FILE: cert},
    });
    const pending = new Map();
    let counter = 0;
    let stderr = '';
    child.stderr.on('data', (data) => {stderr += data.toString().slice(0, 1000);});
    child.on('error', (error) => {for (const reject of pending.values()) {reject(error);} pending.clear();});
    child.on('exit', (code) => {
        for (const reject of pending.values()) {reject(new Error(`MCP exited ${code}: ${stderr}`));}
        pending.clear();
    });
    const lines = readline.createInterface({input: child.stdout});
    lines.on('line', (line) => {
        let response;
        try {response = JSON.parse(line);} catch {return;}
        const done = pending.get(response.id);
        if (done) {pending.delete(response.id); done(null, response);}
    });

    function rpc(method, params = {}) {
        const id = ++counter;
        return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {pending.delete(id); reject(new Error(`MCP ${method} timed out: ${stderr}`));}, 12_000);
            pending.set(id, (error, response) => {
                clearTimeout(timeout);
                if (error) {reject(error);}
                else if (response.error) {reject(new Error(JSON.stringify(response.error)));}
                else {resolve(response.result);}
            });
            child.stdin.write(`${JSON.stringify({jsonrpc: '2.0', id, method, params})}\n`);
        });
    }

    try {
        await rpc('initialize', {protocolVersion: '2025-11-25', capabilities: {}, clientInfo: {name: 'twinmate-e2e', version: '1.0'}});
        child.stdin.write(`${JSON.stringify({jsonrpc: '2.0', method: 'notifications/initialized'})}\n`);
    } catch (error) {
        child.kill();
        throw error;
    }
    return {
        list: () => rpc('tools/list'),
        call: (name, args) => rpc('tools/call', {name, arguments: args}),
        stop: async () => {
            lines.close();
            if (child.exitCode !== null) {return;}
            const ended = new Promise((resolve) => child.once('exit', resolve));
            child.stdin.end();
            await Promise.race([ended, new Promise((resolve) => setTimeout(() => {child.kill(); resolve();}, 1500))]);
        },
    };
}

module.exports = {trustedProxy, startMCP};
