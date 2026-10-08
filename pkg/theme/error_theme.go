package theme

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/sablierapp/sablier/pkg/version"
)

type ErrorThemes struct {
	templates *template.Template
	l         *slog.Logger
}

func NewErrorThemes(logger *slog.Logger) (*ErrorThemes, error) {
	et := &ErrorThemes{
		templates: template.New("root-errors"),
		l:         logger,
	}

	err := et.ParseTemplatesFS(embeddedThemesFS)
	if err != nil {
		logger.Error("could not parse embedded error templates", slog.Any("reason", err))
		return nil, err
	}

	return et, nil
}

func NewErrorThemesWithCustomThemes(custom fs.FS, logger *slog.Logger) (*ErrorThemes, error) {
	et, err := NewErrorThemes(logger)
	if err != nil {
		return nil, err
	}

	err = et.ParseAndBundleTemplatesFS(custom)
	if err != nil {
		logger.Error("could not parse custom error templates", slog.Any("reason", err))
		return nil, err
	}

	return et, nil
}

func NewErrorThemesWithCustomThemesFromPath(dirPath string, logger *slog.Logger) (*ErrorThemes, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("invalid custom themes path %q: %w", dirPath, err)
	}
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("custom themes path %q is not accessible: %w", absPath, err)
	}
	custom := noSymlinkFS{
		FS:   os.DirFS(absPath),
		root: absPath,
	}
	return NewErrorThemesWithCustomThemes(custom, logger)
}

func (et *ErrorThemes) ParseTemplatesFS(f fs.FS) error {
	return parseTemplatesFS(f, et.templates, et.l, "error theme", isErrorTemplateEntry)
}

func (et *ErrorThemes) ParseAndBundleTemplatesFS(f fs.FS) error {
	return parseAndBundleTemplatesFS(f, et.templates, et.l, "error theme", isErrorTemplateEntry)
}

// Render renders an error page. If a theme named name is loaded and ships
// a <name>.error.html template it is used; otherwise the built-in error.html
// fallback is rendered.
func (et *ErrorThemes) Render(name string, opts ErrorOptions, writer io.Writer) error {
	if !slices.IsSorted(opts.AvailableGroups) {
		opts.AvailableGroups = slices.Clone(opts.AvailableGroups)
		slices.Sort(opts.AvailableGroups)
	}
	if !slices.IsSorted(opts.AvailableThemes) {
		opts.AvailableThemes = slices.Clone(opts.AvailableThemes)
		slices.Sort(opts.AvailableThemes)
	}

	options := ErrorTemplateData{
		ErrorOptions: opts,
		Version:      version.Version,
	}

	var tpl *template.Template
	if name != "" {
		tpl = et.templates.Lookup(name + ".error.html")
		if tpl == nil {
			et.l.Info("per-theme error template not found, falling back to default", slog.String("theme", name))
		}
	}
	if tpl == nil {
		tpl = et.templates.Lookup("error.html")
	}
	if tpl == nil {
		return fmt.Errorf("fallback error template %q is missing from error template registry", "error.html")
	}

	return tpl.Execute(writer, options)
}
