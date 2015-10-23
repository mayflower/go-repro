package lib

import (
	"bytes"
	"net/http"
	"regexp"
)

type genericResponseRewriter struct {
	mappings      []Mapping
	rewriteRoutes []*regexp.Regexp
}

func (r *genericResponseRewriter) Matches(request *http.Request, response *http.Response) bool {
	for _, route := range r.rewriteRoutes {
		if route.MatchString(request.RequestURI) {
			return true
		}
	}

	return false
}

func (r *genericResponseRewriter) RewriteResponse(response []byte) []byte {
	for _, mapping := range r.mappings {
		if bytes.Contains(response, []byte(mapping.remote)) {
			response = bytes.Replace(
				response, []byte(mapping.remote), []byte("http://"+mapping.local), -1)
		}
	}

	return response
}

func newGenericResponseRewriter(mappings []Mapping, rewriteRoutes []*regexp.Regexp) *genericResponseRewriter {
	return &genericResponseRewriter{
		mappings:      mappings,
		rewriteRoutes: rewriteRoutes,
	}
}
