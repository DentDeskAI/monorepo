// src/hooks/useTranslation.js
import { useState, useEffect, useCallback, useMemo } from 'react';
import translations from '../i18n/translations.json';

// Default language
const DEFAULT_LANG = 'ru';
const LANG_STORAGE_KEY = 'crm_lang';
const LANG_EVENT = 'crm_lang_change';

export function useTranslation() {
    // Initialize from localStorage immediately (not deferred to useEffect)
    const [lang, setLang] = useState(() => {
        if (typeof window === 'undefined') return DEFAULT_LANG;
        const saved = localStorage.getItem(LANG_STORAGE_KEY);
        return (saved && ['ru', 'en', 'kz'].includes(saved)) ? saved : DEFAULT_LANG;
    });

    const t = useCallback((key) => {
        const root = translations[lang] ?? translations[DEFAULT_LANG];
        const keys = key.split('.');
        let result = root;
        for (const k of keys) {
            if (result?.[k] !== undefined) result = result[k];
            else return key;
        }
        return typeof result === 'string' ? result : key;
    }, [lang]);

    // Helper to switch language
    const setLanguage = useCallback((newLang) => {
        if (['ru', 'en', 'kz'].includes(newLang)) {
            setLang(newLang);
            if (typeof window !== 'undefined') {
                localStorage.setItem(LANG_STORAGE_KEY, newLang);
                window.dispatchEvent(new CustomEvent(LANG_EVENT, { detail: newLang }));
            }
        }
    }, []);

    // Listen for changes from other tabs/windows or component instances
    useEffect(() => {
        const onLangChange = (event) => {
            const next = event?.detail;
            if (next && ['ru', 'en', 'kz'].includes(next)) {
                setLang(next);
            }
        };

        const onStorage = (event) => {
            if (event.key !== LANG_STORAGE_KEY || !event.newValue) return;
            if (['ru', 'en', 'kz'].includes(event.newValue)) {
                setLang(event.newValue);
            }
        };

        window.addEventListener(LANG_EVENT, onLangChange);
        window.addEventListener('storage', onStorage);
        return () => {
            window.removeEventListener(LANG_EVENT, onLangChange);
            window.removeEventListener('storage', onStorage);
        };
    }, []);

    return { t, lang, setLanguage };
}
