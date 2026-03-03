// pistar.js — WebSocket client, theme switcher, i18n engine
'use strict';

var PiStar = (function() {
    var ws = null;
    var i18nStrings = {};

    // WebSocket client
    function connectWebSocket() {
        var proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
        ws = new WebSocket(proto + '//' + location.host + '/ws');

        ws.onmessage = function(event) {
            var msg = JSON.parse(event.data);
            document.dispatchEvent(new CustomEvent('pistar:message', { detail: msg }));
        };

        ws.onclose = function() {
            // Reconnect after 3 seconds
            setTimeout(connectWebSocket, 3000);
        };
    }

    // i18n engine
    function loadTranslations(lang) {
        return fetch('/i18n/' + lang + '.json')
            .then(function(r) { return r.json(); })
            .then(function(strings) {
                i18nStrings = strings;
                applyTranslations();
            });
    }

    function applyTranslations() {
        var elements = document.querySelectorAll('[data-i18n]');
        elements.forEach(function(el) {
            var key = el.getAttribute('data-i18n');
            if (i18nStrings[key]) {
                el.textContent = i18nStrings[key];
            }
        });
    }

    function i18n(key, fallback) {
        return i18nStrings[key] || fallback || key;
    }

    // Theme switcher — persists to localStorage
    function setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        try { localStorage.setItem('pistar-theme', theme); } catch (e) {}
    }

    // Navigation toggle (mobile)
    function initNav() {
        var toggle = document.querySelector('.nav-toggle');
        var menu = document.getElementById('nav-menu');
        if (toggle && menu) {
            toggle.addEventListener('click', function() {
                var expanded = toggle.getAttribute('aria-expanded') === 'true';
                toggle.setAttribute('aria-expanded', !expanded);
                menu.classList.toggle('open');
            });
        }
    }

    // Init
    function init() {
        initNav();
        connectWebSocket();

        // Restore theme from localStorage and sync picker
        var savedTheme = null;
        try { savedTheme = localStorage.getItem('pistar-theme'); } catch (e) {}
        if (savedTheme) {
            setTheme(savedTheme);
        }
        var themePicker = document.getElementById('theme-picker');
        if (themePicker) {
            if (savedTheme) {
                themePicker.value = savedTheme;
            }
            themePicker.addEventListener('change', function() {
                setTheme(this.value);
            });
        }

        // Restore language from localStorage, auto-load translations
        var savedLang = null;
        try { savedLang = localStorage.getItem('pistar-lang'); } catch (e) {}
        var lang = savedLang || 'en';

        var langPicker = document.getElementById('lang-picker');
        if (langPicker) {
            if (savedLang) {
                langPicker.value = savedLang;
            }
            langPicker.addEventListener('change', function() {
                var newLang = this.value;
                try { localStorage.setItem('pistar-lang', newLang); } catch (e) {}
                loadTranslations(newLang);
            });
        }

        // Always load translations on init
        loadTranslations(lang);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    return { i18n: i18n, setTheme: setTheme, loadTranslations: loadTranslations };
})();

// Global i18n helper for modules
function i18n(key, fallback) {
    return PiStar.i18n(key, fallback);
}
