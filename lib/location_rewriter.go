package lib

import (
	"net/http"
	"strings"
)

type LocationRewriter struct {
	mappings []hostMapping
}

func (l *LocationRewriter) RewriteHeaders(headers http.Header) {
	if location := headers.Get("location"); location != "" {
		for _, mapping := range l.mappings {
			if strings.Contains(location, mapping.remote) {
				location = strings.Replace(
					location, mapping.remote, mapping.local, -1)
			}
		}

		headers.Set("location", location)
	}
}

func (l *LocationRewriter) SetMappings(mappings []hostMapping) {
	l.mappings = mappings
}

func NewLocationRewriter() *LocationRewriter {
	return &LocationRewriter{}
}
