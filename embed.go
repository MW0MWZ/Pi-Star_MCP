package main

import "embed"

// content holds all web assets, built-in modules, and shell-level i18n
// files embedded at compile time. This prevents users from modifying
// the served files on disk — changes require recompiling the binary.
//
// Paths within the FS mirror the repo layout:
//
//	web/templates/shell.html
//	web/static/css/base.css
//	modules/core/module.json
//	i18n/en.json
//
//go:embed web modules i18n
var content embed.FS
