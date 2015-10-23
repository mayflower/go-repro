package lib

import (
	"net/http"
	"strings"
)

type locationRewriter struct {
	mappings []Mapping
}

func (l *locationRewriter) RewriteHeaders(headers http.Header) {
	if location := headers.Get("location"); location != "" {
		for _, mapping := range l.mappings {
			if strings.Contains(location, mapping.remote) {
				location = strings.Replace(
					location, mapping.remote, "http://"+mapping.local, -1)
			}
		}

		headers.Set("location", location)
	}
}

func newLocationRewriter(mappings []Mapping) *locationRewriter {
	return &locationRewriter{
		mappings: mappings,
	}
}
