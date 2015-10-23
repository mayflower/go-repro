package lib

import (
	"net/http"
	"strings"
)

type locationRewriter struct {
	mappings []Mapping
}

func (l *locationRewriter) Rewrite(headers http.Header) {
	if location := headers.Get("location"); location != "" {
		for _, mapping := range l.mappings {
			location = strings.Replace(
				location, mapping.remote, "http://"+mapping.local, -1)
		}

		headers.Set("location", location)
	}
}

func newLocationRewriter(mappings []Mapping) *locationRewriter {
	return &locationRewriter{
		mappings: mappings,
	}
}
