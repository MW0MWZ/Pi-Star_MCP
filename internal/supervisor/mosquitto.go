package supervisor

// mosquitto.go handles Mosquitto-specific logic: port availability
// detection, private config generation (localhost-only, no auth),
// and spawning the broker as a child process.
