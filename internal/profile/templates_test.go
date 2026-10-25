package profile

import (
	"strings"
	"testing"
)

func TestCatalogTemplatesAreEmbeddedAndOperable(t *testing.T) {
	t.Parallel()
	for _, template := range Templates() {
		metadata, files, err := LoadTemplate(template.Name)
		if err != nil {
			t.Errorf("LoadTemplate(%q) error = %v", template.Name, err)
			continue
		}
		if metadata.Name != template.Name || len(files) == 0 {
			t.Errorf("LoadTemplate(%q) returned incomplete data", template.Name)
			continue
		}
		entrypoint := templateFileContents(files, metadata.Entrypoint)
		for _, expected := range []string{"SUPER", "RETURN", "kitty", "hl.dsp.exit"} {
			if !strings.Contains(entrypoint, expected) && !templateFilesContain(files, expected) {
				t.Errorf("template %q does not contain %q", template.Name, expected)
			}
		}
	}
}

func TestLoadTemplateRejectsUnknownAndUnsafeNames(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"unknown", "../minimal", "/minimal"} {
		if _, _, err := LoadTemplate(name); err == nil {
			t.Errorf("LoadTemplate(%q) unexpectedly succeeded", name)
		}
	}
}

func TestSerpantinumTemplateContainsModules(t *testing.T) {
	t.Parallel()
	_, files, err := LoadTemplate("serpantinum")
	if err != nil {
		t.Fatalf("LoadTemplate(serpantinum) error = %v", err)
	}
	for _, expected := range []string{"config/autostart.lua", "config/keybinds.lua", "config/settings.lua"} {
		if templateFileContents(files, expected) == "" {
			t.Errorf("serpantinum template does not contain %q", expected)
		}
	}
}

func templateFileContents(files []TemplateFile, name string) string {
	for _, file := range files {
		if file.Path == name {
			return string(file.Data)
		}
	}
	return ""
}

func templateFilesContain(files []TemplateFile, expected string) bool {
	for _, file := range files {
		if strings.Contains(string(file.Data), expected) {
			return true
		}
	}
	return false
}
