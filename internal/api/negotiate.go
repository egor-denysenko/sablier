package api

import (
	"mime"
	"strconv"
	"strings"
)

// accepts parses an Accept header value (as sent by the client; multiple header
// lines must already be joined with a comma) into a list of media ranges with
// their quality weights, per RFC 9110 §12.5.1.
//
// A missing q parameter defaults to 1. Ranges that fail to parse are skipped.
type acceptRange struct {
	mediaRange string // e.g. "text/html", "text/*", "*/*", "application/json"
	q          float64
}

func parseAccept(header string) []acceptRange {
	raw := strings.Split(header, ",")
	ranges := make([]acceptRange, 0, len(raw))
	for _, r := range raw {
		mt, params, err := mime.ParseMediaType(strings.TrimSpace(r))
		if err != nil || mt == "" {
			continue
		}
		weight := 1.0
		if q, ok := params["q"]; ok {
			qv, err := strconv.ParseFloat(q, 64)
			if err != nil || qv < 0 || qv > 1 {
				continue
			}
			weight = qv
		}
		ranges = append(ranges, acceptRange{mediaRange: mt, q: weight})
	}
	return ranges
}

// rangeSpecificity ranks a media range by precedence: exact type/subtype is more
// specific than type/*, which is more specific than */*. Media range parameters
// are not used for matching (only the q parameter is honored), so exact-type
// matches with different parameters are treated equally.
func rangeSpecificity(mr string) int {
	switch {
	case mr == "*/*":
		return 0
	case strings.HasSuffix(mr, "/*"):
		return 1
	default:
		return 2
	}
}

// matches reports whether the media range mr matches the given media type mime,
// honoring the RFC wildcard rules (*/* and type/*).
func matches(mr, mime string) bool {
	if mr == "*/*" {
		return true
	}
	if strings.HasSuffix(mr, "/*") {
		return strings.HasPrefix(mime, mr[:len(mr)-1])
	}
	return mr == mime
}

// qfactor returns the quality value the client assigns to mime, per RFC 9110
// §12.5.1: the q of the most specific media range matching mime. A media type
// with no matching range is unacceptable (0).
func qfactor(ranges []acceptRange, mime string) float64 {
	best := -1
	q := 0.0
	for _, r := range ranges {
		if !matches(r.mediaRange, mime) {
			continue
		}
		spec := rangeSpecificity(r.mediaRange)
		if spec > best || (spec == best && r.q > q) {
			best = spec
			q = r.q
		}
	}
	if best == -1 {
		return 0
	}
	return q
}

// prefers reports whether the client's Accept header signals that it prefers a
// representation from primary over one from fallback, honoring q-values: the
// best quality value over primary must strictly beat the best over fallback.
//
// It is deliberately conservative: an empty header, "*/*", or a tie yields false,
// so callers keep their existing default representation.
func prefers(accept string, primary, fallback []string) bool {
	ranges := parseAccept(accept)

	primaryQ := 0.0
	for _, mt := range primary {
		if q := qfactor(ranges, mt); q > primaryQ {
			primaryQ = q
		}
	}

	fallbackQ := 0.0
	for _, mt := range fallback {
		if q := qfactor(ranges, mt); q > fallbackQ {
			fallbackQ = q
		}
	}

	return primaryQ > fallbackQ
}
