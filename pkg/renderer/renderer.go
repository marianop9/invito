package renderer

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"
	"unicode"

	"invitation/pkg/domain"
	"invitation/web"
)

// Renderer manages parsed HTML templates and handles SSR and SSG output.
type Renderer struct {
	tmpl      *template.Template
	adminTmpl *template.Template
}

// New creates a new Renderer and parses all embedded templates and partials.
func New() (*Renderer, error) {
	var rootTmpl *template.Template

	funcs := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"icon": func(name string) template.HTML {
			data, err := web.Files.ReadFile("static/icons/" + name + ".svg")
			if err != nil {
				return ""
			}
			return template.HTML(data)
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
		"capitalizeFirst": func(s string) string {
			if s == "" {
				return s
			}
			r := []rune(s)
			r[0] = unicode.ToUpper(r[0])

			return string(r)
		},
		"formatDateTime": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			return t.Format("Jan 02, 2006 3:04 PM")
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}

	tmpl := template.New("invito").Funcs(funcs)

	// Parse all public templates and sub-templates in partials
	parsed, err := tmpl.ParseFS(web.Files,
		"templates/*.html",
		"templates/partials/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates from embed.FS: %w", err)
	}

	rootTmpl = parsed

	// Parse admin dashboard views
	adminTmpl := template.New("admin").Funcs(funcs)
	adminParsed, err := adminTmpl.ParseFS(web.Files,
		"templates/admin/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin templates from embed.FS: %w", err)
	}

	return &Renderer{
		tmpl:      parsed,
		adminTmpl: adminParsed,
	}, nil
}

// RenderInvitation renders a dynamic invitation page to the given writer.
func (r *Renderer) RenderInvitation(w io.Writer, inv *domain.Invitation) error {
	return r.tmpl.ExecuteTemplate(w, "invitation.html", inv)
}

// RenderIndex renders the landing page.
func (r *Renderer) RenderIndex(w io.Writer, data any) error {
	return r.tmpl.ExecuteTemplate(w, "index.html", data)
}

// RenderAdminIndex renders the events overview dashboard.
func (r *Renderer) RenderAdminIndex(w io.Writer, data any) error {
	return r.adminTmpl.ExecuteTemplate(w, "index.html", data)
}

// RenderAdminRSVPs renders the RSVP tracking dashboard for a single event.
func (r *Renderer) RenderAdminRSVPs(w io.Writer, data any) error {
	return r.adminTmpl.ExecuteTemplate(w, "rsvps.html", data)
}

// RenderAdminEditor renders the invitation builder & live preview editor.
func (r *Renderer) RenderAdminEditor(w io.Writer, data any) error {
	return r.adminTmpl.ExecuteTemplate(w, "editor.html", data)
}

// RenderToHTMLString performs Static Site Generation (SSG) for a single invitation.
func (r *Renderer) RenderToHTMLString(inv *domain.Invitation) (string, error) {
	var buf bytes.Buffer
	if err := r.RenderInvitation(&buf, inv); err != nil {
		return "", err
	}
	return buf.String(), nil
}
