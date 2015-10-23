package lib

import (
	"net/http"
	"strings"
)

type LocationRewriter struct {
	mappings []Mapping
}

func (l *LocationRewriter) RewriteHeaders(headers http.Header) {
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

func NewLocationRewriter(mappings []Mapping) *LocationRewriter {
	return &LocationRewriter{
		mappings: mappings,
	}
}
