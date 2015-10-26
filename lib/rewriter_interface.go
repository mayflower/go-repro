package lib

import (
	"net/http"
)

type HeaderRewriter interface {
	RewriteHeaders(headers http.Header, mappings []HostMapping)
}

type IncomingHeaderRewriter interface {
	RewriteIncomingHeaders(headers http.Header, mappings []HostMapping)
}

type ResponseRewriter interface {
	RewriteResponse(response []byte, mappings []HostMapping) []byte
	Matches(*http.Request, *http.Response) bool
}

type Rewriter interface{}
