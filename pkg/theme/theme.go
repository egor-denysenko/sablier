package theme

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

// List of built-it themes
//
//go:embed embedded/*.html
var embeddedThemesFS embed.FS

type Themes struct {
	themes      *template.Template
	errorThemes *ErrorThemes
	l           *slog.Logger
}

func New(logger *slog.Logger) (*Themes, error) {
	themes := &Themes{
		themes: template.New("root"),
		l:      logger,
	}

	err := themes.ParseTemplatesFS(embeddedThemesFS)
	if err != nil {
		// Should never happen
		logger.Error("could not parse embedded templates", slog.Any("reason", err))
		return nil, err
	}

	errorThemes, err := NewErrorThemes(logger)
	if err != nil {
		return nil, err
	}
	themes.errorThemes = errorThemes

	return themes, nil
}

func NewWithCustomThemes(custom fs.FS, logger *slog.Logger) (*Themes, error) {
	themes, err := New(logger)
	if err != nil {
		return nil, err
	}

	err = themes.ParseAndBundleTemplatesFS(custom)
	if err != nil {
		logger.Error("could not parse custom templates", slog.Any("reason", err))
		return nil, err
	}
	if err = themes.errorThemes.ParseAndBundleTemplatesFS(custom); err != nil {
		logger.Error("could not parse custom error templates", slog.Any("reason", err))
		return nil, err
	}

	return themes, nil
}

// NewWithCustomThemesFromPath loads custom themes from a directory on the OS
// filesystem. It is the preferred constructor for production use because it
// wraps the directory in a noSymlinkFS that prevents symlinked files from
// being followed and inlined into the served HTML.
func (t *Themes) ErrorThemes() *ErrorThemes {
	return t.errorThemes
}

func NewWithCustomThemesFromPath(dirPath string, logger *slog.Logger) (*Themes, error) {
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
	return NewWithCustomThemes(custom, logger)
}
