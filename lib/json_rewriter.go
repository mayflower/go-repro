package lib

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

type JsonRewriter struct {
	rewriteRoutes []*regexp.Regexp
}

func (r *JsonRewriter) Matches(request *http.Request, response *http.Response) bool {
	if response.Header.Get("content-type") != "application/json" {
		return false
	}

	for _, route := range r.rewriteRoutes {
		if route.MatchString(request.RequestURI) {
			return true
		}
	}

	return false
}

func (r *JsonRewriter) RewriteResponse(response []byte, mappings []HostMapping) []byte {
	var err error

	stack := make([]interface{}, 0, 50)

	var unmarshalledResponse interface{}
	err = json.Unmarshal(response, &unmarshalledResponse)
	if err != nil {
		return response
	}

	if responseString, ok := unmarshalledResponse.(string); ok {
		unmarshalledResponse = r.stringReplace(responseString, mappings)
	} else {
		stack = append(stack, unmarshalledResponse)
	}

	for len(stack) > 0 {
		elt := stack[0]
		stack = stack[1:]

		switch elt := elt.(type) {
		case []interface{}:
			for i, value := range elt {
				switch value := value.(type) {
				case string:
					elt[i] = r.stringReplace(value, mappings)

				case []interface{}:
					stack = append(stack, value)

				case map[string]interface{}:
					stack = append(stack, value)
				}
			}

		case map[string]interface{}:
			for key, value := range elt {
				switch value := value.(type) {
				case string:
					elt[key] = r.stringReplace(value, mappings)

				case []interface{}:
					stack = append(stack, value)

				case map[string]interface{}:
					stack = append(stack, value)
				}

			}
		}
	}

	filteredResponse, err := json.Marshal(unmarshalledResponse)

	if err != nil {
		return response
	} else {
		return filteredResponse
	}
}

func (*JsonRewriter) stringReplace(in string, mappings []HostMapping) string {
	for _, mapping := range mappings {
		in = strings.Replace(in, mapping.remote, mapping.local, -1)
	}

	return in
}

func NewJsonRewriter(rewriteRoutes []*regexp.Regexp) *JsonRewriter {
	return &JsonRewriter{
		rewriteRoutes: rewriteRoutes,
	}
}
