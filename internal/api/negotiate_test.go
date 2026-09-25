package api

import "testing"

var (
	testHTMLFamily = []string{"text/html", "application/xhtml+xml"}
	testJSONFamily = []string{"application/json", "application/problem+json"}
)

func TestPrefers(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   bool
	}{
		{
			name:   "browser navigation",
			accept: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			want:   true,
		},
		{
			name:   "wildcard only",
			accept: "*/*",
			want:   false,
		},
		{
			name:   "application json only",
			accept: "application/json",
			want:   false,
		},
		{
			name:   "empty header",
			accept: "",
			want:   false,
		},
		{
			name:   "json preferred over html",
			accept: "application/json;q=0.9,text/html;q=0.5",
			want:   false,
		},
		{
			name:   "wildcard preferred over low html",
			accept: "text/html;q=0.5,*/*;q=0.8",
			want:   false,
		},
		{
			name:   "text star matches text html",
			accept: "text/*",
			want:   true,
		},
		{
			name:   "text html explicitly refused",
			accept: "text/html;q=0",
			want:   false,
		},
		{
			name:   "problem json only",
			accept: "application/problem+json",
			want:   false,
		},
		{
			name:   "xhtml only",
			accept: "application/xhtml+xml",
			want:   true,
		},
		{
			name:   "invalid html quality is ignored",
			accept: "text/html;q=bogus,application/json;q=0.5",
			want:   false,
		},
		{
			name:   "out of range html quality is ignored",
			accept: "text/html;q=2,application/json;q=0.5",
			want:   false,
		},
		{
			name:   "tie between families",
			accept: "text/html,application/json",
			want:   false,
		},
		{
			name:   "image only",
			accept: "image/avif,image/webp,image/png;q=0.9,*/*;q=0.5",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prefers(tt.accept, testHTMLFamily, testJSONFamily)
			if got != tt.want {
				t.Errorf("prefers(%q) = %v, want %v", tt.accept, got, tt.want)
			}
		})
	}
}

func TestQfactor(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		mime   string
		want   float64
	}{
		{
			name:   "empty header",
			accept: "",
			mime:   "text/html",
			want:   0,
		},
		{
			name:   "no matching range",
			accept: "application/json",
			mime:   "text/html",
			want:   0,
		},
		{
			name:   "exact match with default q",
			accept: "text/html",
			mime:   "text/html",
			want:   1,
		},
		{
			name:   "exact match with explicit q",
			accept: "text/html;q=0.5",
			mime:   "text/html",
			want:   0.5,
		},
		{
			name:   "type wildcard",
			accept: "text/*;q=0.7",
			mime:   "text/html",
			want:   0.7,
		},
		{
			name:   "any wildcard",
			accept: "*/*;q=0.2",
			mime:   "image/png",
			want:   0.2,
		},
		{
			name:   "most specific range wins",
			accept: "*/*;q=0.2,text/*;q=0.5,text/html;q=0.8",
			mime:   "text/html",
			want:   0.8,
		},
		{
			name:   "type wildcard beats any wildcard",
			accept: "*/*;q=0.9,text/*;q=0.5",
			mime:   "text/html",
			want:   0.5,
		},
		{
			name:   "explicit q=0 refuses",
			accept: "text/html;q=0",
			mime:   "text/html",
			want:   0,
		},
		{
			name:   "explicit q=0 overrides wildcard",
			accept: "text/html;q=0,*/*;q=0.9",
			mime:   "text/html",
			want:   0,
		},
		{
			name:   "application wildcard does not match text",
			accept: "application/*;q=0.9",
			mime:   "text/html",
			want:   0,
		},
		{
			name:   "unparseable range skipped",
			accept: "not a media type,text/html",
			mime:   "text/html",
			want:   1,
		},
		{
			name:   "equal specificity picks higher q (ascending)",
			accept: "text/html;q=0.3,text/html;q=0.8",
			mime:   "text/html",
			want:   0.8,
		},
		{
			name:   "equal specificity picks higher q (descending)",
			accept: "text/html;q=0.8,text/html;q=0.3",
			mime:   "text/html",
			want:   0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := qfactor(parseAccept(tt.accept), tt.mime)
			if got != tt.want {
				t.Errorf("qfactor(%q, %q) = %v, want %v", tt.accept, tt.mime, got, tt.want)
			}
		})
	}
}
