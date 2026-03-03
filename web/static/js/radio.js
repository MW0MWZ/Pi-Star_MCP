// radio.js — Radio configuration page for the admin panel
'use strict';

var RadioConfig = (function() {
    var schema = null;
    var values = {};
    var services = [];
    var hwInterface = 'mmdvmhost'; // 'mmdvmhost' or 'dstarrepeater'
    var dstarVariants = [];  // populated from services API
    var dstarHWType = '';    // current hardware type key

    function init() {
        var page = document.getElementById('page-radio');
        if (!page) return;
        loadAll();
    }

    function loadAll() {
        var container = document.getElementById('radio-content');
        if (container) {
            container.innerHTML = '<p class="loading">' + label('radio.loading', 'Loading radio settings...') + '</p>';
        }

        // Fetch radio settings and service list in parallel
        Promise.all([
            fetch('/admin/api/radio/settings').then(function(r) {
                if (!r.ok) throw new Error('Failed to load radio settings');
                return r.json();
            }),
            fetch('/admin/api/services').then(function(r) {
                if (!r.ok) throw new Error('Failed to load services');
                return r.json();
            })
        ])
        .then(function(results) {
            schema = results[0].schema;
            values = results[0].values;
            services = results[1];

            // Determine current hardware interface from service states
            var mmdvmEnabled = isServiceEnabled('mmdvmhost');
            var dstarRptEnabled = isServiceEnabled('dstarrepeater');
            if (dstarRptEnabled && !mmdvmEnabled) {
                hwInterface = 'dstarrepeater';
            } else {
                hwInterface = 'mmdvmhost';
            }

            // Capture DStarRepeater hardware type info from services response
            for (var i = 0; i < services.length; i++) {
                if (services[i].name === 'dstarrepeater') {
                    dstarVariants = services[i].hwVariants || [];
                    dstarHWType = services[i].hwType || '';
                    break;
                }
            }

            renderForm();
        })
        .catch(function(err) {
            if (container) {
                container.innerHTML = '<p class="error-message">' + esc(err.message) + '</p>';
            }
        });
    }

    // --- rendering ---

    function renderForm() {
        var container = document.getElementById('radio-content');
        if (!container) return;

        var html = '';

        // Hardware interface selector (above the form)
        html += renderHardwareInterface();

        html += '<form class="settings-form radio-form" novalidate>';

        schema.forEach(function(group) {
            // Mode enables only shown for MMDVMHost
            if (group.i18nKey === 'radio.modes') {
                html += '<div class="settings-group hw-mmdvmhost-only"';
                if (hwInterface !== 'mmdvmhost') html += ' style="display:none"';
                html += '>';
                html += '<h3>' + esc(label(group.i18nKey, group.name)) + '</h3>';
                html += renderModesSection(group.fields);
                html += '</div>';
                return;
            }

            html += '<div class="settings-group">';
            html += '<h3>' + esc(label(group.i18nKey, group.name)) + '</h3>';

            if (group.i18nKey === 'radio.frequencies') {
                html += renderFrequencySection(group.fields);
            } else {
                group.fields.forEach(function(field) {
                    // Radio ID and NXDN ID only relevant in MMDVMHost mode
                    if (field.key === 'dmrId' || field.key === 'nxdnId') {
                        html += '<div class="hw-mmdvmhost-only"';
                        if (hwInterface !== 'mmdvmhost') html += ' style="display:none"';
                        html += '>';
                        html += renderField(field);
                        html += '</div>';
                    } else {
                        html += renderField(field);
                    }
                });
            }

            html += '</div>';
        });

        // DStarRepeater hardware type selector (shown when DStarRepeater selected)
        html += '<div class="settings-group hw-dstarrepeater-only"';
        if (hwInterface !== 'dstarrepeater') html += ' style="display:none"';
        html += '>';
        html += '<h3>' + esc(label('radio.dstarHwType', 'Hardware Type')) + '</h3>';
        html += '<p class="help-text">' + esc(label('radio.dstarHwType.desc',
            'Select your D-Star hardware. Each type uses a different daemon binary.')) + '</p>';
        html += renderDStarHWTypeSelector();
        html += '<div id="dstar-hwtype-status"></div>';
        html += '</div>';

        html += '<button type="submit" class="btn btn-save">' +
                label('radio.save', 'Save Radio Settings') + '</button>';
        html += '<span class="save-status" id="radio-save-status"></span>';
        html += '</form>';

        container.innerHTML = html;

        // Bind hardware interface selector
        bindHardwareInterface(container);

        // Bind DStarRepeater hardware type selector
        bindDStarHWType(container);

        // Bind duplex segmented control
        var duplexInputs = container.querySelectorAll('input[name="duplex"]');
        duplexInputs.forEach(function(input) {
            input.addEventListener('change', function() {
                updateDuplexState(this.value === '1');
            });
        });
        var currentDuplex = values['duplex'] === '1';
        updateDuplexState(currentDuplex);

        // Bind mode toggles to control bridge sub-items
        bindModeToggles(container);

        // Bind form submit
        var form = container.querySelector('form');
        if (form) {
            form.addEventListener('submit', function(e) {
                e.preventDefault();
                saveRadioSettings(form);
            });
        }
    }

    // --- hardware interface selector ---

    function renderHardwareInterface() {
        var html = '<div class="settings-group hw-selector">';
        html += '<h3>' + esc(label('radio.hwInterface', 'Hardware Interface')) + '</h3>';
        html += '<div class="duplex-control">';

        // MMDVMHost option
        html += '<label class="duplex-option">';
        html += '<input type="radio" name="hwInterface" value="mmdvmhost"';
        if (hwInterface === 'mmdvmhost') html += ' checked';
        html += '>';
        html += '<span class="duplex-label">' + esc(label('radio.hwMMDVM', 'MMDVMHost')) + '</span>';
        html += '<span class="duplex-desc">' + esc(label('radio.hwMMDVM.desc',
            'Multi-mode digital voice — D-Star, DMR, YSF, P25, NXDN, POCSAG, FM')) + '</span>';
        html += '</label>';

        // DStarRepeater option
        html += '<label class="duplex-option">';
        html += '<input type="radio" name="hwInterface" value="dstarrepeater"';
        if (hwInterface === 'dstarrepeater') html += ' checked';
        html += '>';
        html += '<span class="duplex-label">' + esc(label('radio.hwDStarRpt', 'DStarRepeater')) + '</span>';
        html += '<span class="duplex-desc">' + esc(label('radio.hwDStarRpt.desc',
            'Legacy D-Star boards, DVAP — D-Star only')) + '</span>';
        html += '</label>';

        html += '</div>';
        html += '<div id="hw-switch-status"></div>';
        html += '</div>';
        return html;
    }

    function bindHardwareInterface(container) {
        var inputs = container.querySelectorAll('input[name="hwInterface"]');
        inputs.forEach(function(input) {
            input.addEventListener('change', function() {
                switchHardwareInterface(this.value);
            });
        });
    }

    function switchHardwareInterface(newInterface) {
        if (newInterface === hwInterface) return;

        var statusEl = document.getElementById('hw-switch-status');
        var inputs = document.querySelectorAll('input[name="hwInterface"]');
        inputs.forEach(function(i) { i.disabled = true; });

        if (statusEl) {
            statusEl.className = 'hw-switch-status';
            statusEl.textContent = 'Switching...';
        }

        var enableSvc = newInterface;
        var disableSvc = newInterface === 'mmdvmhost' ? 'dstarrepeater' : 'mmdvmhost';

        // Disable the old one first, then enable the new one
        fetch('/admin/api/services/' + encodeURIComponent(disableSvc) + '/disable', { method: 'PUT' })
        .then(function(r) {
            if (!r.ok) return r.json().then(function(d) { throw d; });
            return fetch('/admin/api/services/' + encodeURIComponent(enableSvc) + '/enable', { method: 'PUT' });
        })
        .then(function(r) {
            if (!r.ok) return r.json().then(function(d) { throw d; });
            return r.json();
        })
        .then(function() {
            hwInterface = newInterface;

            // Update local service states
            updateServiceState(enableSvc, true);
            updateServiceState(disableSvc, false);

            // Show/hide mode enables and DStarRepeater note
            var mmdvmSections = document.querySelectorAll('.hw-mmdvmhost-only');
            var dstarSections = document.querySelectorAll('.hw-dstarrepeater-only');
            mmdvmSections.forEach(function(el) {
                el.style.display = newInterface === 'mmdvmhost' ? '' : 'none';
            });
            dstarSections.forEach(function(el) {
                el.style.display = newInterface === 'dstarrepeater' ? '' : 'none';
            });

            if (statusEl) {
                statusEl.className = 'hw-switch-status hw-switch-success';
                statusEl.textContent = (newInterface === 'mmdvmhost' ? 'MMDVMHost' : 'DStarRepeater') + ' selected';
                setTimeout(function() { statusEl.textContent = ''; }, 3000);
            }

            // Notify other modules (e.g. configurator) of the change
            document.dispatchEvent(new CustomEvent('pistar:hwchange', { detail: { hwInterface: newInterface } }));
        })
        .catch(function(err) {
            // Revert radio button
            var revertInput = document.querySelector('input[name="hwInterface"][value="' + hwInterface + '"]');
            if (revertInput) revertInput.checked = true;

            if (statusEl) {
                statusEl.className = 'hw-switch-status hw-switch-error';
                var msg = err.error || 'Failed to switch hardware interface';
                if (err.dependents) {
                    msg += ' — disable ' + err.dependents.join(', ') + ' first';
                }
                statusEl.textContent = msg;
                setTimeout(function() { statusEl.textContent = ''; }, 5000);
            }
        })
        .finally(function() {
            inputs.forEach(function(i) { i.disabled = false; });
        });
    }

    // --- dstar hardware type selector ---

    function renderDStarHWTypeSelector() {
        if (!dstarVariants || dstarVariants.length === 0) {
            return '<p class="help-text">No hardware variants available</p>';
        }

        var html = '<div class="hwtype-grid">';
        dstarVariants.forEach(function(v) {
            var isSelected = v.key === dstarHWType;
            html += '<label class="hwtype-option' + (isSelected ? ' hwtype-selected' : '') + '">';
            html += '<input type="radio" name="dstarHWType" value="' + attr(v.key) + '"';
            if (isSelected) html += ' checked';
            html += '>';
            html += '<span class="hwtype-label">' + esc(v.displayName) + '</span>';
            html += '</label>';
        });
        html += '</div>';
        return html;
    }

    function bindDStarHWType(container) {
        var inputs = container.querySelectorAll('input[name="dstarHWType"]');
        inputs.forEach(function(input) {
            input.addEventListener('change', function() {
                setDStarHWType(this.value);
            });
        });
    }

    function setDStarHWType(newType) {
        if (newType === dstarHWType) return;

        var statusEl = document.getElementById('dstar-hwtype-status');
        var inputs = document.querySelectorAll('input[name="dstarHWType"]');
        inputs.forEach(function(i) { i.disabled = true; });

        if (statusEl) {
            statusEl.className = 'hw-switch-status';
            statusEl.textContent = label('radio.dstarHwType.saving', 'Saving...');
        }

        fetch('/admin/api/dstarrepeater/hwtype', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ hwType: newType })
        })
        .then(function(r) {
            if (!r.ok) return r.json().then(function(d) { throw d; });
            return r.json();
        })
        .then(function(data) {
            dstarHWType = newType;

            // Update visual selection
            var options = document.querySelectorAll('.hwtype-option');
            options.forEach(function(opt) {
                var radio = opt.querySelector('input');
                if (radio && radio.value === newType) {
                    opt.classList.add('hwtype-selected');
                } else {
                    opt.classList.remove('hwtype-selected');
                }
            });

            if (statusEl) {
                statusEl.className = 'hw-switch-status hw-switch-success';
                statusEl.textContent = data.displayName + ' ' + label('radio.dstarHwType.selected', 'selected');
                setTimeout(function() { statusEl.textContent = ''; }, 3000);
            }
            announce(data.displayName + ' selected');
        })
        .catch(function(err) {
            // Revert radio button
            var revert = document.querySelector('input[name="dstarHWType"][value="' + dstarHWType + '"]');
            if (revert) revert.checked = true;

            if (statusEl) {
                statusEl.className = 'hw-switch-status hw-switch-error';
                statusEl.textContent = err.error || 'Failed to set hardware type';
                setTimeout(function() { statusEl.textContent = ''; }, 5000);
            }
        })
        .finally(function() {
            inputs.forEach(function(i) { i.disabled = false; });
        });
    }

    // --- field rendering ---

    function renderField(field) {
        var html = '<div class="form-group">';
        var fieldId = 'radio-' + field.key;
        var value = values[field.key] !== undefined ? values[field.key] : field['default'];
        var fieldLabel = label(field.i18nLabel, field.label);

        if (field.fieldType === 'boolean') {
            html += '<label class="toggle-switch">';
            html += '<input type="checkbox" id="' + attr(fieldId) + '" name="' + attr(field.key) + '"';
            if (value === '1' || value === 'true') html += ' checked';
            html += '>';
            html += '<span class="toggle-slider"></span>';
            html += '</label>';
            html += ' <label for="' + attr(fieldId) + '">' + esc(fieldLabel) + '</label>';
        } else {
            html += '<label for="' + attr(fieldId) + '">' + esc(fieldLabel) + '</label>';
            var inputType = field.fieldType === 'number' ? 'number' : 'text';
            html += '<input type="' + inputType + '" id="' + attr(fieldId) + '" name="' + attr(field.key) + '"';
            html += ' value="' + attr(value || '') + '"';
            if (field.validate) html += ' data-validate="' + attr(field.validate) + '"';
            html += ' data-label="' + attr(fieldLabel) + '">';
        }

        if (field.helpI18n) {
            html += '<div class="help-text">' + esc(label(field.helpI18n, '')) + '</div>';
        }

        html += '</div>';
        return html;
    }

    // --- duplex / simplex segmented control ---

    function renderFrequencySection(fields) {
        var html = '<div class="freq-section">';
        var duplexValue = values['duplex'] === '1';

        fields.forEach(function(field) {
            if (field.fieldType === 'duplex') {
                // Render as segmented control
                html += '<div class="form-group">';
                html += '<div class="duplex-control">';
                html += '<label class="duplex-option">';
                html += '<input type="radio" name="duplex" value="0"' + (!duplexValue ? ' checked' : '') + '>';
                html += '<span class="duplex-label">' + esc(label('radio.simplex', 'Simplex')) + '</span>';
                html += '<span class="duplex-desc">' + esc(label('radio.simplex.desc', 'TX and RX on same frequency')) + '</span>';
                html += '</label>';
                html += '<label class="duplex-option">';
                html += '<input type="radio" name="duplex" value="1"' + (duplexValue ? ' checked' : '') + '>';
                html += '<span class="duplex-label">' + esc(label('radio.duplexMode', 'Duplex')) + '</span>';
                html += '<span class="duplex-desc">' + esc(label('radio.duplex.desc', 'Separate TX and RX frequencies')) + '</span>';
                html += '</label>';
                html += '</div>';
                html += '</div>';
            } else if (field.key === 'txFrequency') {
                html += '<div class="form-group freq-tx-group">';
                html += renderFieldInner(field);
                html += '</div>';
            } else {
                html += '<div class="form-group">';
                html += renderFieldInner(field);
                html += '</div>';
            }
        });

        html += '</div>';
        return html;
    }

    function renderFieldInner(field) {
        var html = '';
        var fieldId = 'radio-' + field.key;
        var value = values[field.key] !== undefined ? values[field.key] : field['default'];
        var fieldLabel = label(field.i18nLabel, field.label);

        html += '<label for="' + attr(fieldId) + '">' + esc(fieldLabel) + '</label>';
        html += '<input type="text" id="' + attr(fieldId) + '" name="' + attr(field.key) + '"';
        html += ' value="' + attr(value || '') + '"';
        if (field.validate) html += ' data-validate="' + attr(field.validate) + '"';
        html += ' data-label="' + attr(fieldLabel) + '">';

        if (field.helpI18n) {
            html += '<div class="help-text">' + esc(label(field.helpI18n, '')) + '</div>';
        }
        return html;
    }

    function updateDuplexState(isDuplex) {
        var txGroup = document.querySelector('.freq-tx-group');
        if (!txGroup) return;

        if (isDuplex) {
            txGroup.style.display = '';
        } else {
            txGroup.style.display = 'none';
            var rxInput = document.querySelector('[name="rxFrequency"]');
            var txInput = document.querySelector('[name="txFrequency"]');
            if (rxInput && txInput) {
                txInput.value = rxInput.value;
            }
        }

        var rxInput = document.querySelector('[name="rxFrequency"]');
        if (rxInput) {
            rxInput.removeEventListener('input', syncTxToRx);
            if (!isDuplex) {
                rxInput.addEventListener('input', syncTxToRx);
            }
        }
    }

    function syncTxToRx() {
        var txInput = document.querySelector('[name="txFrequency"]');
        if (txInput) txInput.value = this.value;
    }

    // --- mode enables with bridge sub-items ---

    function renderModesSection(fields) {
        var html = '<div class="mode-list">';

        fields.forEach(function(field) {
            var fieldId = 'radio-' + field.key;
            var value = values[field.key] !== undefined ? values[field.key] : field['default'];
            var fieldLabel = label(field.i18nLabel, field.label);
            var hasBridges = field.bridges && field.bridges.length > 0;

            html += '<div class="mode-item' + (hasBridges ? ' has-bridges' : '') + '">';

            // Main mode toggle
            html += '<div class="mode-toggle-row">';
            html += '<label class="toggle-switch">';
            html += '<input type="checkbox" id="' + attr(fieldId) + '" name="' + attr(field.key) + '"';
            html += ' data-mode-parent="true"';
            if (value === '1' || value === 'true') html += ' checked';
            html += '>';
            html += '<span class="toggle-slider"></span>';
            html += '</label>';
            html += '<label for="' + attr(fieldId) + '" class="mode-label">' + esc(fieldLabel) + '</label>';
            html += '</div>';

            // Bridge sub-items
            if (hasBridges) {
                html += '<div class="bridge-list" data-parent-mode="' + attr(field.key) + '">';
                field.bridges.forEach(function(bridge) {
                    var bridgeId = 'bridge-' + bridge.service;
                    var bridgeEnabled = isBridgeEnabled(bridge.service);
                    var modeEnabled = value === '1' || value === 'true';

                    html += '<div class="bridge-item">';
                    html += '<label class="toggle-switch toggle-switch-sm">';
                    html += '<input type="checkbox" id="' + attr(bridgeId) + '"';
                    html += ' data-bridge="' + attr(bridge.service) + '"';
                    if (bridgeEnabled) html += ' checked';
                    if (!modeEnabled) html += ' disabled';
                    html += '>';
                    html += '<span class="toggle-slider"></span>';
                    html += '</label>';
                    html += '<label for="' + attr(bridgeId) + '" class="bridge-label">' + esc(bridge.label) + '</label>';
                    html += '</div>';
                });
                html += '</div>';
            }

            html += '</div>';
        });

        html += '</div>';
        return html;
    }

    function isBridgeEnabled(serviceName) {
        for (var i = 0; i < services.length; i++) {
            if (services[i].name === serviceName) {
                return services[i].enabled;
            }
        }
        return false;
    }

    function isServiceEnabled(serviceName) {
        for (var i = 0; i < services.length; i++) {
            if (services[i].name === serviceName) {
                return services[i].enabled;
            }
        }
        return false;
    }

    function updateServiceState(serviceName, enabled) {
        for (var i = 0; i < services.length; i++) {
            if (services[i].name === serviceName) {
                services[i].enabled = enabled;
                return;
            }
        }
    }

    function bindModeToggles(container) {
        // When a parent mode is toggled, enable/disable its bridge sub-items
        var modeInputs = container.querySelectorAll('[data-mode-parent]');
        modeInputs.forEach(function(input) {
            input.addEventListener('change', function() {
                var modeKey = this.name;
                var bridgeList = container.querySelector('[data-parent-mode="' + modeKey + '"]');
                if (!bridgeList) return;

                var bridgeInputs = bridgeList.querySelectorAll('input[data-bridge]');
                bridgeInputs.forEach(function(bi) {
                    bi.disabled = !input.checked;
                    if (!input.checked) {
                        bi.checked = false;
                    }
                });
            });
        });

        // When a bridge toggle is changed, call the service enable/disable API
        var bridgeInputs = container.querySelectorAll('[data-bridge]');
        bridgeInputs.forEach(function(input) {
            input.addEventListener('change', function() {
                var svc = this.getAttribute('data-bridge');
                var enable = this.checked;
                toggleBridge(svc, enable, this);
            });
        });
    }

    function toggleBridge(svc, enable, checkbox) {
        var action = enable ? 'enable' : 'disable';
        fetch('/admin/api/services/' + encodeURIComponent(svc) + '/' + action, {
            method: 'PUT'
        })
        .then(function(r) {
            if (!r.ok) return r.json().then(function(data) { throw data; });
            return r.json();
        })
        .then(function() {
            updateServiceState(svc, enable);
            showBridgeToast(checkbox, svc + ' ' + (enable ? 'enabled' : 'disabled'), false);
        })
        .catch(function(err) {
            checkbox.checked = !enable; // revert
            var msg = err.error || 'Failed to update bridge';
            if (err.missingDeps) {
                msg += ': requires ' + err.missingDeps.join(', ');
            }
            if (err.dependents) {
                msg += ': depended on by ' + err.dependents.join(', ');
            }
            showBridgeToast(checkbox, msg, true);
        });
    }

    function showBridgeToast(anchor, message, isError) {
        // Remove any existing toast
        var old = document.querySelector('.bridge-toast');
        if (old) old.remove();

        var toast = document.createElement('div');
        toast.className = 'bridge-toast' + (isError ? ' bridge-toast-error' : '');
        toast.textContent = message;

        // Insert after the bridge item
        var item = anchor.closest('.bridge-item') || anchor.closest('.mode-item');
        if (item) {
            item.style.position = 'relative';
            item.appendChild(toast);
        }

        setTimeout(function() { toast.remove(); }, 4000);
    }

    // --- save ---

    function saveRadioSettings(form) {
        var formValues = {};
        var valid = true;
        var firstInvalid = null;

        schema.forEach(function(group) {
            group.fields.forEach(function(field) {
                if (field.fieldType === 'duplex') {
                    var checked = form.querySelector('input[name="duplex"]:checked');
                    formValues['duplex'] = checked ? checked.value : '0';
                    return;
                }

                var input = form.querySelector('[name="' + field.key + '"]');
                if (!input) return;

                if (field.fieldType === 'boolean') {
                    formValues[field.key] = input.checked ? '1' : '0';
                } else {
                    formValues[field.key] = input.value;
                    if (input.getAttribute('data-validate') && typeof Validate !== 'undefined' && !Validate.validate(input)) {
                        valid = false;
                        if (!firstInvalid) firstInvalid = input;
                    }
                }
            });
        });

        // In simplex mode, force TX = RX
        if (formValues['duplex'] === '0') {
            formValues['txFrequency'] = formValues['rxFrequency'];
        }

        if (!valid) {
            if (firstInvalid) firstInvalid.focus();
            return;
        }

        var statusEl = document.getElementById('radio-save-status');
        var saveBtn = form.querySelector('.btn-save');
        if (saveBtn) saveBtn.disabled = true;

        fetch('/admin/api/radio/settings', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formValues)
        })
        .then(function(r) {
            if (!r.ok) return r.json().then(function(data) { throw data; });
            return r.json();
        })
        .then(function(data) {
            if (statusEl) {
                statusEl.className = 'save-status success';
                statusEl.textContent = label('radio.saved', 'Radio settings saved') +
                    ' (' + data.filesWritten + ' ' + label('radio.filesUpdated', 'files updated') + ')';
            }
            announce(label('radio.saved', 'Radio settings saved'));
            setTimeout(function() { if (statusEl) statusEl.textContent = ''; }, 5000);
        })
        .catch(function(err) {
            if (statusEl) {
                statusEl.className = 'save-status error';
                statusEl.textContent = err.error || 'Save failed';
            }
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

    // --- helpers ---

    function label(i18nKey, fallback) {
        return i18n(i18nKey, fallback);
    }

    function announce(message) {
        var region = document.getElementById('admin-status');
        if (region) {
            region.textContent = message;
            setTimeout(function() { region.textContent = ''; }, 5000);
        }
    }

    function esc(str) {
        var div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    function attr(str) {
        return String(str).replace(/&/g, '&amp;').replace(/"/g, '&quot;')
                          .replace(/'/g, '&#39;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    return { load: loadAll };
})();
