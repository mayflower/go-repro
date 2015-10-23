package lib

import (
	"bytes"
	"net/http"
	"regexp"
)

type GenericResponseRewriter struct {
	mappings      []Mapping
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

func (r *GenericResponseRewriter) RewriteResponse(response []byte) []byte {
	for _, mapping := range r.mappings {
		if bytes.Contains(response, []byte(mapping.remote)) {
			response = bytes.Replace(
				response, []byte(mapping.remote), []byte("http://"+mapping.local), -1)
		}
	}

	return response
}

func NewGenericResponseRewriter(mappings []Mapping, rewriteRoutes []*regexp.Regexp) *GenericResponseRewriter {
	return &GenericResponseRewriter{
		mappings:      mappings,
		rewriteRoutes: rewriteRoutes,
	}
}
