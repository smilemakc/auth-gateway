
import React, { createContext, useState, useContext, ReactNode, useEffect } from 'react';

type Language = 'ru' | 'en';

interface LanguageContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  t: (key: string) => string;
}

const translations: Record<Language, Record<string, string>> = {
  ru: {
    // Common
    'common.save': 'Сохранить',
    'common.saving': 'Сохранение...',
    'common.saved': 'Сохранено!',
    'common.cancel': 'Отмена',
    'common.delete': 'Удалить',
    'common.edit': 'Редактировать',
    'common.create': 'Создать',
    'common.back': 'Назад',
    'common.loading': 'Загрузка...',
    'common.actions': 'Действия',
    'common.search': 'Поиск...',
    'common.status': 'Статус',
    'common.created': 'Создано',
    'common.yes': 'Да',
    'common.no': 'Нет',
    'common.confirm_delete': 'Вы уверены, что хотите удалить этот элемент?',
    
    // Auth
    'auth.title': 'Auth Gateway',
    'auth.subtitle': 'Войдите для доступа к консоли администратора',
    'auth.email': 'Email адрес',
    'auth.password': 'Пароль',
    'auth.signin': 'Войти',
    'auth.forgot_password': 'Забыли пароль?',
    'auth.reset_here': 'Сбросить здесь',

    // Navigation
    'nav.dashboard': 'Дашборд',
    'nav.users': 'Пользователи',
    'nav.api_keys': 'API Ключи',
    'nav.oauth': 'OAuth Провайдеры',
    'nav.audit_logs': 'Журнал аудита',
    'nav.settings': 'Настройки',
    'nav.developers': 'Разработчики',
    'nav.webhooks': 'Вебхуки',
    'nav.service_accounts': 'Сервисные аккаунты',
    'nav.token_inspector': 'Инспектор токенов',
    'nav.logout': 'Выйти',
    'nav.menu': 'Меню',

    // Dashboard
    'dash.title': 'Обзор системы',
    'dash.total_users': 'Всего пользователей',
    'dash.active_now': 'Активны сейчас',
    'dash.2fa_enabled': '2FA Включена',
    'dash.api_keys': 'API Ключи',
    'dash.registrations': 'Регистрации (30 дней)',
    'dash.login_activity': 'Активность входов',

    // Users
    'users.title': 'Пользователи',
    'users.create_new': 'Создать пользователя',
    'users.filter_role': 'Все роли',
    'users.col_user': 'Пользователь',
    'users.col_role': 'Роль',
    'users.col_status': 'Статус',
    'users.col_2fa': '2FA',
    'users.col_created': 'Создан',
    'users.active': 'Активен',
    'users.blocked': 'Заблокирован',
    'users.view_details': 'Подробнее',
    
    // User Details
    'user.id': 'ID пользователя',
    'user.edit_profile': 'Ред. профиль',
    'user.security': 'Безопасность',
    'user.email_verified': 'Email подтвержден',
    'user.sessions': 'Активные сессии',
    'user.linked_accounts': 'Связанные аккаунты',
    'user.recent_activity': 'Недавняя активность',
    'user.revoke': 'Отозвать',
    'user.current': 'Текущая',
    'user.no_sessions': 'Нет активных сессий',
    'user.no_keys': 'Ключи не найдены',
    'user.danger_zone': 'Опасная зона',
    'user.reset_2fa': 'Сбросить 2FA',
    'user.reset_password_email': 'Отправить сброс пароля',
    'user.reset_2fa_confirm': 'Вы уверены? Пользователю придется настроить 2FA заново.',

    // User Edit
    'user.edit.title': 'Редактировать пользователя',
    'user.create.title': 'Новый пользователь',
    'user.form.username': 'Имя пользователя',
    'user.form.role': 'Роль',
    'user.form.fullname': 'Полное имя',
    'user.form.phone': 'Телефон',
    'user.form.active': 'Аккаунт активен',
    'user.form.active_desc': 'Снимите галочку, чтобы заблокировать вход.',
    'user.form.2fa_force': 'Принудительная 2FA',
    'user.form.save': 'Сохранить изменения',

    // API Keys
    'keys.title': 'API Ключи',
    'keys.generate': 'Генерировать ключ',
    'keys.owner': 'Владелец',
    'keys.prefix': 'Префикс',
    'keys.revoke_confirm': 'Отозвать этот ключ? Это действие необратимо.',
    'keys.revoked': 'Отозван',

    // OAuth
    'oauth.title': 'OAuth Провайдеры',
    'oauth.add': 'Добавить провайдера',
    'oauth.manage_desc': 'Управление входом через соцсети',
    'oauth.client_id': 'Client ID',
    'oauth.client_secret': 'Client Secret',
    'oauth.redirect_uris': 'Redirect URIs',
    'oauth.enable': 'Включить провайдера',
    'oauth.configure': 'Настроить',
    
    // Settings
    'settings.title': 'Настройки системы',
    'settings.branding': 'Внешний вид',
    'settings.branding_desc': 'Настройка страницы входа',
    'settings.roles': 'Роли и Права',
    'settings.roles_desc': 'Управление доступом',
    'settings.ip_security': 'IP Безопасность',
    'settings.ip_desc': 'Черные и белые списки IP',
    'settings.security_policies': 'Политики безопасности',
    'settings.password_policy': 'Политика паролей',
    'settings.email_smtp': 'Email и SMTP',
    'settings.manage_templates': 'Шаблоны писем',
    'settings.jwt_ttl': 'Время жизни JWT (мин)',
    'settings.refresh_ttl': 'Время жизни Refresh (дней)',
    'settings.min_pass': 'Мин. длина пароля',
    'settings.require_2fa_admin': 'Обязательная 2FA для админов',
    'settings.smtp_host': 'SMTP Хост',
    'settings.smtp_port': 'Порт',
    'settings.from_addr': 'Адрес отправителя',
    'settings.req_uppercase': 'Требовать заглавные',
    'settings.req_lowercase': 'Требовать строчные',
    'settings.req_numbers': 'Требовать цифры',
    'settings.req_special': 'Требовать спецсимволы',
    'settings.pass_history': 'История паролей',
    'settings.pass_expiry': 'Срок действия пароля (дней)',

    // IP Rules
    'ip.title': 'IP Безопасность',
    'ip.blacklist': 'Черный список (Block)',
    'ip.whitelist': 'Белый список (Allow)',
    'ip.add_block': 'Добавить блок',
    'ip.add_allow': 'Добавить разрешение',
    'ip.address': 'IP Адрес / CIDR',
    'ip.added_by': 'Добавил',

    // Webhooks
    'hooks.title': 'Вебхуки',
    'hooks.add': 'Добавить эндпоинт',
    'hooks.url': 'Endpoint URL',
    'hooks.events': 'События',
    'hooks.secret': 'Секретный ключ',
    'hooks.failures': 'Ошибки',

    // Service Accounts
    'sa.title': 'Сервисные аккаунты',
    'sa.create': 'Создать аккаунт',
    'sa.desc': 'M2M аутентификация',
    'sa.generated': 'Аккаунт создан',
    'sa.generated_desc': 'Скопируйте эти данные сейчас. Client Secret больше не будет показан.',
    
    // Roles & Permissions
    'roles.title': 'Роли',
    'perms.title': 'Права доступа (Permissions)',
    'perms.name': 'Название',
    'perms.resource': 'Ресурс',
    'perms.action': 'Действие',
    'roles.permissions': 'Права доступа',
    'roles.system_role': 'Системная',
    'roles.users_count': 'Пользователей',

    // Branding
    'brand.company': 'Название компании',
    'brand.logo': 'URL Логотипа',
    'brand.colors': 'Цвета',
    'brand.primary': 'Основной',
    'brand.bg': 'Фон',
    'brand.content': 'Контент',
    'brand.heading': 'Заголовок входа',
    'brand.subtitle': 'Подзаголовок',
    'brand.preview': 'Предпросмотр',
    'brand.socials': 'Показать соцсети',
    
    // Email Templates
    'email.templates': 'Шаблоны писем',
    'email.subject': 'Тема письма',
    'email.body': 'HTML Тело',
    'email.vars': 'Переменные',
    'email.preview': 'Предпросмотр',

    // SMS & System
    'sms.title': 'SMS Провайдеры',
    'sms.desc': 'Настройка шлюза для отправки SMS',
    'sms.provider': 'Выберите провайдера',
    'sms.test': 'Тест отправки',
    'sys.health': 'Здоровье системы',
    'sys.maintenance_on': 'Режим обслуживания ВКЛ',
    'sys.maintenance_off': 'Режим обслуживания ВЫКЛ',
    'sys.confirm_enable': 'Вы уверены, что хотите включить режим обслуживания? Пользователи не смогут войти.',
    'sys.confirm_disable': 'Выключить режим обслуживания?',

    // Inspector
    'inspector.title': 'Инспектор токенов',
    'inspector.desc': 'Декодирование и проверка JWT токенов',
    'inspector.paste': 'Вставьте JWT токен здесь',
    'inspector.header': 'Заголовок',
    'inspector.payload': 'Полезная нагрузка (Payload)',
    'inspector.invalid': 'Неверный формат токена',
  },
  en: {
    // Common
    'common.save': 'Save Changes',
    'common.saving': 'Saving...',
    'common.saved': 'Saved!',
    'common.cancel': 'Cancel',
    'common.delete': 'Delete',
    'common.edit': 'Edit',
    'common.create': 'Create',
    'common.back': 'Back',
    'common.loading': 'Loading...',
    'common.actions': 'Actions',
    'common.search': 'Search...',
    'common.status': 'Status',
    'common.created': 'Created',
    'common.yes': 'Yes',
    'common.no': 'No',
    'common.confirm_delete': 'Are you sure you want to delete this item?',

    // Auth
    'auth.title': 'Auth Gateway',
    'auth.subtitle': 'Sign in to access the admin console',
    'auth.email': 'Email Address',
    'auth.password': 'Password',
    'auth.signin': 'Sign In',
    'auth.forgot_password': 'Forgot password?',
    'auth.reset_here': 'Reset here',

    // Navigation
    'nav.dashboard': 'Dashboard',
    'nav.users': 'Users',
    'nav.api_keys': 'API Keys',
    'nav.oauth': 'OAuth Providers',
    'nav.audit_logs': 'Audit Logs',
    'nav.settings': 'Settings',
    'nav.developers': 'Developers',
    'nav.webhooks': 'Webhooks',
    'nav.service_accounts': 'Service Accounts',
    'nav.token_inspector': 'Token Inspector',
    'nav.logout': 'Sign Out',
    'nav.menu': 'Menu',

    // Dashboard
    'dash.title': 'System Overview',
    'dash.total_users': 'Total Users',
    'dash.active_now': 'Active Now',
    'dash.2fa_enabled': '2FA Enabled',
    'dash.api_keys': 'API Keys',
    'dash.registrations': 'Registrations (30 Days)',
    'dash.login_activity': 'Login Activity',

    // Users
    'users.title': 'Users',
    'users.create_new': 'Create User',
    'users.filter_role': 'All Roles',
    'users.col_user': 'User',
    'users.col_role': 'Role',
    'users.col_status': 'Status',
    'users.col_2fa': '2FA',
    'users.col_created': 'Created',
    'users.active': 'Active',
    'users.blocked': 'Blocked',
    'users.view_details': 'View Details',

    // User Details
    'user.id': 'User ID',
    'user.edit_profile': 'Edit Profile',
    'user.security': 'Security',
    'user.email_verified': 'Email Verified',
    'user.sessions': 'Active Sessions',
    'user.linked_accounts': 'Linked Accounts',
    'user.recent_activity': 'Recent Activity',
    'user.revoke': 'Revoke',
    'user.current': 'Current',
    'user.no_sessions': 'No active sessions found',
    'user.no_keys': 'No keys found',
    'user.danger_zone': 'Danger Zone',
    'user.reset_2fa': 'Reset 2FA',
    'user.reset_password_email': 'Send Password Reset',
    'user.reset_2fa_confirm': 'Are you sure? The user will need to set up 2FA again.',


    // User Edit
    'user.edit.title': 'Edit User',
    'user.create.title': 'Create New User',
    'user.form.username': 'Username',
    'user.form.role': 'Role',
    'user.form.fullname': 'Full Name',
    'user.form.phone': 'Phone number',
    'user.form.active': 'Account Active',
    'user.form.active_desc': 'Uncheck to block this user from signing in.',
    'user.form.2fa_force': 'Force 2FA',
    'user.form.save': 'Save Changes',

    // API Keys
    'keys.title': 'API Keys',
    'keys.generate': 'Generate New Key',
    'keys.owner': 'Owner',
    'keys.prefix': 'Prefix',
    'keys.revoke_confirm': 'Revoke this API key? This cannot be undone.',
    'keys.revoked': 'Revoked',

    // OAuth
    'oauth.title': 'OAuth Providers',
    'oauth.add': 'Add Provider',
    'oauth.manage_desc': 'Manage social login connections',
    'oauth.client_id': 'Client ID',
    'oauth.client_secret': 'Client Secret',
    'oauth.redirect_uris': 'Redirect URIs',
    'oauth.enable': 'Enable provider',
    'oauth.configure': 'Configure',

    // Settings
    'settings.title': 'System Settings',
    'settings.branding': 'Look & Feel',
    'settings.branding_desc': 'Customize the hosted login page',
    'settings.roles': 'Roles & Permissions',
    'settings.roles_desc': 'Manage user roles and access',
    'settings.ip_security': 'IP Security',
    'settings.ip_desc': 'Manage blocked IPs and whitelists',
    'settings.security_policies': 'Security Policies',
    'settings.password_policy': 'Password Policy',
    'settings.email_smtp': 'Email & SMTP',
    'settings.manage_templates': 'Manage Templates',
    'settings.jwt_ttl': 'JWT Access Token TTL (minutes)',
    'settings.refresh_ttl': 'JWT Refresh Token TTL (days)',
    'settings.min_pass': 'Password Minimum Length',
    'settings.require_2fa_admin': 'Require 2FA for Admins',
    'settings.smtp_host': 'SMTP Host',
    'settings.smtp_port': 'Port',
    'settings.from_addr': 'From Address',
    'settings.req_uppercase': 'Require Uppercase',
    'settings.req_lowercase': 'Require Lowercase',
    'settings.req_numbers': 'Require Numbers',
    'settings.req_special': 'Require Special Chars',
    'settings.pass_history': 'Password History',
    'settings.pass_expiry': 'Password Expiry (days)',

    // IP Rules
    'ip.title': 'IP Security',
    'ip.blacklist': 'Blocked IPs (Blacklist)',
    'ip.whitelist': 'Allowed IPs (Whitelist)',
    'ip.add_block': 'Add Block Rule',
    'ip.add_allow': 'Add Allow Rule',
    'ip.address': 'IP Address / CIDR',
    'ip.added_by': 'Added By',

    // Webhooks
    'hooks.title': 'Webhooks',
    'hooks.add': 'Add Endpoint',
    'hooks.url': 'Endpoint URL',
    'hooks.events': 'Events',
    'hooks.secret': 'Signing Secret',
    'hooks.failures': 'Failures',

    // Service Accounts
    'sa.title': 'Service Accounts',
    'sa.create': 'Create Service Account',
    'sa.desc': 'Manage machine-to-machine identities',
    'sa.generated': 'Service Account Created',
    'sa.generated_desc': 'Please copy these credentials now. You won\'t be able to see the Client Secret again.',

    // Roles & Permissions
    'roles.title': 'Roles',
    'perms.title': 'Permissions',
    'perms.name': 'Name',
    'perms.resource': 'Resource',
    'perms.action': 'Action',
    'roles.permissions': 'Permissions',
    'roles.system_role': 'System',
    'roles.users_count': 'Users',

    // Branding
    'brand.company': 'Company Name',
    'brand.logo': 'Logo URL',
    'brand.colors': 'Colors',
    'brand.primary': 'Primary Color',
    'brand.bg': 'Background',
    'brand.content': 'Page Content',
    'brand.heading': 'Login Heading',
    'brand.subtitle': 'Subtitle',
    'brand.preview': 'Live Preview',
    'brand.socials': 'Show Social Logins',

    // Email Templates
    'email.templates': 'Email Templates',
    'email.subject': 'Email Subject',
    'email.body': 'HTML Content',
    'email.vars': 'Available variables',
    'email.preview': 'Live Preview',

    // SMS & System
    'sms.title': 'SMS Providers',
    'sms.desc': 'Configure SMS Gateway for sending messages',
    'sms.provider': 'Select Provider',
    'sms.test': 'Test Sending',
    'sys.health': 'System Health',
    'sys.maintenance_on': 'Maintenance Mode ON',
    'sys.maintenance_off': 'Maintenance Mode OFF',
    'sys.confirm_enable': 'Are you sure you want to enable maintenance mode? Users will be locked out.',
    'sys.confirm_disable': 'Disable maintenance mode?',

    // Inspector
    'inspector.title': 'Token Inspector',
    'inspector.desc': 'Decode and verify JWT tokens',
    'inspector.paste': 'Paste JWT here',
    'inspector.header': 'Header',
    'inspector.payload': 'Payload',
    'inspector.invalid': 'Invalid Token Format',
  }
};

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

export const LanguageProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [language, setLanguageState] = useState<Language>('ru');

  useEffect(() => {
    const savedLang = localStorage.getItem('app_language') as Language;
    if (savedLang && (savedLang === 'en' || savedLang === 'ru')) {
      setLanguageState(savedLang);
    }
  }, []);

  const setLanguage = (lang: Language) => {
    setLanguageState(lang);
    localStorage.setItem('app_language', lang);
  };

  const t = (key: string): string => {
    return translations[language][key] || key;
  };

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  );
};

export const useLanguage = (): LanguageContextType => {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};