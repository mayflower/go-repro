package lib

import (
	"net/http"
)

type LocationRewriter struct {
	GenericHeaderRewriter
}

func (l *LocationRewriter) RewriteHeaders(headers http.Header, mappings []HostMapping) {
	l.GenericHeaderRewriter.RewriteSpecifiedHeaders(
		[]string{"location"}, headers, mappings)
}

func NewLocationRewriter() *LocationRewriter {
	return &LocationRewriter{}
}
