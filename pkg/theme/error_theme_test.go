package theme_test

import (
	"bytes"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/neilotoole/slogt"
	"github.com/sablierapp/sablier/pkg/theme"
	"gotest.tools/v3/assert"
)

func TestErrorThemes_Render(t *testing.T) {
	t.Run("built-in fallback with empty name", func(t *testing.T) {
		themes, err := theme.NewErrorThemes(slogt.New(t))
		assert.NilError(t, err)

		buf := new(bytes.Buffer)
		err = themes.Render("", theme.ErrorOptions{
			StatusCode:      404,
			StatusText:      "Not Found",
			Title:           "Group not found",
			Detail:          "The group you requested does not exist.",
			RequestedGroup:  "missing",
			AvailableGroups: []string{"nginx"},
		}, buf)
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(buf.String(), "Group not found"))
		assert.Assert(t, strings.Contains(buf.String(), "missing"))
	})

	t.Run("per-theme override for embedded themes", func(t *testing.T) {
		themes, err := theme.NewErrorThemes(slogt.New(t))
		assert.NilError(t, err)

		buf := new(bytes.Buffer)
		err = themes.Render("shuffle", theme.ErrorOptions{
			StatusCode: 404,
			StatusText: "Not Found",
			Title:      "Group not found",
		}, buf)
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(buf.String(), "Group not found"))
	})

	t.Run("unregistered theme falls back to built-in", func(t *testing.T) {
		themes, err := theme.NewErrorThemes(slogt.New(t))
		assert.NilError(t, err)

		buf := new(bytes.Buffer)
		err = themes.Render("does-not-exist", theme.ErrorOptions{
			StatusCode: 404,
			StatusText: "Not Found",
			Title:      "Theme not found",
		}, buf)
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(buf.String(), "Theme not found"))
	})

	t.Run("custom theme error override", func(t *testing.T) {
		custom := `
<!DOCTYPE html>
<html><body>{{ .StatusCode }} {{ .Title }} custom-error</body></html>
`
		errorThemes, err := theme.NewErrorThemesWithCustomThemes(fstest.MapFS{
			"custom.error.html": &fstest.MapFile{Data: []byte(custom)},
		}, slogt.New(t))
		assert.NilError(t, err)

		buf := new(bytes.Buffer)
		err = errorThemes.Render("custom", theme.ErrorOptions{
			StatusCode: 404,
			StatusText: "Not Found",
			Title:      "Group not found",
		}, buf)
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(buf.String(), "custom-error"))

		// loading themes must not include error themes
		loadingThemes, err := theme.NewWithCustomThemes(fstest.MapFS{
			"custom.error.html": &fstest.MapFile{Data: []byte(custom)},
		}, slogt.New(t))
		assert.NilError(t, err)
		assert.Assert(t, !slices.Contains(loadingThemes.List(), "custom"))
		assert.Assert(t, !slices.Contains(loadingThemes.List(), "error"))
	})

	t.Run("nested custom theme error override", func(t *testing.T) {
		custom := `<!DOCTYPE html><html><body>{{ .Title }} nested-error</body></html>`
		errorThemes, err := theme.NewErrorThemesWithCustomThemes(fstest.MapFS{
			"nested/custom.error.html": &fstest.MapFile{Data: []byte(custom)},
		}, slogt.New(t))
		assert.NilError(t, err)

		buf := new(bytes.Buffer)
		err = errorThemes.Render("nested/custom", theme.ErrorOptions{
			StatusCode: 404,
			Title:      "Group not found",
		}, buf)
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(buf.String(), "nested-error"))
	})

	t.Run("custom error theme overrides embedded error theme", func(t *testing.T) {
		custom := `<!DOCTYPE html><html><body>overridden ghost error</body></html>`
		errorThemes, err := theme.NewErrorThemesWithCustomThemes(fstest.MapFS{
			"ghost.error.html": &fstest.MapFile{Data: []byte(custom)},
		}, slogt.New(t))
		assert.NilError(t, err)

		buf := new(bytes.Buffer)
		err = errorThemes.Render("ghost", theme.ErrorOptions{
			StatusCode: 404,
		}, buf)
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(buf.String(), "overridden ghost error"))
	})

	t.Run("broken error template surfaces the error", func(t *testing.T) {
		themes, err := theme.NewErrorThemesWithCustomThemes(fstest.MapFS{
			"broken.error.html": &fstest.MapFile{Data: []byte(`{{ .Missing.Field }}`)},
		}, slogt.New(t))
		assert.NilError(t, err)

		buf := new(bytes.Buffer)
		err = themes.Render("broken", theme.ErrorOptions{
			StatusCode: 404,
			Title:      "Group not found",
		}, buf)
		assert.ErrorContains(t, err, "can't evaluate field Missing")
	})

	t.Run("subdirectories and non-html files are ignored during walk", func(t *testing.T) {
		themes, err := theme.NewErrorThemesWithCustomThemes(fstest.MapFS{
			"dir.html/ignored.txt": &fstest.MapFile{Data: []byte("hello")},
			"backup.html.bak":      &fstest.MapFile{Data: []byte("backup")},
			"valid.error.html":     &fstest.MapFile{Data: []byte("<!DOCTYPE html><html><body>{{.Title}}</body></html>")},
		}, slogt.New(t))
		assert.NilError(t, err)

		buf := new(bytes.Buffer)
		err = themes.Render("valid", theme.ErrorOptions{
			StatusCode: 404,
			Title:      "Found Valid",
		}, buf)
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(buf.String(), "Found Valid"))
	})
}
