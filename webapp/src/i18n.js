// The post body uses one language-neutral AI provenance marker because it is
// stored once for readers in every locale. Web UI labels use the viewer's
// Mattermost profile locale instead.
const messages = {
    en: {
        badge: 'AI twin', badge_title: 'Sent by a personal AI twin, not typed by the person',
        menu: 'Twinmate · My digital twin', settings_title: 'Twinmate settings',
        access_title: 'OpenClaw / MCP agent access',
        access_intro: 'Give your local agent a revocable Twinmate key to read linked Mattermost discussions and publish clearly marked AI messages where you have permission. Keep it secret.',
        generate_key: 'Generate agent key', rotate_key: 'Rotate agent key', revoke_key: 'Revoke key',
        one_time_key: 'One-time Twinmate agent key', copy: 'Copy token',
        automatic_title: 'Automatic DM replies (optional)',
        automatic_intro: 'Connect a reachable HTTPS agent to reply to new DMs automatically. Only DMs you explicitly enable can send messages to this endpoint.',
        agent_endpoint: 'Agent HTTPS endpoint', bearer_token: 'Agent bearer token',
        remove_token: 'Remove saved token', enable_connection: 'Enable my personal Twinmate connection',
        dm_help: 'To enable in one DM, send /twinmate on there. Use /twinmate off to stop.',
        save: 'Save connection', saved: 'Saved. Use /twinmate on in a DM to activate automatic replies.',
        token_created: 'Copy this token now. Twinmate will never display it again.', token_revoked: 'Access token revoked.',
        token_rotating: 'Generating token…',
        token_revoke_confirm: 'Revoke access for your configured OpenClaw/MCP agent?',
        token_rotate_confirm: 'Rotate your agent token? The previous connection will stop working.',
        close: 'Close',
    },
    ru: {
        badge: 'ИИ-двойник', badge_title: 'Сообщение написал персональный ИИ-двойник, а не человек',
        menu: 'Twinmate · Мой цифровой двойник', settings_title: 'Настройки Twinmate',
        access_title: 'Доступ OpenClaw / MCP',
        access_intro: 'Создайте отзывной ключ для локального агента. Он сможет читать обсуждения по ссылке и публиковать сообщения с отметкой ИИ от вашего имени в доступных каналах. Храните ключ в секрете.',
        generate_key: 'Создать ключ агента', rotate_key: 'Перевыпустить ключ', revoke_key: 'Отозвать ключ',
        one_time_key: 'Ключ агента (показывается один раз)', copy: 'Скопировать ключ',
        automatic_title: 'Автоответы в личных сообщениях (необязательно)',
        automatic_intro: 'Подключите доступного по HTTPS агента для автоответов в личке. Сообщения передаются только из диалогов, где вы явно включили эту функцию.',
        agent_endpoint: 'HTTPS-адрес агента', bearer_token: 'Bearer-токен агента',
        remove_token: 'Удалить сохранённый токен', enable_connection: 'Включить подключение Twinmate',
        dm_help: 'Чтобы включить автоответы в личке, отправьте /twinmate on. Для остановки — /twinmate off.',
        save: 'Сохранить подключение', saved: 'Сохранено. Включите автоответы командой /twinmate on в личке.',
        token_created: 'Скопируйте ключ сейчас: Twinmate больше его не покажет.', token_revoked: 'Ключ доступа отозван.',
        token_rotating: 'Создаём ключ…',
        token_revoke_confirm: 'Отозвать доступ вашего OpenClaw/MCP-агента?',
        token_rotate_confirm: 'Перевыпустить ключ? Старое подключение перестанет работать.',
        close: 'Закрыть',
    },
    de: {badge:'KI-Zwilling',badge_title:'Von einem persönlichen KI-Zwilling verfasst',menu:'Twinmate · Mein digitaler Zwilling',settings_title:'Twinmate-Einstellungen',access_title:'OpenClaw / MCP-Zugriff',generate_key:'Agentenschlüssel erstellen',rotate_key:'Schlüssel erneuern',revoke_key:'Schlüssel widerrufen',one_time_key:'Einmaliger Agentenschlüssel',copy:'Schlüssel kopieren',automatic_title:'Automatische Direktnachrichten (optional)',agent_endpoint:'HTTPS-Endpunkt des Agenten',bearer_token:'Bearer-Token',save:'Verbindung speichern',close:'Schließen'},
    fr: {badge:'Double IA',badge_title:'Message envoyé par un double IA personnel',menu:'Twinmate · Mon double numérique',settings_title:'Paramètres Twinmate',access_title:'Accès OpenClaw / MCP',generate_key:'Créer une clé agent',rotate_key:'Renouveler la clé',revoke_key:'Révoquer la clé',one_time_key:'Clé agent à usage unique',copy:'Copier la clé',automatic_title:'Réponses automatiques en privé (facultatif)',agent_endpoint:'Adresse HTTPS de l’agent',bearer_token:'Jeton Bearer',save:'Enregistrer',close:'Fermer'},
    es: {badge:'Gemelo de IA',badge_title:'Mensaje enviado por un gemelo de IA personal',menu:'Twinmate · Mi gemelo digital',settings_title:'Configuración de Twinmate',access_title:'Acceso OpenClaw / MCP',generate_key:'Crear clave del agente',rotate_key:'Rotar clave',revoke_key:'Revocar clave',one_time_key:'Clave del agente (una sola vez)',copy:'Copiar clave',automatic_title:'Respuestas automáticas por DM (opcional)',agent_endpoint:'Endpoint HTTPS del agente',bearer_token:'Token Bearer',save:'Guardar conexión',close:'Cerrar'},
    'pt-BR': {badge:'Gêmeo de IA',badge_title:'Mensagem enviada por um gêmeo de IA pessoal',menu:'Twinmate · Meu gêmeo digital',settings_title:'Configurações do Twinmate',access_title:'Acesso OpenClaw / MCP',generate_key:'Gerar chave do agente',rotate_key:'Trocar chave',revoke_key:'Revogar chave',one_time_key:'Chave de acesso única',copy:'Copiar chave',automatic_title:'Respostas automáticas em DMs (opcional)',agent_endpoint:'Endpoint HTTPS do agente',bearer_token:'Token Bearer',save:'Salvar conexão',close:'Fechar'},
    it: {badge:'Gemello IA',badge_title:'Messaggio inviato da un gemello IA personale',menu:'Twinmate · Il mio gemello digitale',settings_title:'Impostazioni Twinmate',access_title:'Accesso OpenClaw / MCP',generate_key:'Crea chiave agente',rotate_key:'Ruota chiave',revoke_key:'Revoca chiave',one_time_key:'Chiave agente monouso',copy:'Copia chiave',automatic_title:'Risposte automatiche DM (opzionale)',agent_endpoint:'Endpoint HTTPS agente',bearer_token:'Token Bearer',save:'Salva connessione',close:'Chiudi'},
    nl: {badge:'AI-tweeling',badge_title:'Bericht van een persoonlijke AI-tweeling',menu:'Twinmate · Mijn digitale tweeling',settings_title:'Twinmate-instellingen',access_title:'OpenClaw / MCP-toegang',generate_key:'Agentsleutel maken',rotate_key:'Sleutel vernieuwen',revoke_key:'Sleutel intrekken',one_time_key:'Eenmalige agentsleutel',copy:'Sleutel kopiëren',automatic_title:'Automatische DM-antwoorden (optioneel)',agent_endpoint:'HTTPS-eindpunt agent',bearer_token:'Bearer-token',save:'Verbinding opslaan',close:'Sluiten'},
    ja: {badge:'AI分身',badge_title:'本人ではなく個人のAI分身が送信したメッセージ',menu:'Twinmate · 自分のデジタル分身',settings_title:'Twinmate 設定',access_title:'OpenClaw / MCP 接続',generate_key:'エージェントキーを発行',rotate_key:'キーを再発行',revoke_key:'キーを無効化',one_time_key:'一度だけ表示されるキー',copy:'キーをコピー',automatic_title:'DM 自動返信（任意）',agent_endpoint:'エージェント HTTPS エンドポイント',bearer_token:'Bearer トークン',save:'接続を保存',close:'閉じる'},
    ko: {badge:'AI 분신',badge_title:'본인이 아닌 개인 AI 분신이 보낸 메시지',menu:'Twinmate · 내 디지털 트윈',settings_title:'Twinmate 설정',access_title:'OpenClaw / MCP 액세스',generate_key:'에이전트 키 생성',rotate_key:'키 교체',revoke_key:'키 폐기',one_time_key:'한 번만 표시되는 키',copy:'키 복사',automatic_title:'DM 자동 응답(선택)',agent_endpoint:'에이전트 HTTPS 주소',bearer_token:'Bearer 토큰',save:'연결 저장',close:'닫기'},
    'zh-CN': {badge:'AI 分身',badge_title:'由个人 AI 分身而非本人发送',menu:'Twinmate · 我的数字分身',settings_title:'Twinmate 设置',access_title:'OpenClaw / MCP 访问',generate_key:'生成代理密钥',rotate_key:'轮换密钥',revoke_key:'撤销密钥',one_time_key:'仅显示一次的代理密钥',copy:'复制密钥',automatic_title:'私信自动回复（可选）',agent_endpoint:'代理 HTTPS 地址',bearer_token:'Bearer 令牌',save:'保存连接',close:'关闭'},
    uk: {badge:'ШІ-двійник',badge_title:'Повідомлення надіслав персональний ШІ-двійник',menu:'Twinmate · Мій цифровий двійник',settings_title:'Налаштування Twinmate',access_title:'Доступ OpenClaw / MCP',generate_key:'Створити ключ агента',rotate_key:'Оновити ключ',revoke_key:'Відкликати ключ',one_time_key:'Одноразовий ключ агента',copy:'Скопіювати ключ',automatic_title:'Автовідповіді в особистих (необов’язково)',agent_endpoint:'HTTPS-адреса агента',bearer_token:'Bearer-токен',save:'Зберегти підключення',close:'Закрити'},
};

function normalize(locale) {
    const raw = String(locale || 'en').replace('_', '-').toLowerCase();
    if (raw.startsWith('pt')) { return 'pt-BR'; }
    if (raw.startsWith('zh')) { return 'zh-CN'; }
    const base = raw.split('-')[0];
    return messages[base] ? base : 'en';
}
function t(key, locale) { return messages[normalize(locale)][key] || messages.en[key] || key; }
function viewerLocale(store) {
    try {
        const state = store.getState();
        const users = state.entities.users;
        return users.profiles[users.currentUserId]?.locale || state.entities.general?.config?.DefaultClientLocale || 'en';
    } catch { return 'en'; }
}

module.exports = {t, viewerLocale};
