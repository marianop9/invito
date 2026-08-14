package renderer

import (
	"bytes"
	"fmt"
	"html/template"
	"io"

	"invitation/pkg/domain"
	"invitation/web"
)

// Renderer manages parsed HTML templates and handles SSR and SSG output.
type Renderer struct {
	tmpl *template.Template
}

// New creates a new Renderer and parses all embedded templates and partials.
func New() (*Renderer, error) {
	tmpl := template.New("invito").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
	})

	// Parse all templates and sub-templates in partials
	parsed, err := tmpl.ParseFS(web.Files,
		"templates/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates from embed.FS: %w", err)
	}

	return &Renderer{tmpl: parsed}, nil
}

// RenderInvitation renders a dynamic invitation page to the given writer.
func (r *Renderer) RenderInvitation(w io.Writer, inv *domain.Invitation) error {
	return r.tmpl.ExecuteTemplate(w, "invitation.html", inv)
}

// RenderIndex renders the landing page.
func (r *Renderer) RenderIndex(w io.Writer, data any) error {
	return r.tmpl.ExecuteTemplate(w, "index.html", data)
}

// RenderToHTMLString performs Static Site Generation (SSG) for a single invitation.
func (r *Renderer) RenderToHTMLString(inv *domain.Invitation) (string, error) {
	var buf bytes.Buffer
	if err := r.RenderInvitation(&buf, inv); err != nil {
		return "", err
	}
	return buf.String(), nil
}
