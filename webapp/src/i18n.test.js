const test = require('node:test');
const assert = require('node:assert/strict');
const {t, viewerLocale} = require('./i18n');

test('AI twin badge follows major Mattermost display locales', () => {
    assert.equal(t('badge', 'en'), 'AI twin');
    assert.equal(t('badge', 'ru-RU'), 'ИИ-двойник');
    assert.equal(t('badge', 'de'), 'KI-Zwilling');
    assert.equal(t('badge', 'fr'), 'Double IA');
    assert.equal(t('badge', 'es'), 'Gemelo de IA');
    assert.equal(t('badge', 'pt-BR'), 'Gêmeo de IA');
    assert.equal(t('badge', 'ja'), 'AI分身');
    assert.equal(t('badge', 'ko'), 'AI 분신');
    assert.equal(t('badge', 'zh-CN'), 'AI 分身');
    assert.equal(t('badge', 'uk'), 'ШІ-двійник');
    assert.equal(t('badge', 'it'), 'Gemello IA');
    assert.equal(t('badge', 'nl'), 'AI-tweeling');
});

test('unknown locales and untranslated keys fall back to English', () => {
    assert.equal(t('badge', 'xx-YY'), 'AI twin');
    assert.equal(t('settings_title', 'xx-YY'), 'Twinmate settings');
});

test('locale comes from the Mattermost user profile, not browser language', () => {
    const store = {getState: () => ({entities: {users: {currentUserId: 'misha', profiles: {misha: {locale: 'ru'}}}}})};
    assert.equal(viewerLocale(store), 'ru');
    assert.equal(viewerLocale(null), 'en');
});
