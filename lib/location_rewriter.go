package lib

import (
	"net/http"
)

type LocationRewriter struct {
	GenericHeaderRewriter
}

func (l *LocationRewriter) RewriteHeaders(headers http.Header, ctx RequestContext) {
	if l.GenericHeaderRewriter.RewriteSpecifiedHeaders([]string{"location"}, headers, ctx) {
		ctx.Log("rewrote location")
	}
}

func NewLocationRewriter() *LocationRewriter {
	return &LocationRewriter{}
}
