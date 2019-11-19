// +build prod

package assets

import (
	"html/template"

	"github.com/gin-gonic/gin/render"
)

//go:generate go-bindata -pkg $GOPACKAGE -o assets.generated.go -prefix ../ ../static/... ../templates/...

// TemplateRender return a render.HTMLRender that is suitable for production
// use.  It will loaded templates that are embedded into the application binary.
// The templates are loaded and parsed a single time and remain constant
// afterwards.
func TemplateRender() *templateRender {
	return &templateRender{
		templates: make(map[string]*template.Template),
	}
}

type templateRender struct {
	templates map[string]*template.Template
}

// Instance renders a template into an HTML instance.
func (r *templateRender) Instance(name string, data interface{}) render.Render {
	tmpl, ok := r.templates[name]
	if !ok {
		body := MustAssetString(name)
		tmpl = template.Must(template.New(name).Parse(body))
		r.templates[name] = tmpl
	}

	return render.HTML{
		Template: tmpl,
		Data:     data,
	}
}

// TemplateRenderer must implement the render.HTMLRenderer interface
var _ render.HTMLRender = (*templateRender)(nil)
