package main

import (
	"io/fs"
	"testing"
)

func TestEmbeddedContent(t *testing.T) {
	// Verify key files are embedded
	files := []string{
		"web/templates/shell.html",
		"web/templates/login.html",
		"web/templates/admin.html",
		"web/static/css/base.css",
		"web/static/css/components.css",
		"web/static/css/admin.css",
		"web/static/js/pistar.js",
		"web/static/js/validate.js",
		"web/static/js/radio.js",
		"web/static/js/configurator.js",
		"modules/core/module.json",
		"modules/core/panel.html",
		"modules/core/themes/default-light.css",
		"modules/lastHeard/module.json",
		"i18n/en.json",
	}

	for _, path := range files {
		data, err := fs.ReadFile(content, path)
		if err != nil {
			t.Errorf("missing embedded file %s: %v", path, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("embedded file %s is empty", path)
		}
	}

	// Count total embedded files
	count := 0
	fs.WalkDir(content, ".", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			count++
		}
		return nil
	})
	t.Logf("total embedded files: %d", count)
}
