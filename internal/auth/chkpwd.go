package auth

// chkpwd.go provides a fallback password verification mechanism by
// exec'ing unix_chkpwd for hash algorithms not handled in pure Go
// (e.g. yescrypt $y$ on Debian 12+).
