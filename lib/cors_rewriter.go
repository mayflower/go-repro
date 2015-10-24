package lib

import (
	"net/http"
)

type CorsRewriter struct {
	GenericHeaderRewriter
}

func (c *CorsRewriter) RewriteIncomingHeaders(headers http.Header) {
	c.GenericHeaderRewriter.RewriteSpecifiedIncomingHeaders(
		[]string{"origin"}, headers)
}

func (c *CorsRewriter) RewriteHeaders(headers http.Header) {
	c.GenericHeaderRewriter.RewriteSpecifiedHeaders(
		[]string{"access-control-allow-origin"}, headers)
}

func NewCorsRewriter() *CorsRewriter {
	return &CorsRewriter{}
}
