package lib

import (
	"net/http"
)

type RefererRewriter struct {
	GenericHeaderRewriter
}

func (r *RefererRewriter) RewriteIncomingHeaders(headers http.Header, mappings []HostMapping) {
	r.GenericHeaderRewriter.RewriteSpecifiedIncomingHeaders(
		[]string{"referer"}, headers, mappings)
}

func NewRefererRewriter() *RefererRewriter {
	return &RefererRewriter{}
}
