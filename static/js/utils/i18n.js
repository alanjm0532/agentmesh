import en from '../locales/en.js';
import zhCN from '../locales/zh-CN.js';

const messages = { en, 'zh-CN': zhCN };
const languageKey = 'agentmesh.language';
let language = 'en';

export function hasTranslation(key) {
    return typeof key === 'string' && (Object.hasOwn(messages[language], key) || Object.hasOwn(en, key));
}

export function t(key, params = {}) {
    const message = messages[language][key] ?? en[key] ?? key;
    return message.replace(/\{(\w+)\}/g, (match, name) => params[name] ?? match);
}

export function setText(element, key, params = {}) {
    element.dataset.i18n = key;
    element.dataset.i18nParams = JSON.stringify(params);
    element.textContent = t(key, params);
}

export function renderTranslations() {
    document.documentElement.lang = language;
    document.querySelectorAll('[data-i18n]').forEach((element) => {
        element.textContent = t(element.dataset.i18n, JSON.parse(element.dataset.i18nParams ?? '{}'));
    });
    document.querySelectorAll('[data-i18n-aria-label]').forEach((element) => {
        element.setAttribute('aria-label', t(element.dataset.i18nAriaLabel));
    });
    document.querySelectorAll('[data-i18n-title]').forEach((element) => {
        element.title = t(element.dataset.i18nTitle);
    });
    document.querySelectorAll('[data-language]').forEach((button) => {
        button.setAttribute('aria-pressed', String(button.dataset.language === language));
    });
}

export function initI18n() {
    language = navigator.language.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en';
    try {
        const savedLanguage = localStorage.getItem(languageKey);
        if (Object.hasOwn(messages, savedLanguage)) {
            language = savedLanguage;
        }
    } catch {
        // Keep browser language detection available when storage is blocked.
    }
    renderTranslations();

    document.querySelectorAll('[data-language]').forEach((button) => {
        button.addEventListener('click', () => {
            language = button.dataset.language;
            renderTranslations();
            try {
                localStorage.setItem(languageKey, language);
            } catch {
                // Language switching still works for the current page without storage.
            }
        });
    });
}
