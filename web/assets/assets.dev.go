// +build !prod

package assets

import (
	"html/template"

	"github.com/gin-gonic/gin/render"
)

//go:generate go-bindata -debug -pkg $GOPACKAGE -o assets.generated.go -prefix ../ ../static/... ../templates/...

// TemplateRender returns a render.HTMLRender that is suitable for development
// purposes.  It will load templates from the web/template directory using
// bindata and compile them each time they are loaded.  This allows the template
// to change on disk while the server is running and its changes to be seen
// immediately.
func TemplateRender() templateRender {
	return templateRender{}
}

type templateRender struct{}

// Instance loads a template and returns it in an HTML instance.
func (r templateRender) Instance(name string, data interface{}) render.Render {
	body := MustAssetString(name)
	tmpl := template.Must(template.New(name).Parse(body))

	return render.HTML{
		Template: tmpl,
		Data:     data,
	}
}

// templateRenderer must implement the render.HTMLRenderer interface
var _ render.HTMLRender = templateRender{}
