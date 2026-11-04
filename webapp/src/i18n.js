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
        settings_load_failed:'Unable to load Twinmate settings.', token_status_failed:'Unable to check agent access.',
        settings_save_failed:'Unable to save settings. Check the HTTPS endpoint.', token_action_failed:'Unable to change agent access.',
        saving:'Saving…', token_placeholder:'Leave blank to keep the saved token', optional:'Optional',
        host_changed:'Changing the agent host clears the saved token. Enter a new token if the new endpoint needs one.',
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
        settings_load_failed:'Не удалось загрузить настройки Twinmate.', token_status_failed:'Не удалось проверить доступ агента.',
        settings_save_failed:'Не удалось сохранить настройки. Проверьте HTTPS-адрес агента.', token_action_failed:'Не удалось изменить доступ агента.',
        saving:'Сохраняем…', token_placeholder:'Оставьте пустым, чтобы сохранить прежний токен', optional:'Необязательно',
        host_changed:'При смене адреса сервера прежний токен удаляется. Если новому агенту нужен токен, введите его.',
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

// Complete the settings copy for every explicitly supported locale. English
// fallback is reserved for unknown locales, not untranslated UI fragments.
Object.assign(messages.de, {
    access_intro:'Gib deinem lokalen Agenten einen widerrufbaren Twinmate-Schlüssel. Er kann verlinkte Nachrichten lesen und gekennzeichnete KI-Beiträge in erlaubten Kanälen senden. Halte den Schlüssel geheim.',
    automatic_intro:'Verbinde einen per HTTPS erreichbaren Agenten für automatische DM-Antworten. Nur ausdrücklich aktivierte Gespräche werden weitergegeben.',
    remove_token:'Gespeichertes Token entfernen',enable_connection:'Meine Twinmate-Verbindung aktivieren',
    dm_help:'Schreibe /twinmate on in einer DM, um Antworten zu aktivieren, und /twinmate off zum Deaktivieren.',
    saved:'Gespeichert. Aktiviere automatische Antworten in einer DM mit /twinmate on.',
    token_created:'Kopiere den Schlüssel jetzt. Twinmate zeigt ihn nie wieder an.',token_revoked:'Zugriffsschlüssel widerrufen.',token_rotating:'Schlüssel wird erstellt…',
    token_revoke_confirm:'Zugriff für deinen OpenClaw-/MCP-Agenten widerrufen?',token_rotate_confirm:'Schlüssel erneuern? Die bisherige Verbindung funktioniert dann nicht mehr.',
    settings_load_failed:'Twinmate-Einstellungen konnten nicht geladen werden.',token_status_failed:'Agentenzugriff konnte nicht geprüft werden.',
    settings_save_failed:'Einstellungen konnten nicht gespeichert werden. Prüfe den HTTPS-Endpunkt.',token_action_failed:'Agentenzugriff konnte nicht geändert werden.',
    saving:'Speichern…',token_placeholder:'Leer lassen, um das gespeicherte Token zu behalten',optional:'Optional',
    host_changed:'Beim Wechsel des Agentenhosts wird das gespeicherte Token gelöscht. Gib bei Bedarf ein neues Token ein.',
});
Object.assign(messages.fr, {
    access_intro:'Donnez à votre agent local une clé Twinmate révocable pour lire des messages liés et publier des messages IA signalés dans les canaux autorisés. Gardez cette clé secrète.',
    automatic_intro:'Connectez un agent HTTPS accessible pour répondre automatiquement aux messages privés. Seules les conversations explicitement activées sont transmises.',
    remove_token:'Supprimer le jeton enregistré',enable_connection:'Activer ma connexion Twinmate',
    dm_help:'Dans un message privé, envoyez /twinmate on pour activer les réponses et /twinmate off pour les arrêter.',
    saved:'Enregistré. Activez les réponses dans un message privé avec /twinmate on.',
    token_created:'Copiez cette clé maintenant. Twinmate ne l’affichera plus.',token_revoked:'Clé d’accès révoquée.',token_rotating:'Création de la clé…',
    token_revoke_confirm:'Révoquer l’accès de votre agent OpenClaw/MCP ?',token_rotate_confirm:'Renouveler la clé ? La connexion actuelle cessera de fonctionner.',
    settings_load_failed:'Impossible de charger les paramètres Twinmate.',token_status_failed:'Impossible de vérifier l’accès de l’agent.',
    settings_save_failed:'Impossible d’enregistrer. Vérifiez l’adresse HTTPS.',token_action_failed:'Impossible de modifier l’accès de l’agent.',
    saving:'Enregistrement…',token_placeholder:'Laissez vide pour conserver le jeton',optional:'Facultatif',
    host_changed:'Changer l’hôte de l’agent efface le jeton enregistré. Saisissez un nouveau jeton si nécessaire.',
});
Object.assign(messages.es, {
    access_intro:'Entrega a tu agente local una clave Twinmate revocable para leer mensajes enlazados y publicar mensajes de IA identificados en canales permitidos. Mantén la clave en secreto.',
    automatic_intro:'Conecta un agente HTTPS accesible para responder automáticamente a mensajes directos. Solo se compartirán los chats que actives expresamente.',
    remove_token:'Eliminar el token guardado',enable_connection:'Activar mi conexión Twinmate',
    dm_help:'Envía /twinmate on en un mensaje directo para activar respuestas y /twinmate off para detenerlas.',
    saved:'Guardado. Activa las respuestas en un chat directo con /twinmate on.',
    token_created:'Copia la clave ahora. Twinmate no volverá a mostrarla.',token_revoked:'Clave de acceso revocada.',token_rotating:'Generando clave…',
    token_revoke_confirm:'¿Revocar el acceso de tu agente OpenClaw/MCP?',token_rotate_confirm:'¿Cambiar la clave? La conexión anterior dejará de funcionar.',
    settings_load_failed:'No se pudo cargar la configuración de Twinmate.',token_status_failed:'No se pudo comprobar el acceso del agente.',
    settings_save_failed:'No se pudo guardar. Comprueba la dirección HTTPS.',token_action_failed:'No se pudo modificar el acceso del agente.',
    saving:'Guardando…',token_placeholder:'Déjalo vacío para conservar el token',optional:'Opcional',
    host_changed:'Al cambiar el host del agente se borra el token guardado. Introduce uno nuevo si hace falta.',
});
Object.assign(messages['pt-BR'], {
    access_intro:'Dê ao seu agente local uma chave Twinmate revogável para ler mensagens vinculadas e publicar mensagens de IA identificadas em canais permitidos. Mantenha a chave em segredo.',
    automatic_intro:'Conecte um agente HTTPS acessível para responder automaticamente às mensagens diretas. Somente as conversas que você ativar serão enviadas.',
    remove_token:'Remover token salvo',enable_connection:'Ativar minha conexão Twinmate',
    dm_help:'Envie /twinmate on em uma DM para ativar respostas e /twinmate off para parar.',
    saved:'Salvo. Ative as respostas em uma DM com /twinmate on.',
    token_created:'Copie a chave agora. O Twinmate não vai mostrá-la novamente.',token_revoked:'Chave de acesso revogada.',token_rotating:'Gerando chave…',
    token_revoke_confirm:'Revogar o acesso do seu agente OpenClaw/MCP?',token_rotate_confirm:'Trocar a chave? A conexão anterior deixará de funcionar.',
    settings_load_failed:'Não foi possível carregar as configurações do Twinmate.',token_status_failed:'Não foi possível verificar o acesso do agente.',
    settings_save_failed:'Não foi possível salvar. Confira o endereço HTTPS.',token_action_failed:'Não foi possível alterar o acesso do agente.',
    saving:'Salvando…',token_placeholder:'Deixe em branco para manter o token salvo',optional:'Opcional',
    host_changed:'Alterar o host do agente remove o token salvo. Informe outro token se necessário.',
});
Object.assign(messages.it, {
    access_intro:'Fornisci al tuo agente locale una chiave Twinmate revocabile per leggere messaggi collegati e pubblicare messaggi IA contrassegnati nei canali autorizzati. Mantienila segreta.',
    automatic_intro:'Collega un agente HTTPS raggiungibile per rispondere automaticamente ai messaggi privati. Verranno inoltrate solo le conversazioni che abiliti esplicitamente.',
    remove_token:'Rimuovi token salvato',enable_connection:'Attiva la mia connessione Twinmate',
    dm_help:'Invia /twinmate on in un DM per attivare le risposte e /twinmate off per fermarle.',
    saved:'Salvato. Attiva le risposte in un DM con /twinmate on.',
    token_created:'Copia ora la chiave. Twinmate non la mostrerà più.',token_revoked:'Chiave di accesso revocata.',token_rotating:'Generazione della chiave…',
    token_revoke_confirm:'Revocare l’accesso del tuo agente OpenClaw/MCP?',token_rotate_confirm:'Cambiare la chiave? La connessione precedente smetterà di funzionare.',
    settings_load_failed:'Impossibile caricare le impostazioni di Twinmate.',token_status_failed:'Impossibile verificare l’accesso dell’agente.',
    settings_save_failed:'Impossibile salvare. Controlla l’indirizzo HTTPS.',token_action_failed:'Impossibile modificare l’accesso dell’agente.',
    saving:'Salvataggio…',token_placeholder:'Lascia vuoto per mantenere il token salvato',optional:'Facoltativo',
    host_changed:'Cambiando l’host dell’agente il token salvato verrà cancellato. Inseriscine uno nuovo se necessario.',
});
Object.assign(messages.nl, {
    access_intro:'Geef je lokale agent een intrekbare Twinmate-sleutel om gelinkte berichten te lezen en gemarkeerde AI-berichten te plaatsen in toegestane kanalen. Houd de sleutel geheim.',
    automatic_intro:'Verbind een bereikbare HTTPS-agent voor automatische DM-antwoorden. Alleen gesprekken die je uitdrukkelijk inschakelt worden doorgestuurd.',
    remove_token:'Opgeslagen token verwijderen',enable_connection:'Mijn Twinmate-verbinding inschakelen',
    dm_help:'Stuur /twinmate on in een DM om antwoorden aan te zetten en /twinmate off om ze te stoppen.',
    saved:'Opgeslagen. Schakel antwoorden in een DM in met /twinmate on.',
    token_created:'Kopieer de sleutel nu. Twinmate toont deze niet opnieuw.',token_revoked:'Toegangssleutel ingetrokken.',token_rotating:'Sleutel aanmaken…',
    token_revoke_confirm:'Toegang voor je OpenClaw-/MCP-agent intrekken?',token_rotate_confirm:'Sleutel vernieuwen? De vorige verbinding werkt dan niet meer.',
    settings_load_failed:'Kan Twinmate-instellingen niet laden.',token_status_failed:'Kan agenttoegang niet controleren.',
    settings_save_failed:'Kan instellingen niet opslaan. Controleer het HTTPS-adres.',token_action_failed:'Kan agenttoegang niet wijzigen.',
    saving:'Opslaan…',token_placeholder:'Laat leeg om het opgeslagen token te behouden',optional:'Optioneel',
    host_changed:'Bij een andere agenthost wordt het opgeslagen token verwijderd. Voer indien nodig een nieuw token in.',
});
Object.assign(messages.ja, {
    access_intro:'ローカルエージェント用の取り消し可能な Twinmate キーを発行します。権限のあるチャンネルでリンク先のメッセージを読み、AI 生成と明示した投稿を送れます。キーは秘密にしてください。',
    automatic_intro:'HTTPS で接続できるエージェントを使い、DM に自動返信します。明示的に有効化した会話だけが転送されます。',
    remove_token:'保存したトークンを削除',enable_connection:'自分の Twinmate 接続を有効化',
    dm_help:'DM で /twinmate on を送ると自動返信を開始し、/twinmate off で停止します。',
    saved:'保存しました。DM で /twinmate on を送ると自動返信が有効になります。',
    token_created:'今すぐキーをコピーしてください。Twinmate に再表示されることはありません。',token_revoked:'アクセスキーを取り消しました。',token_rotating:'キーを生成中…',
    token_revoke_confirm:'OpenClaw/MCP エージェントのアクセスを取り消しますか？',token_rotate_confirm:'キーを再発行しますか？以前の接続は使えなくなります。',
    settings_load_failed:'Twinmate 設定を読み込めませんでした。',token_status_failed:'エージェントのアクセス状態を確認できませんでした。',
    settings_save_failed:'保存できませんでした。HTTPS アドレスを確認してください。',token_action_failed:'エージェントのアクセスを変更できませんでした。',
    saving:'保存中…',token_placeholder:'保存済みのトークンを維持するには空欄にしてください',optional:'任意',
    host_changed:'エージェントのホストを変更すると、保存済みトークンが削除されます。必要なら新しいトークンを入力してください。',
});
Object.assign(messages.ko, {
    access_intro:'로컬 에이전트에 취소 가능한 Twinmate 키를 발급하세요. 권한이 있는 채널에서 링크된 메시지를 읽고 AI 생성 표시가 있는 게시물을 보낼 수 있습니다. 키를 안전하게 보관하세요.',
    automatic_intro:'접속 가능한 HTTPS 에이전트를 연결해 DM에 자동 응답합니다. 명시적으로 활성화한 대화만 전달됩니다.',
    remove_token:'저장된 토큰 삭제',enable_connection:'내 Twinmate 연결 활성화',
    dm_help:'DM에서 /twinmate on을 보내면 자동 응답을 켜고 /twinmate off로 끌 수 있습니다.',
    saved:'저장했습니다. DM에서 /twinmate on을 보내 자동 응답을 활성화하세요.',
    token_created:'지금 키를 복사하세요. Twinmate는 다시 표시하지 않습니다.',token_revoked:'액세스 키가 폐기되었습니다.',token_rotating:'키 생성 중…',
    token_revoke_confirm:'OpenClaw/MCP 에이전트의 액세스를 취소할까요?',token_rotate_confirm:'키를 교체할까요? 기존 연결이 중단됩니다.',
    settings_load_failed:'Twinmate 설정을 불러올 수 없습니다.',token_status_failed:'에이전트 액세스를 확인할 수 없습니다.',
    settings_save_failed:'저장하지 못했습니다. HTTPS 주소를 확인하세요.',token_action_failed:'에이전트 액세스를 변경할 수 없습니다.',
    saving:'저장 중…',token_placeholder:'기존 토큰을 유지하려면 비워 두세요',optional:'선택 사항',
    host_changed:'에이전트 호스트를 바꾸면 저장된 토큰이 삭제됩니다. 필요한 경우 새 토큰을 입력하세요.',
});
Object.assign(messages['zh-CN'], {
    access_intro:'为本地代理生成可撤销的 Twinmate 密钥。代理可以读取链接的消息，并在你有权限的频道发送明确标记为 AI 生成的内容。请妥善保管密钥。',
    automatic_intro:'连接可通过 HTTPS 访问的代理，自动回复私信。只有你明确启用的对话才会转发。',
    remove_token:'删除已保存的令牌',enable_connection:'启用我的 Twinmate 连接',
    dm_help:'在私信中发送 /twinmate on 开启自动回复，发送 /twinmate off 关闭。',
    saved:'已保存。在私信中使用 /twinmate on 启用自动回复。',
    token_created:'请立即复制密钥。Twinmate 不会再次显示。',token_revoked:'访问密钥已撤销。',token_rotating:'正在生成密钥…',
    token_revoke_confirm:'撤销 OpenClaw/MCP 代理的访问权限吗？',token_rotate_confirm:'轮换密钥吗？旧连接将停止工作。',
    settings_load_failed:'无法加载 Twinmate 设置。',token_status_failed:'无法检查代理访问权限。',
    settings_save_failed:'保存失败，请检查 HTTPS 地址。',token_action_failed:'无法更改代理访问权限。',
    saving:'正在保存…',token_placeholder:'留空以保留已保存的令牌',optional:'可选',
    host_changed:'更改代理主机会清除已保存的令牌。如有需要，请输入新令牌。',
});
Object.assign(messages.uk, {
    access_intro:'Надайте локальному агенту відкличний ключ Twinmate, щоб читати повідомлення за посиланням і надсилати позначені ШІ-повідомлення в доступні канали. Зберігайте ключ у таємниці.',
    automatic_intro:'Підключіть доступного через HTTPS агента для автоматичних відповідей у приватних чатах. Передаватимуться лише явно ввімкнені розмови.',
    remove_token:'Видалити збережений токен',enable_connection:'Увімкнути моє підключення Twinmate',
    dm_help:'У приватному чаті надішліть /twinmate on, щоб увімкнути відповіді, та /twinmate off, щоб вимкнути.',
    saved:'Збережено. Увімкніть відповіді в приватному чаті командою /twinmate on.',
    token_created:'Скопіюйте ключ зараз. Twinmate більше не показуватиме його.',token_revoked:'Ключ доступу відкликано.',token_rotating:'Створюємо ключ…',
    token_revoke_confirm:'Відкликати доступ агента OpenClaw/MCP?',token_rotate_confirm:'Оновити ключ? Попереднє підключення перестане працювати.',
    settings_load_failed:'Не вдалося завантажити налаштування Twinmate.',token_status_failed:'Не вдалося перевірити доступ агента.',
    settings_save_failed:'Не вдалося зберегти. Перевірте адресу HTTPS.',token_action_failed:'Не вдалося змінити доступ агента.',
    saving:'Зберігаємо…',token_placeholder:'Залиште порожнім, щоб зберегти поточний токен',optional:'Необов’язково',
    host_changed:'Зміна хоста агента видаляє збережений токен. За потреби введіть новий.',
});

