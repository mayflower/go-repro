package lib

import (
	"net/http"
	"strings"
)

type GenericHeaderRewriter struct{}

func (*GenericHeaderRewriter) RewriteSpecifiedHeaders(keys []string, headers http.Header, ctx RequestContext) (rewritten bool) {
	for _, key := range keys {

		if value := headers.Get(key); value != "" {
			for _, mapping := range ctx.HostMappings() {
				if strings.Contains(value, mapping.remote) {
					value = strings.Replace(value, mapping.remote, mapping.local, -1)
					rewritten = true
				}
			}

			headers.Set(key, value)
		}
	}

	return
}

func (*GenericHeaderRewriter) RewriteSpecifiedIncomingHeaders(keys []string, headers http.Header, ctx RequestContext) (rewritten bool) {
	for _, key := range keys {

		if value := headers.Get(key); value != "" {
			for _, mapping := range ctx.HostMappings() {
				if strings.Contains(value, mapping.local) {
					value = strings.Replace(value, mapping.local, mapping.remote, -1)
					rewritten = true
				}
			}

			headers.Set(key, value)
		}
	}

	return
}
