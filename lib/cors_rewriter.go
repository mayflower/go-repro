package lib

import (
	"net/http"
)

type CorsRewriter struct {
	GenericHeaderRewriter
}

func (c *CorsRewriter) RewriteIncomingHeaders(headers http.Header, ctx RequestContext) {
	if c.GenericHeaderRewriter.RewriteSpecifiedIncomingHeaders([]string{"origin"}, headers, ctx) {
		ctx.Log("rewrote origin")
	}
}

func (c *CorsRewriter) RewriteHeaders(headers http.Header, ctx RequestContext) {
	if c.GenericHeaderRewriter.RewriteSpecifiedHeaders([]string{"access-control-allow-origin"}, headers, ctx) {
		ctx.Log("rewrote access-control-allow-origin")
	}
}

func NewCorsRewriter() *CorsRewriter {
	return &CorsRewriter{}
}
