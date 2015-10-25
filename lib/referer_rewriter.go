package lib

import (
	"net/http"
)

type RefererRewriter struct {
	GenericHeaderRewriter
}

func (r *RefererRewriter) RewriteIncomingHeaders(headers http.Header) {
	r.GenericHeaderRewriter.RewriteSpecifiedIncomingHeaders(
		[]string{"referer"}, headers)
}

func NewRefererRewriter() *RefererRewriter {
	return &RefererRewriter{}
}
