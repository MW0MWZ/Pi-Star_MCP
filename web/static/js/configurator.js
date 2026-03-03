// configurator.js — Service configurator for the admin panel
'use strict';

var Configurator = (function() {
    var services = [];
    var selectedService = null;

    // Services only available in DStarRepeater mode
    var dstarRepeaterOnly = { 'dstarrepeater': true };
    // Services only available in MMDVMHost mode (everything except the above, but dstargateway works in both)
    var mmdvmhostOnly = {
        'mmdvmhost': true,
        'dmrgateway': true, 'ysfgateway': true, 'p25gateway': true, 'nxdngateway': true,
        'dgidgateway': true, 'fmgateway': true, 'aprsgateway': true, 'dapnetgateway': true,
        'dmr2ysf': true, 'dmr2nxdn': true, 'ysf2dmr': true, 'ysf2nxdn': true,
        'ysf2p25': true, 'nxdn2dmr': true,
        'ysfparrot': true, 'p25parrot': true, 'nxdnparrot': true
    };

    function init() {
        loadServices();

        // Handle sidebar navigation (page switching)
        var sidebarLinks = document.querySelectorAll('.sidebar-link:not(.disabled)');
        sidebarLinks.forEach(function(link) {
            link.addEventListener('click', function(e) {
                e.preventDefault();
                sidebarLinks.forEach(function(l) { l.classList.remove('active'); });
                this.classList.add('active');

                // Show/hide pages based on sidebar link href
                var targetId = this.getAttribute('href').replace('#', '');
                var pages = document.querySelectorAll('.admin-page');
                pages.forEach(function(page) {
                    var pageId = page.id.replace('page-', '');
                    page.style.display = (pageId === targetId) ? '' : 'none';
                });

                // Reload services when switching to Configuration page
                if (targetId === 'configuration') {
                    loadServices();
                }
            });
        });

        // Re-render when hardware interface changes on the Radio page
        document.addEventListener('pistar:hwchange', function() {
            loadServices();
        });
    }

    function loadServices() {
        fetch('/admin/api/services')
            .then(function(r) {
                if (!r.ok) throw new Error('Failed to load services');
                return r.json();
            })
            .then(function(data) {
                services = data;
                renderServiceList();
            })
            .catch(function(err) {
                var el = document.getElementById('service-list');
                if (el) el.innerHTML = '<p class="error-message">' + escapeHTML(err.message) + '</p>';
            });
    }

    function detectHwInterface() {
        for (var i = 0; i < services.length; i++) {
            if (services[i].name === 'dstarrepeater' && services[i].enabled) return 'dstarrepeater';
        }
        return 'mmdvmhost';
    }

    function isServiceVisible(svc, hwMode) {
        if (hwMode === 'dstarrepeater') {
            // Hide MMDVMHost-only services
            return !mmdvmhostOnly[svc.name];
        } else {
            // Hide DStarRepeater-only services
            return !dstarRepeaterOnly[svc.name];
        }
    }

    function renderServiceList() {
        var container = document.getElementById('service-list');
        if (!container) return;

        container.innerHTML = '';
        var currentCategory = '';
        var hwMode = detectHwInterface();

        services.forEach(function(svc) {
            if (!isServiceVisible(svc, hwMode)) return;

            // Category header
            if (svc.category !== currentCategory) {
                currentCategory = svc.category;
                var label = document.createElement('div');
                label.className = 'service-category-label';
                label.setAttribute('data-i18n', 'admin.config.category.' + svc.category);
                label.textContent = i18n('admin.config.category.' + svc.category, categoryLabel(svc.category));
                container.appendChild(label);
            }

            var item = document.createElement('div');
            item.className = 'service-item';
            item.setAttribute('role', 'button');
            item.setAttribute('tabindex', '0');
            item.setAttribute('aria-label', svc.displayName);

            var nameSpan = document.createElement('span');
            nameSpan.className = 'service-name';
            nameSpan.textContent = svc.displayName;
            if (svc.hasSettings) {
                var icon = document.createElement('span');
                icon.className = 'settings-icon';
                icon.textContent = '\u2699';
                icon.setAttribute('aria-hidden', 'true');
                nameSpan.appendChild(icon);
            }
            item.appendChild(nameSpan);

            // Toggle switch
            var toggle = document.createElement('label');
            toggle.className = 'toggle-switch';
            var checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.checked = svc.enabled;
            checkbox.setAttribute('aria-label', i18n('admin.config.enableService', 'Enable') + ' ' + svc.displayName);
            var slider = document.createElement('span');
            slider.className = 'toggle-slider';
            toggle.appendChild(checkbox);
            toggle.appendChild(slider);
            item.appendChild(toggle);

            // Toggle enable/disable
            checkbox.addEventListener('change', (function(service) {
                return function(e) {
                    e.stopPropagation();
                    toggleService(service, e.target.checked, e.target);
                };
            })(svc));

            // Click to load settings
            item.addEventListener('click', (function(service) {
                return function(e) {
                    if (e.target.closest('.toggle-switch')) return;
                    selectService(service);
                };
            })(svc));

            item.addEventListener('keydown', (function(service) {
                return function(e) {
                    if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        selectService(service);
                    }
                };
            })(svc));

            container.appendChild(item);
        });
    }

    function toggleService(svc, enable, checkbox) {
        var action = enable ? 'enable' : 'disable';
        fetch('/admin/api/services/' + encodeURIComponent(svc.name) + '/' + action, {
            method: 'PUT'
        })
        .then(function(r) {
            if (!r.ok) return r.json().then(function(data) { throw data; });
            return r.json();
        })
        .then(function() {
            svc.enabled = enable;
            announce(svc.displayName + ' ' + (enable ? 'enabled' : 'disabled'));
        })
        .catch(function(err) {
            checkbox.checked = !enable; // revert
            var msg = err.error || 'Failed to update service';
            if (err.missingDeps) {
                msg += ': requires ' + err.missingDeps.join(', ');
            }
            if (err.dependents) {
                msg += ': depended on by ' + err.dependents.join(', ');
            }
            announce(msg);
        });
    }

    function selectService(svc) {
        selectedService = svc;

        // Highlight selected
        var items = document.querySelectorAll('.service-item');
        items.forEach(function(item) { item.classList.remove('selected'); });
        // Find the item in the visible list
        var visibleItems = document.querySelectorAll('.service-item');
        var visibleServices = services.filter(function(s) { return isServiceVisible(s, detectHwInterface()); });
        var idx = visibleServices.indexOf(svc);
        if (idx >= 0 && visibleItems[idx]) visibleItems[idx].classList.add('selected');

        if (!svc.hasSettings) {
            var content = document.getElementById('settings-content');
            if (content) {
                content.innerHTML = '<h2>' + escapeHTML(svc.displayName) + '</h2>' +
                    '<p class="placeholder-text" data-i18n="admin.config.noSettings">' +
                    i18n('admin.config.noSettings', 'No configurable settings for this service') + '</p>';
            }
            return;
        }

        loadSettings(svc);
    }

    function loadSettings(svc) {
        var content = document.getElementById('settings-content');
        if (content) {
            content.innerHTML = '<p class="loading">' + i18n('admin.config.loadingSettings', 'Loading settings...') + '</p>';
        }

        fetch('/admin/api/services/' + encodeURIComponent(svc.name) + '/settings')
            .then(function(r) {
                if (!r.ok) throw new Error('Failed to load settings');
                return r.json();
            })
            .then(function(data) {
                renderSettingsForm(svc, data.schema, data.values);
            })
            .catch(function(err) {
                if (content) {
                    content.innerHTML = '<p class="error-message">' + escapeHTML(err.message) + '</p>';
                }
            });
    }

    function renderSettingsForm(svc, schema, values) {
        var content = document.getElementById('settings-content');
        if (!content) return;

        var html = '<h2>' + escapeHTML(svc.displayName) + '</h2>';
        html += '<form class="settings-form" novalidate>';

        schema.groups.forEach(function(group) {
            html += '<div class="settings-group">';
            html += '<h3 data-i18n="' + escapeAttr(group.i18nKey) + '">' + escapeHTML(i18n(group.i18nKey, group.name)) + '</h3>';

            group.fields.forEach(function(field) {
                html += renderField(field, values[field.key]);
            });

            html += '</div>';
        });

        html += '<button type="submit" class="btn btn-save" data-i18n="admin.config.save">' +
                i18n('admin.config.save', 'Save Settings') + '</button>';
        html += '<span class="save-status" id="save-status"></span>';
        html += '</form>';

        content.innerHTML = html;

        // Bind form submit
        var form = content.querySelector('form');
        if (form) {
            form.addEventListener('submit', function(e) {
                e.preventDefault();
                saveSettings(svc, schema, form);
            });
        }
    }

    function renderField(field, value) {
        var html = '<div class="form-group">';
        var fieldId = 'field-' + field.key;
        var label = i18n(field.i18nLabel, field.key);

        if (field.fieldType === 'boolean') {
            html += '<label class="toggle-switch">';
            html += '<input type="checkbox" id="' + escapeAttr(fieldId) + '" name="' + escapeAttr(field.key) + '"';
            if (value === '1' || value === 'true') html += ' checked';
            html += '>';
            html += '<span class="toggle-slider"></span>';
            html += '</label>';
            html += ' <label for="' + escapeAttr(fieldId) + '" data-i18n="' + escapeAttr(field.i18nLabel) + '">' + escapeHTML(label) + '</label>';
        } else if (field.fieldType === 'select') {
            html += '<label for="' + escapeAttr(fieldId) + '" data-i18n="' + escapeAttr(field.i18nLabel) + '">' + escapeHTML(label) + '</label>';
            html += '<select id="' + escapeAttr(fieldId) + '" name="' + escapeAttr(field.key) + '"';
            if (field.validate) html += ' data-validate="' + escapeAttr(field.validate) + '"';
            html += ' data-label="' + escapeAttr(label) + '">';
            field.options.forEach(function(opt) {
                html += '<option value="' + escapeAttr(opt.value) + '"';
                if (opt.value === value) html += ' selected';
                html += ' data-i18n="' + escapeAttr(opt.i18nKey) + '">';
                html += escapeHTML(i18n(opt.i18nKey, opt.value));
                html += '</option>';
            });
            html += '</select>';
        } else {
            html += '<label for="' + escapeAttr(fieldId) + '" data-i18n="' + escapeAttr(field.i18nLabel) + '">' + escapeHTML(label) + '</label>';
            var inputType = field.fieldType === 'number' ? 'number' : 'text';
            html += '<input type="' + inputType + '" id="' + escapeAttr(fieldId) + '" name="' + escapeAttr(field.key) + '"';
            html += ' value="' + escapeAttr(value || '') + '"';
            if (field.validate) html += ' data-validate="' + escapeAttr(field.validate) + '"';
            html += ' data-label="' + escapeAttr(label) + '">';
        }

        if (field.helpI18n) {
            html += '<div class="help-text" data-i18n="' + escapeAttr(field.helpI18n) + '">' +
                    escapeHTML(i18n(field.helpI18n, '')) + '</div>';
        }

        html += '</div>';
        return html;
    }

    function saveSettings(svc, schema, form) {
        var values = {};
        var valid = true;
        var firstInvalid = null;

        // Collect values from form
        schema.groups.forEach(function(group) {
            group.fields.forEach(function(field) {
                var input = form.querySelector('[name="' + field.key + '"]');
                if (!input) return;

                if (field.fieldType === 'boolean') {
                    values[field.key] = input.checked ? '1' : '0';
                } else {
                    values[field.key] = input.value;
                    // Client-side validation
                    if (input.getAttribute('data-validate') && !Validate.validate(input)) {
                        valid = false;
                        if (!firstInvalid) firstInvalid = input;
                    }
                }
            });
        });

        if (!valid) {
            if (firstInvalid) firstInvalid.focus();
            return;
        }

        var statusEl = document.getElementById('save-status');
        var saveBtn = form.querySelector('.btn-save');
        if (saveBtn) saveBtn.disabled = true;

        fetch('/admin/api/services/' + encodeURIComponent(svc.name) + '/settings', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(values)
        })
        .then(function(r) {
            if (!r.ok) return r.json().then(function(data) { throw data; });
            return r.json();
        })
        .then(function() {
            if (statusEl) {
                statusEl.className = 'save-status success';
                statusEl.textContent = i18n('admin.config.saved', 'Settings saved');
            }
            announce(svc.displayName + ' settings saved');
            setTimeout(function() { if (statusEl) statusEl.textContent = ''; }, 3000);
        })
        .catch(function(err) {
            if (statusEl) {
                statusEl.className = 'save-status error';
                statusEl.textContent = err.error || 'Save failed';
            }
            // Show server-side field errors
            if (err.fields) {
                err.fields.forEach(function(fieldErr) {
                    var input = form.querySelector('[name="' + fieldErr.key + '"]');
                    if (input) {
                        input.setAttribute('aria-invalid', 'true');
                        var errorEl = document.createElement('div');
                        errorEl.className = 'error-message';
                        errorEl.textContent = fieldErr.message;
                        input.parentNode.appendChild(errorEl);
                    }
                });
            }
        })
        .finally(function() {
            if (saveBtn) saveBtn.disabled = false;
        });
    }

    function announce(message) {
        var region = document.getElementById('admin-status');
        if (region) {
            region.textContent = message;
            setTimeout(function() { region.textContent = ''; }, 5000);
        }
    }

    function categoryLabel(category) {
        switch (category) {
            case 'core': return 'Core';
            case 'gateway': return 'Gateways';
            case 'bridge': return 'Bridges';
            case 'utility': return 'Utilities';
            default: return category;
        }
    }

    function escapeHTML(str) {
        var div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    function escapeAttr(str) {
        return String(str).replace(/&/g, '&amp;').replace(/"/g, '&quot;')
                          .replace(/'/g, '&#39;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    return { loadServices: loadServices };
})();
