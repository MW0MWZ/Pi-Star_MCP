package mqttclient

// relay.go bridges MQTT messages to the WebSocket hub, forwarding
// each received message with its topic and timestamp to all connected
// browser clients.
