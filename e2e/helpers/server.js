const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const {execFileSync} = require('node:child_process');
const {GenericContainer, Network, Wait} = require('testcontainers');
const {PostgreSqlContainer} = require('@testcontainers/postgresql');

const repoRoot = path.resolve(__dirname, '../..');
const bundle = path.join(repoRoot, 'dist', 'com.darksoftwarefactory.twinmate-0.1.0.tar.gz');
const pluginID = 'com.darksoftwarefactory.twinmate';
const password = 'Twinmate!2026';

const users = {
    admin: {username: 'tmt_admin', email: 'tmt_admin@example.test', password, admin: true},
    misha: {username: 'tmt_misha', email: 'tmt_misha@example.test', password},
    petya: {username: 'tmt_petya', email: 'tmt_petya@example.test', password},
    outsider: {username: 'tmt_outsider', email: 'tmt_outsider@example.test', password},
};

async function startServer() {
    if (!fs.existsSync(bundle)) {
        throw new Error(`Built Twinmate archive missing at ${bundle}. Run make bundle first.`);
    }

    let network;
    let postgres;
    let mattermost;
    let fakeAgent;
    let certificateDir;
    let logFile;
    const logDir = path.join(repoRoot, 'e2e', 'logs');
    fs.mkdirSync(logDir, {recursive: true});

    async function stop() {
        for (const container of [mattermost, fakeAgent, postgres, network]) {
            if (!container) {continue;}
            try {await container.stop();} catch (err) {console.error('[e2e] cleanup:', err.message);}
        }
        await new Promise((resolve) => logFile ? logFile.end(resolve) : resolve());
        if (certificateDir) {fs.rmSync(certificateDir, {recursive: true, force: true});}
    }
    async function exec(args) {
        const result = await mattermost.exec(args);
        if (result.exitCode !== 0) {
            throw new Error(`Mattermost mmctl failed (${result.exitCode}): ${args.slice(0, 4).join(' ')}\n${result.output}`);
        }
        return result.output;
    }

    try {
        network = await new Network().start();
        postgres = await new PostgreSqlContainer('postgres:15').
            withDatabase('twinmate').withUsername('twinmate').withPassword('twinmate').
            withNetwork(network).withNetworkAliases('postgres').
            withStartupTimeout(240_000).start();

        certificateDir = fs.mkdtempSync(path.join(os.tmpdir(), 'twinmate-e2e-agent-'));
        const cert = path.join(certificateDir, 'cert.pem');
        const key = path.join(certificateDir, 'key.pem');
        execFileSync('openssl', ['req', '-x509', '-newkey', 'rsa:2048', '-nodes', '-sha256',
            '-keyout', key, '-out', cert, '-days', '1', '-subj', '/CN=twinmate-agent',
            '-addext', 'subjectAltName=DNS:twinmate-agent'], {stdio: 'ignore'});

        fakeAgent = await new GenericContainer('node:24-alpine').
            withNetwork(network).withNetworkAliases('twinmate-agent').
            withExposedPorts(8080).
            withCopyFilesToContainer([
                {source: path.join(repoRoot, 'e2e/fixtures/fake-agent.js'), target: '/opt/twinmate-test/agent.js'},
                {source: cert, target: '/opt/twinmate-test/cert.pem'},
                {source: key, target: '/opt/twinmate-test/key.pem'},
            ]).
            withCommand(['node', '/opt/twinmate-test/agent.js']).
            withWaitStrategy(Wait.forLogMessage('Twinmate test agent ready')).
            withStartupTimeout(120_000).start();

        mattermost = await new GenericContainer(process.env.MM_E2E_IMAGE || 'mattermost/mattermost-team-edition:release-11.0').
            withEnvironment({
                MM_SQLSETTINGS_DRIVERNAME: 'postgres',
                MM_SQLSETTINGS_DATASOURCE: 'postgres://twinmate:twinmate@postgres:5432/twinmate?sslmode=disable',
                MM_SERVICESETTINGS_ENABLELOCALMODE: 'true',
                MM_PLUGINSETTINGS_ENABLEUPLOADS: 'true',
                MM_PLUGINSETTINGS_AUTOMATICPREPACKAGEDPLUGINS: 'false',
                MM_PLUGINSETTINGS_REQUIREPLUGINSIGNATURE: 'false',
                MM_SERVICESETTINGS_ENABLEONBOARDINGFLOW: 'false',
                MM_SERVICESETTINGS_ENABLEDESKTOPLANDINGPAGE: 'false',
                MM_TEAMSETTINGS_ENABLEOPEN: 'true',
                MM_EMAILSETTINGS_ENABLESIGNUPWITHEMAIL: 'true',
                SSL_CERT_FILE: '/tmp/twinmate-e2e-agent-cert.pem',
            }).
            withExposedPorts(8065).
            withNetwork(network).withNetworkAliases('mattermost').
            withWaitStrategy(Wait.forLogMessage('Server is listening on')).
            withCopyFilesToContainer([{source: cert, target: '/tmp/twinmate-e2e-agent-cert.pem'}]).
            withStartupTimeout(240_000).
            withLogConsumer((stream) => {
                logFile = fs.createWriteStream(path.join(logDir, 'mattermost.log'), {flags: 'w'});
                stream.on('data', (chunk) => logFile.write(chunk));
            }).
            start();

        const baseURL = `http://${mattermost.getHost()}:${mattermost.getMappedPort(8065)}`;
        await exec(['mmctl', '--local', 'config', 'set', 'ServiceSettings.SiteURL', baseURL]);

        for (const person of Object.values(users)) {
            const args = ['mmctl', '--local', 'user', 'create', '--email', person.email,
                '--username', person.username, '--password', person.password, '--email-verified'];
            if (person.admin) {args.push('--system-admin');}
            await exec(args);
        }

        await exec(['mmctl', '--local', 'team', 'create', '--name', 'tmt-e2e', '--display-name', 'Twinmate E2E']);
        for (const person of Object.values(users)) {
            await exec(['mmctl', '--local', 'team', 'users', 'add', 'tmt-e2e', person.username]);
        }

        await mattermost.copyFilesToContainer([{source: bundle, target: '/tmp/twinmate-e2e.tar.gz'}]);
        await exec(['mmctl', '--local', 'plugin', 'add', '/tmp/twinmate-e2e.tar.gz']);
        await exec(['mmctl', '--local', 'plugin', 'enable', pluginID]);
        // This toggle is enabled only on the disposable E2E server, to allow
        // HTTPS callbacks into its private, isolated Testcontainers network.
        await exec(['mmctl', '--local', 'config', 'set', `PluginSettings.Plugins.${pluginID}.AllowPrivateEndpoints`, 'true']);

        // Verify activation through the Mattermost API in installation.spec.js.
        // Avoid depending on mmctl's JSON envelope, which changes between releases.
        await exec(['mmctl', '--local', 'plugin', 'list']);

        const agentStatusURL = `http://${fakeAgent.getHost()}:${fakeAgent.getMappedPort(8080)}`;
        return {baseURL, agentStatusURL, users, stop};
    } catch (error) {
        await stop();
        throw error;
    }
}

module.exports = {startServer};
