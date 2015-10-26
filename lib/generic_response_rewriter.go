package lib

import (
	"bytes"
	"net/http"
	"regexp"
)

type GenericResponseRewriter struct {
	rewriteRoutes []*regexp.Regexp
}

func (r *GenericResponseRewriter) Matches(request *http.Request, response *http.Response) bool {
	if request.Header.Get("content-type") == "application/json" {
		return false
	}

	for _, route := range r.rewriteRoutes {
		if route.MatchString(request.RequestURI) {
			return true
		}
	}

	return false
}

func (*GenericResponseRewriter) RewriteResponse(response []byte, mappings []HostMapping) []byte {
	for _, mapping := range mappings {
		if bytes.Contains(response, []byte(mapping.remote)) {
			response = bytes.Replace(
				response, []byte(mapping.remote), []byte(mapping.local), -1)
		}
	}

	return response
}

func NewGenericResponseRewriter(rewriteRoutes []*regexp.Regexp) *GenericResponseRewriter {
	return &GenericResponseRewriter{
		rewriteRoutes: rewriteRoutes,
	}
}
