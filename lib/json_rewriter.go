package lib

import (
	"encoding/json"
	"regexp"
	"strings"
)

type JsonRewriter struct {
	rewriteRoutes []*regexp.Regexp
}

func (r *JsonRewriter) Matches(ctx RequestContext) bool {
	request := ctx.IncomingRequest()
	response := ctx.UpstreamResponse()

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

func (r *JsonRewriter) RewriteResponse(response []byte, ctx RequestContext) []byte {
	var err error

	rewritten := false

	stack := make([]interface{}, 0, 50)

	var unmarshalledResponse interface{}
	err = json.Unmarshal(response, &unmarshalledResponse)
	if err != nil {
		return response
	}

	if responseString, ok := unmarshalledResponse.(string); ok {
		unmarshalledResponse = r.stringReplace(responseString, ctx, &rewritten)
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
					elt[i] = r.stringReplace(value, ctx, &rewritten)

				case []interface{}:
					stack = append(stack, value)

				case map[string]interface{}:
					stack = append(stack, value)
				}
			}

		case map[string]interface{}:
			for key, value := range elt {

				rewriteKey := false
				newKey := r.stringReplace(key, ctx, &rewriteKey)

				if _, ok := elt[newKey]; ok {
					rewriteKey = false
				}

				if rewriteKey {
					delete(elt, key)
					elt[newKey] = value
				}

				rewritten = rewritten || rewriteKey

				switch value := value.(type) {
				case string:
					elt[newKey] = r.stringReplace(value, ctx, &rewritten)

				case []interface{}:
					stack = append(stack, value)

				case map[string]interface{}:
					stack = append(stack, value)
				}

			}
		}
	}

	var filteredResponse []byte
	if rewritten {
		filteredResponse, err = json.Marshal(unmarshalledResponse)
	}

	if filteredResponse != nil && err == nil {
		ctx.Log("json rewriter: response rewritten")
		return filteredResponse
	} else {
		return response
	}
}

func (*JsonRewriter) stringReplace(in string, ctx RequestContext, rewritten *bool) string {
	for _, mapping := range ctx.HostMappings() {
		if strings.Contains(in, mapping.remote) {
			in = strings.Replace(in, mapping.remote, mapping.local, -1)
			*rewritten = true
		}
	}

	return in
}

func NewJsonRewriter(rewriteRoutes []*regexp.Regexp) *JsonRewriter {
	return &JsonRewriter{
		rewriteRoutes: rewriteRoutes,
	}
}
