package profile

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

// DefaultTemplate is used when create does not receive an explicit selection.
const DefaultTemplate = "minimal"

// Template describes an embedded profile starting point.
type Template struct {
	Name             string
	Description      string
	Entrypoint       string
	RequiredCommands []string
}

// TemplateFile is one regular file in an embedded template tree.
type TemplateFile struct {
	Path string
	Data []byte
}

var templateCatalog = []Template{
	{
		Name:             "bare",
		Description:      "terminal, window controls and session exit",
		Entrypoint:       DefaultEntrypoint,
		RequiredCommands: []string{"kitty"},
	},
	{
		Name:             "minimal",
		Description:      "interactive desktop with launcher, bar and notifications",
		Entrypoint:       DefaultEntrypoint,
		RequiredCommands: []string{"kitty", "wofi"},
	},
	{
		Name:        "serpantinum",
		Description: "modular desktop configured for the Serpantinum shell",
		Entrypoint:  DefaultEntrypoint,
		RequiredCommands: []string{
			"kitty", "serpantinum", "serpantinumd",
		},
	},
}

//go:embed templates
var templateFiles embed.FS

// Templates returns the available templates in display order.
func Templates() []Template {
	result := make([]Template, len(templateCatalog))
	copy(result, templateCatalog)
	return result
}

// LoadTemplate returns metadata and all embedded files for a named template.
func LoadTemplate(name string) (Template, []TemplateFile, error) {
	if err := ValidateID(name); err != nil {
		return Template{}, nil, fmt.Errorf("invalid template name %q", name)
	}
	metadata, found := findTemplate(name)
	if !found {
		return Template{}, nil, fmt.Errorf("template %q is unavailable", name)
	}

	root := "templates/" + name
	files := make([]TemplateFile, 0)
	walkErr := fs.WalkDir(templateFiles, root, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative := strings.TrimPrefix(filePath, root+"/")
		if relative == filePath || relative == "" || !entry.Type().IsRegular() {
			return fmt.Errorf("unsafe file in template %q", name)
		}
		if err := ValidateRelativePath(relative); err != nil {
			return fmt.Errorf("unsafe path in template %q: %w", name, err)
		}
		data, err := fs.ReadFile(templateFiles, filePath)
		if err != nil {
			return fmt.Errorf("read embedded file %q: %w", relative, err)
		}
		files = append(files, TemplateFile{Path: relative, Data: data})
		return nil
	})
	if walkErr != nil {
		return Template{}, nil, fmt.Errorf("load template %q: %w", name, walkErr)
	}
	if !slices.ContainsFunc(files, func(file TemplateFile) bool { return file.Path == metadata.Entrypoint }) {
		return Template{}, nil, fmt.Errorf("template %q has no entrypoint", name)
	}
	return metadata, files, nil
}

func findTemplate(name string) (Template, bool) {
	for _, template := range templateCatalog {
		if template.Name == name {
			return template, true
		}
	}
	return Template{}, false
}
