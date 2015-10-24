package lib

import (
	"bytes"
	"net/http"
	"regexp"
)

type GenericResponseRewriter struct {
	mappings      []hostMapping
	rewriteRoutes []*regexp.Regexp
}

func (r *GenericResponseRewriter) SetMappings(mappings []hostMapping) {
	r.mappings = mappings
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

func (r *GenericResponseRewriter) RewriteResponse(response []byte) []byte {
	for _, mapping := range r.mappings {
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
