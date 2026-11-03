const {defineConfig, devices} = require('@playwright/test');

module.exports = defineConfig({
    testDir: './tests',
    testMatch: '*.spec.js',
    globalSetup: './helpers/setup.js',
    globalTimeout: 12 * 60 * 1000,
    timeout: 45_000,
    fullyParallel: false,
    workers: 1,
    retries: process.env.CI ? 1 : 0,
    forbidOnly: Boolean(process.env.CI),
    outputDir: 'test-results',
    reporter: [['list'], ['junit', {outputFile: 'results/junit.xml'}]],
    use: {trace: 'retain-on-failure', screenshot: 'only-on-failure'},
    projects: [{name: 'chromium', use: {...devices['Desktop Chrome']}}],
});
