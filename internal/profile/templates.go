package profile

import (
	"embed"
	"fmt"
	"io/fs"
)

// DefaultTemplate is used when create does not receive an explicit selection.
const DefaultTemplate = "minimal"

// Template describes an embedded profile starting point.
type Template struct {
	Name        string
	Description string
}

var templateCatalog = []Template{
	{Name: "bare", Description: "terminal, window controls and session exit"},
	{Name: "minimal", Description: "interactive desktop with launcher, bar and notifications"},
}

//go:embed templates/*.lua
var templateFiles embed.FS

// Templates returns the available templates in display order.
func Templates() []Template {
	result := make([]Template, len(templateCatalog))
	copy(result, templateCatalog)
	return result
}

// LoadTemplate returns the embedded Lua source for a named template.
func LoadTemplate(name string) ([]byte, error) {
	if err := ValidateID(name); err != nil {
		return nil, fmt.Errorf("invalid template name %q", name)
	}
	data, err := fs.ReadFile(templateFiles, "templates/"+name+".lua")
	if err != nil {
		return nil, fmt.Errorf("template %q is unavailable", name)
	}
	return data, nil
}
