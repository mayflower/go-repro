package lib

import (
	"net/http"
)

type CorsRewriter struct {
	GenericHeaderRewriter
}

func (c *CorsRewriter) RewriteIncomingHeaders(headers http.Header, mappings []HostMapping) {
	c.GenericHeaderRewriter.RewriteSpecifiedIncomingHeaders(
		[]string{"origin"}, headers, mappings)
}

func (c *CorsRewriter) RewriteHeaders(headers http.Header, mappings []HostMapping) {
	c.GenericHeaderRewriter.RewriteSpecifiedHeaders(
		[]string{"access-control-allow-origin"}, headers, mappings)
}

func NewCorsRewriter() *CorsRewriter {
	return &CorsRewriter{}
}
