package api

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sablierapp/sablier/pkg/sablier"
	"github.com/sablierapp/sablier/pkg/theme"
	"github.com/tniswong/go.rfcx/rfc7807"
)

// htmlFamily and jsonFamily are the media types the dynamic strategy can
// produce for an error: a themed HTML page or an RFC 7807 JSON problem.
var (
	htmlFamily = []string{"text/html", "application/xhtml+xml"}
	jsonFamily = []string{"application/json", "application/problem+json"}
)

// AbortDynamic ends a dynamic-strategy request with a content-negotiated problem.
// When the client prefers HTML over JSON (per the Accept header) and error themes
// are configured, a themed HTML error page is served; otherwise the RFC 7807
// JSON problem is returned, preserving the API contract for non-browser clients.
func AbortDynamic(c *gin.Context, s *ServeStrategy, themeName string, problem rfc7807.Problem, opts theme.ErrorOptions) {
	c.Header("Vary", "Accept")

	acceptHeader := strings.Join(c.Request.Header.Values("Accept"), ", ")
	if !prefers(acceptHeader, htmlFamily, jsonFamily) || s.ErrorTheme == nil {
		AbortWithProblemDetail(c, problem)
		return
	}

	if opts.StatusCode == 0 {
		opts.StatusCode = problem.Status
	}
	if opts.StatusText == "" {
		opts.StatusText = http.StatusText(problem.Status)
	}
	if opts.Title == "" {
		opts.Title = problem.Title
	}
	if opts.Detail == "" {
		opts.Detail = problem.Detail
	}

	// Render into a plain buffer first: rendering must fully succeed before any
	// byte reaches the client, so a broken error template surfaces as a 500
	// problem instead of a 404 with a truncated page.
	buf := new(bytes.Buffer)
	if err := s.ErrorTheme.Render(themeName, opts, buf); err != nil {
		AbortWithProblemDetail(c, ProblemError(err))
		return
	}

	_ = c.Error(problem)
	c.Abort()
	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Length", strconv.Itoa(buf.Len()))
	c.Writer.WriteHeader(problem.Status)
	if _, err := c.Writer.Write(buf.Bytes()); err != nil {
		_ = c.Error(err)
		return
	}
}

// AbortDynamicGroupNotFound is a convenience helper that builds both the RFC 7807
// problem and theme.ErrorOptions for a missing group.
func AbortDynamicGroupNotFound(c *gin.Context, s *ServeStrategy, themeName string, err sablier.ErrGroupNotFound) {
	AbortDynamic(c, s, themeName, ProblemGroupNotFound(err), theme.ErrorOptions{
		RequestedGroup:  err.Group,
		AvailableGroups: err.AvailableGroups,
	})
}

// AbortDynamicThemeNotFound is a convenience helper that builds both the RFC 7807
// problem and theme.ErrorOptions for a missing theme.
func AbortDynamicThemeNotFound(c *gin.Context, s *ServeStrategy, err theme.ErrThemeNotFound) {
	AbortDynamic(c, s, "", ProblemThemeNotFound(err), theme.ErrorOptions{
		RequestedTheme:  err.Theme,
		AvailableThemes: err.AvailableThemes,
	})
}
