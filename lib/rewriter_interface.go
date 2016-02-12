package lib

import (
	"net/http"
)

type RequestContext interface {
	IncomingRequest() *http.Request
	UpstreamResponse() *http.Response
	RequestUrl() string
	HostMappings() []HostMapping
	Log(message string)
}

type HeaderRewriter interface {
	RewriteHeaders(headers http.Header, ctx RequestContext)
}

type IncomingHeaderRewriter interface {
	RewriteIncomingHeaders(headers http.Header, ctx RequestContext)
}

type BodyRewriter interface {
	RewriteResponse(response []byte, ctx RequestContext) []byte
	Matches(ctx RequestContext) bool
}

type Rewriter interface{}