const scopeCopy = {
    en: ['Allow this agent to send messages as me', 'Limit to channel IDs (comma-separated)', 'Empty = all channels I can access'],
    ru: ['Разрешить агенту писать сообщения от моего имени', 'Ограничить ID каналов (через запятую)', 'Пусто = все доступные мне каналы'],
    de: ['Agent darf Nachrichten in meinem Namen senden', 'Auf Kanal-IDs beschränken (kommagetrennt)', 'Leer = alle für mich zugänglichen Kanäle'],
    fr: ['Autoriser cet agent à envoyer des messages en mon nom', 'Limiter aux ID des canaux (virgules)', 'Vide = tous mes canaux accessibles'],
    es: ['Permitir que el agente envíe mensajes en mi nombre', 'Limitar a IDs de canales (separadas por comas)', 'Vacío = todos mis canales accesibles'],
    'pt-BR': ['Permitir que o agente envie mensagens em meu nome', 'Limitar a IDs de canais (separados por vírgula)', 'Vazio = todos os canais a que tenho acesso'],
    it: ['Consenti all’agente di inviare messaggi a mio nome', 'Limita agli ID dei canali (separati da virgole)', 'Vuoto = tutti i canali a cui ho accesso'],
    nl: ['Agent mag namens mij berichten sturen', 'Beperk tot kanaal-ID’s (komma’s)', 'Leeg = alle kanalen waar ik toegang toe heb'],
    ja: ['エージェントが自分の名前で投稿することを許可', 'チャンネル ID を制限（カンマ区切り）', '空欄 = 自分がアクセスできるすべてのチャンネル'],
    ko: ['에이전트가 내 이름으로 메시지 보내기 허용', '채널 ID 제한 (쉼표로 구분)', '비워 두면 접근 가능한 모든 채널'],
    'zh-CN': ['允许代理以我的身份发送消息', '限制频道 ID（以逗号分隔）', '留空 = 我可访问的全部频道'],
    uk: ['Дозволити агенту писати повідомлення від мого імені', 'Обмежити ID каналів (через кому)', 'Порожньо = усі доступні мені канали'],
};
for (const [locale, [agent_send_scope, agent_channel_scope, agent_channel_scope_hint]] of Object.entries(scopeCopy)) {
    Object.assign(messages[locale], {agent_send_scope, agent_channel_scope, agent_channel_scope_hint});
}

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

module.exports = {t, viewerLocale, messageCatalog: messages};
