// lastHeard/panel.js — Last heard table logic
'use strict';

(function() {
    var MAX_ROWS = 20;

    document.addEventListener('pistar:message', function(e) {
        var msg = e.detail;
        if (msg.type !== 'mqtt') return;
        if (!msg.topic || msg.topic.indexOf('mmdvm/') !== 0) return;

        // TODO: Parse MQTT payload and update the last heard table
        // This will be implemented once the MQTT topic schema is finalised
    });
})();
