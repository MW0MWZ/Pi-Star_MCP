// validate.js — data-validate form validation library
'use strict';

var Validate = (function() {
    var validators = {
        required: function(value) {
            return value.trim() !== '';
        },
        numeric: function(value) {
            return /^\d+$/.test(value);
        },
        decimal: function(value) {
            return /^\d+(\.\d+)?$/.test(value);
        },
        callsign: function(value) {
            return /^[A-Z0-9]{1,3}[0-9][A-Z0-9]{0,4}$/i.test(value);
        },
        ip: function(value) {
            return /^(\d{1,3}\.){3}\d{1,3}$/.test(value) &&
                value.split('.').every(function(n) { return parseInt(n, 10) <= 255; });
        },
        port: function(value) {
            var n = parseInt(value, 10);
            return n >= 1 && n <= 65535;
        },
        hostname: function(value) {
            return /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$/.test(value);
        }
    };

    function parseRules(str) {
        return str.split(',').map(function(rule) {
            var parts = rule.split(':');
            return { name: parts[0], args: parts.slice(1) };
        });
    }

    function validateField(input) {
        var rulesStr = input.getAttribute('data-validate');
        if (!rulesStr) return true;

        var label = input.getAttribute('data-label') || input.name || 'Field';
        var rules = parseRules(rulesStr);
        var value = input.value;

        for (var i = 0; i < rules.length; i++) {
            var rule = rules[i];

            if (rule.name === 'required' && !validators.required(value)) {
                showError(input, i18n('validation.required', '{label} is required').replace('{label}', label));
                return false;
            }

            if (value === '' && rule.name !== 'required') continue;

            if (rule.name === 'range') {
                var min = parseFloat(rule.args[0]);
                var max = parseFloat(rule.args[1]);
                var num = parseFloat(value);
                if (isNaN(num) || num < min || num > max) {
                    showError(input, i18n('validation.range', '{label} must be between {min} and {max}')
                        .replace('{label}', label).replace('{min}', min).replace('{max}', max));
                    return false;
                }
            } else if (rule.name === 'minlen') {
                if (value.length < parseInt(rule.args[0], 10)) {
                    showError(input, i18n('validation.minlen', '{label} must be at least {n} characters')
                        .replace('{label}', label).replace('{n}', rule.args[0]));
                    return false;
                }
            } else if (rule.name === 'maxlen') {
                if (value.length > parseInt(rule.args[0], 10)) {
                    showError(input, i18n('validation.maxlen', '{label} must be at most {n} characters')
                        .replace('{label}', label).replace('{n}', rule.args[0]));
                    return false;
                }
            } else if (rule.name === 'pattern') {
                var regex = new RegExp(rule.args.join(':'));
                if (!regex.test(value)) {
                    showError(input, i18n('validation.pattern', '{label} format is invalid').replace('{label}', label));
                    return false;
                }
            } else if (validators[rule.name] && !validators[rule.name](value)) {
                showError(input, i18n('validation.' + rule.name, '{label} is invalid').replace('{label}', label));
                return false;
            }
        }

        clearError(input);
        return true;
    }

    function showError(input, message) {
        clearError(input);
        input.setAttribute('aria-invalid', 'true');
        var errorEl = document.createElement('div');
        errorEl.className = 'error-message';
        errorEl.id = input.id + '-error';
        errorEl.textContent = message;
        input.setAttribute('aria-describedby', errorEl.id);
        input.parentNode.appendChild(errorEl);
    }

    function clearError(input) {
        input.removeAttribute('aria-invalid');
        input.removeAttribute('aria-describedby');
        var existing = input.parentNode.querySelector('.error-message');
        if (existing) existing.remove();
    }

    function register(name, fn) {
        validators[name] = fn;
    }

    // Auto-bind to forms
    function init() {
        document.addEventListener('blur', function(e) {
            if (e.target.getAttribute('data-validate')) {
                validateField(e.target);
            }
        }, true);

        document.addEventListener('submit', function(e) {
            var form = e.target;
            var inputs = form.querySelectorAll('[data-validate]');
            var valid = true;
            var firstInvalid = null;

            inputs.forEach(function(input) {
                if (!validateField(input)) {
                    valid = false;
                    if (!firstInvalid) firstInvalid = input;
                }
            });

            if (!valid) {
                e.preventDefault();
                if (firstInvalid) firstInvalid.focus();
            }
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    return { validate: validateField, register: register };
})();
