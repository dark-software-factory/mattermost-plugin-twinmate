const {isTwinPost} = require('./badge');
const {test} = require('node:test');
const assert = require('node:assert/strict');

test('badge is visible only on generated Twinmate posts', () => {
    assert.equal(isTwinPost({props: {twinmate: {generated: true}}}), true);
    assert.equal(isTwinPost({props: {twinmate: {generated: false}}}), false);
    assert.equal(isTwinPost({props: {}}), false);
    assert.equal(isTwinPost(null), false);
});
