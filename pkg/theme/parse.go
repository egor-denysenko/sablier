package theme

import (
	"html/template"
	"io/fs"
	"log/slog"
	"path"
	"strings"
)

// isErrorTemplate reports whether name (a template file path or base name)
// should be registered as an error-page template rather than a loading-page
// theme. Error templates end with ".error.html"; the built-in "error.html"
// fallback is an error template too.
func isErrorTemplate(name string) bool {
	base := path.Base(name)
	return base == "error.html" || strings.HasSuffix(base, ".error.html")
}

func isLoadingTemplate(filePath string, d fs.DirEntry) bool {
	return d != nil && !d.IsDir() && path.Ext(filePath) == ".html" && !isErrorTemplate(filePath)
}

func isErrorTemplateEntry(filePath string, d fs.DirEntry) bool {
	return d != nil && !d.IsDir() && isErrorTemplate(filePath)
}

func parseTemplatesFS(f fs.FS, tmpl *template.Template, l *slog.Logger, kind string, filter func(string, fs.DirEntry) bool) error {
	return fs.WalkDir(f, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !filter(filePath, d) {
			return nil
		}
		l.Info(kind+" found", slog.String("path", filePath))
		_, err = tmpl.ParseFS(f, filePath)
		if err != nil {
			l.Info("cannot add "+kind, slog.String("path", filePath), slog.Any("reason", err))
			return err
		}
		l.Info("successfully added "+kind, slog.String("path", filePath))
		return nil
	})
}

func parseAndBundleTemplatesFS(f fs.FS, tmpl *template.Template, l *slog.Logger, kind string, filter func(string, fs.DirEntry) bool) error {
	return fs.WalkDir(f, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !filter(filePath, d) {
			return nil
		}
		l.Info(kind+" found", slog.String("path", filePath))
		content, err := bundleHTML(f, filePath)
		if err != nil {
			l.Info("cannot bundle "+kind, slog.String("path", filePath), slog.Any("reason", err))
			return err
		}
		// Relative paths are unique, so the only name a custom theme can already
		// occupy is that of an embedded theme, which it deliberately replaces.
		name := path.Clean(filePath)
		if tmpl.Lookup(name) != nil {
			l.Info(kind+" overrides an embedded "+kind, slog.String("path", filePath), slog.String("name", strings.TrimSuffix(name, ".html")))
		}
		if _, err = tmpl.New(name).Parse(content); err != nil {
			l.Info("cannot add "+kind, slog.String("path", filePath), slog.Any("reason", err))
			return err
		}
		l.Info("successfully added "+kind, slog.String("path", filePath), slog.String("name", strings.TrimSuffix(name, ".html")))
		return nil
	})
}

func (t *Themes) ParseTemplatesFS(f fs.FS) error {
	return parseTemplatesFS(f, t.themes, t.l, "theme", isLoadingTemplate)
}

// ParseAndBundleTemplatesFS walks f and registers every .html file as a named
// template, inlining relative CSS, JS, and image assets so that the resulting
// template is fully self-contained (see bundleHTML for details).
//
// A theme is named after its path relative to the themes directory, without the
// ".html" extension, so that "special/secret.html" is requested as
// "special/secret". Naming themes after their base name instead let two files
// in different sub-directories silently overwrite each other.
func (t *Themes) ParseAndBundleTemplatesFS(f fs.FS) error {
	return parseAndBundleTemplatesFS(f, t.themes, t.l, "theme", isLoadingTemplate)
}
