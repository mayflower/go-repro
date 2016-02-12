package lib

import (
	"bytes"
	"regexp"
)

type GenericBodyRewriter struct {
	rewriteRoutes []*regexp.Regexp
}

func (r *GenericBodyRewriter) Matches(ctx RequestContext) bool {
	request := ctx.IncomingRequest()

	if request.Header.Get("content-type") == "application/json" {
		return false
	}

	for _, route := range r.rewriteRoutes {
		if route.MatchString(ctx.RequestUrl()) {
			return true
		}
	}

	return false
}

func (*GenericBodyRewriter) RewriteResponse(response []byte, ctx RequestContext) []byte {
	rewritten := false

	for _, mapping := range ctx.HostMappings() {
		if bytes.Contains(response, []byte(mapping.remote)) {
			response = bytes.Replace(
				response, []byte(mapping.remote), []byte(mapping.local), -1)

			rewritten = true
		}
	}

	if rewritten {
		ctx.Log("generic body rewriter: body rewritten")
	}

	return response
}

func NewGenericResponseRewriter(rewriteRoutes []*regexp.Regexp) *GenericBodyRewriter {
	return &GenericBodyRewriter{
		rewriteRoutes: rewriteRoutes,
	}
}
