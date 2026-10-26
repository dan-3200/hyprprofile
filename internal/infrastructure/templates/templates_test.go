package templates

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalogTemplatesAreEmbeddedAndOperable(t *testing.T) {
	t.Parallel()
	for _, template := range Templates() {
		metadata, files, err := LoadTemplate(template.Name)
		require.NoError(t, err, "LoadTemplate(%q)", template.Name)
		require.Equal(t, template.Name, metadata.Name)
		require.NotEmpty(t, files)
		entrypoint := templateFileContents(files, metadata.Entrypoint)
		for _, expected := range []string{"SUPER", "RETURN", "kitty", "hl.dsp.exit"} {
			assert.True(t, strings.Contains(entrypoint, expected) || templateFilesContain(files, expected), "template %q does not contain %q", template.Name, expected)
		}
	}
}

func TestLoadTemplateRejectsUnknownAndUnsafeNames(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"unknown", "../minimal", "/minimal"} {
		_, _, err := LoadTemplate(name)
		assert.Error(t, err, "LoadTemplate(%q)", name)
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
