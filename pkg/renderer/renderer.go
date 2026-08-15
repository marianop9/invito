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
	var rootTmpl *template.Template

	funcs := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"renderSection": func(sec domain.Section, inv *domain.Invitation) (template.HTML, error) {
			if sec == nil || rootTmpl == nil {
				return "", nil
			}
			tmplName := sec.TemplateName()
			if tmplName == "" {
				return "", nil
			}
			data := map[string]any{
				"Section":    sec,
				"Invitation": inv,
			}
			var buf bytes.Buffer
			if err := rootTmpl.ExecuteTemplate(&buf, tmplName, data); err != nil {
				return "", fmt.Errorf("failed to render section %q with template %q: %w", sec.Type(), tmplName, err)
			}
			return template.HTML(buf.String()), nil
		},
	}

	tmpl := template.New("invito").Funcs(funcs)

	// Parse all templates and sub-templates in partials
	parsed, err := tmpl.ParseFS(web.Files,
		"templates/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates from embed.FS: %w", err)
	}

	rootTmpl = parsed
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
