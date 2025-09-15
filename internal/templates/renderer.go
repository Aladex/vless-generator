package templates

import (
	"embed"
	"html/template"
	"strings"

	"vless-generator/internal/config"
	"vless-generator/internal/i18n"

	"github.com/sirupsen/logrus"
)

// TemplateRenderer handles HTML template rendering
type TemplateRenderer struct {
	templates map[string]*template.Template
	logger    *logrus.Entry
	htmlFS    embed.FS
}

// NewTemplateRenderer creates a new template renderer with embedded filesystem
func NewTemplateRenderer(htmlFS embed.FS) *TemplateRenderer {
	return &TemplateRenderer{
		templates: make(map[string]*template.Template),
		logger:    logrus.WithField("component", "template_renderer"),
		htmlFS:    htmlFS,
	}
}

// LoadTemplates loads embedded HTML templates from main package
func (tr *TemplateRenderer) LoadTemplates() error {
	tr.logger.Info("Loading embedded HTML templates from web/templates")

	templateNames := []string{"home", "config"}

	for _, name := range templateNames {
		templateFile := "web/templates/" + name + ".html"

		templateContent, err := tr.htmlFS.ReadFile(templateFile)
		if err != nil {
			return err
		}

		tmpl, err := template.New(name).Parse(string(templateContent))
		if err != nil {
			return err
		}

		tr.templates[name] = tmpl
		tr.logger.WithField("template", name).Debug("Template loaded successfully")
	}

	tr.logger.WithField("count", len(tr.templates)).Info("All HTML templates loaded successfully")
	return nil
}

// HomePageData represents data for home page template
type HomePageData struct {
	Title         string
	Language      string
	Texts         i18n.Texts
	DefaultConfig *config.DynamicConfig
}

// ConfigPageData represents data for config page template
type ConfigPageData struct {
	Title          string
	Language       string
	Texts          i18n.Texts
	ConfigType     string // Uppercase for display (e.g., "VLESS")
	ConfigTypeOrig string // Original lowercase for URLs (e.g., "vless")
	UUID           string
	QRCode         string
	VlessURL       string
	QueryString    string
}

// RenderHomePage renders the home page template
func (tr *TemplateRenderer) RenderHomePage(data HomePageData) (string, error) {
	tmpl, exists := tr.templates["home"]
	if !exists {
		return "", ErrTemplateNotFound{"home"}
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderConfigPage renders the config page template
func (tr *TemplateRenderer) RenderConfigPage(data ConfigPageData) (string, error) {
	tmpl, exists := tr.templates["config"]
	if !exists {
		return "", ErrTemplateNotFound{"config"}
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ErrTemplateNotFound represents a template not found error
type ErrTemplateNotFound struct {
	Name string
}

func (e ErrTemplateNotFound) Error() string {
	return "template not found: " + e.Name
}
