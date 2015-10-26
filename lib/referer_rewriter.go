package lib

import (
	"net/http"
)

type RefererRewriter struct {
	GenericHeaderRewriter
}

func (r *RefererRewriter) RewriteIncomingHeaders(headers http.Header, ctx RequestContext) {
	if r.GenericHeaderRewriter.RewriteSpecifiedIncomingHeaders([]string{"referer"}, headers, ctx) {
		ctx.Log("rewrote referer")
	}
}

func NewRefererRewriter() *RefererRewriter {
	return &RefererRewriter{}
}
