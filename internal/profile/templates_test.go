package profile

import (
	"strings"
	"testing"
)

func TestCatalogTemplatesAreEmbeddedAndOperable(t *testing.T) {
	t.Parallel()
	for _, template := range Templates() {
		data, err := LoadTemplate(template.Name)
		if err != nil {
			t.Errorf("LoadTemplate(%q) error = %v", template.Name, err)
			continue
		}
		for _, expected := range []string{"SUPER", "RETURN", "kitty", "hl.dsp.exit"} {
			if !strings.Contains(string(data), expected) {
				t.Errorf("template %q does not contain %q", template.Name, expected)
			}
		}
	}
}

func TestLoadTemplateRejectsUnknownAndUnsafeNames(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"unknown", "../minimal", "/minimal"} {
		if _, err := LoadTemplate(name); err == nil {
			t.Errorf("LoadTemplate(%q) unexpectedly succeeded", name)
		}
	}
}
