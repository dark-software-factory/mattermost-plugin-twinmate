const {test} = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const {mcpBinaryForPlatform} = require('./mcp');

for (const [platform, arch, filename] of [
    ['linux', 'x64', 'twinmate-mcp-linux-amd64'],
    ['darwin', 'x64', 'twinmate-mcp-darwin-amd64'],
    ['darwin', 'arm64', 'twinmate-mcp-darwin-arm64'],
]) {
    test(`MCP E2E launches the executable built for ${platform}/${arch}`, () => {
        assert.equal(path.basename(mcpBinaryForPlatform(platform, arch)), filename);
    });
}

test('MCP E2E fails explicitly on an unsupported host instead of launching an incompatible binary', () => {
    assert.throws(() => mcpBinaryForPlatform('win32', 'x64'), /Unsupported E2E MCP host/);
});
