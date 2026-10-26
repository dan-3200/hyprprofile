package profile

// DefaultTemplate is used when creation omits a template.
const DefaultTemplate = "minimal"

// Template describes an embedded profile starting point.
type Template struct {
	Name             string
	Description      string
	Entrypoint       string
	RequiredCommands []string
}

// TemplateFile is one regular file in a template tree.
type TemplateFile struct {
	Path string
	Data []byte
}
