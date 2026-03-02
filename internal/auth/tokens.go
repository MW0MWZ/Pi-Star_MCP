package auth

// tokens.go implements API bearer token creation, verification, and
// revocation. Tokens are stored as SHA-256 hashes; the plaintext is
// returned once at creation and never stored.
