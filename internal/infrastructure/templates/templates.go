// Package templates provides embedded Hyprland configuration templates.
package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/dan-3200/hyprprofile/internal/domain/profile"
)

// Template is the domain template metadata.
type Template = profile.Template

// TemplateFile is a file from an embedded template.
type TemplateFile = profile.TemplateFile

var templateCatalog = []Template{
	{
		Name:             "bare",
		Description:      "terminal, window controls and session exit",
		Entrypoint:       profile.DefaultEntrypoint,
		RequiredCommands: []string{"kitty"},
	},
	{
		Name:             "minimal",
		Description:      "interactive desktop with launcher, bar and notifications",
		Entrypoint:       profile.DefaultEntrypoint,
		RequiredCommands: []string{"kitty", "wofi"},
	},
	{
		Name:        "serpantinum",
		Description: "modular desktop configured for the Serpantinum shell",
		Entrypoint:  profile.DefaultEntrypoint,
		RequiredCommands: []string{
			"kitty", "serpantinum", "serpantinumd",
		},
	},
}

//go:embed templates
var templateFiles embed.FS

// Catalog exposes the embedded template metadata to application use cases.
type Catalog struct{}

// Templates returns the embedded template catalog.
func (Catalog) Templates() []profile.Template { return Templates() }

// Templates returns the available templates in display order.
func Templates() []Template {
	result := make([]Template, len(templateCatalog))
	copy(result, templateCatalog)
	return result
}

// LoadTemplate returns metadata and all embedded files for a named template.
func LoadTemplate(name string) (Template, []TemplateFile, error) {
	if err := profile.ValidateID(name); err != nil {
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
		if err := profile.ValidateRelativePath(relative); err != nil {
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
