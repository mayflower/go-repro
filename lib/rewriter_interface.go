package lib

import (
	"net/http"
)

type HeaderRewriter interface {
	RewriteHeaders(headers http.Header)
}

type IncomingHeaderRewriter interface {
	RewriteIncomingHeaders(headers http.Header)
}

type ResponseRewriter interface {
	RewriteResponse(response []byte) []byte
	Matches(*http.Request, *http.Response) bool
}

type Rewriter interface {
	SetMappings(mappings []hostMapping)
}
