const {test: playwrightTest, expect} = require('@playwright/test');
const {loadState} = require('../helpers/state');

const test = playwrightTest.extend({
    server: async ({}, use) => {
        await use(loadState());
    },
    baseURL: async ({server}, use) => {
        await use(server.baseURL);
    },
});

module.exports = {test, expect};
