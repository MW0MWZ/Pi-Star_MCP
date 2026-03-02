// core/panel.js — System info panel logic
'use strict';

(function() {
    document.addEventListener('pistar:message', function(e) {
        var msg = e.detail;
        if (msg.type !== 'system') return;

        var p = msg.payload;
        if (p.cpuTemp !== undefined) {
            document.getElementById('core-cpu-temp').textContent = p.cpuTemp.toFixed(1) + '\u00B0C';
        }
        if (p.cpuLoad !== undefined) {
            document.getElementById('core-cpu-load').textContent = (p.cpuLoad * 100).toFixed(0) + '%';
        }
        if (p.uptime !== undefined) {
            var h = Math.floor(p.uptime / 3600);
            var m = Math.floor((p.uptime % 3600) / 60);
            document.getElementById('core-uptime').textContent = h + 'h ' + m + 'm';
        }
        if (p.services) {
            var tbody = document.getElementById('core-services');
            tbody.innerHTML = '';
            Object.keys(p.services).forEach(function(name) {
                var tr = document.createElement('tr');
                var tdName = document.createElement('td');
                tdName.textContent = name;
                var tdStatus = document.createElement('td');
                tdStatus.textContent = i18n('core.service.' + p.services[name], p.services[name]);
                tdStatus.className = 'status-' + p.services[name];
                tr.appendChild(tdName);
                tr.appendChild(tdStatus);
                tbody.appendChild(tr);
            });
        }
    });
})();
