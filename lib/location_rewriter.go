package lib

import (
	"net/http"
)

type LocationRewriter struct {
	GenericHeaderRewriter
}

func (l *LocationRewriter) RewriteHeaders(headers http.Header) {
	l.GenericHeaderRewriter.RewriteSpecifiedHeaders(
		[]string{"location"}, headers)
}

func NewLocationRewriter() *LocationRewriter {
	return &LocationRewriter{}
}
