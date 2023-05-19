package webui

import (
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates
var templatesFS embed.FS

// Template returns a template from the embedded templatesFS.
func Template(name string) (*template.Template, error) {
	tpl, err := template.ParseFS(templatesFS, "templates/"+name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	return tpl, nil
}
