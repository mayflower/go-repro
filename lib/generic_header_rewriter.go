package lib

import (
	"net/http"
	"strings"
)

type GenericHeaderRewriter struct{}

func (*GenericHeaderRewriter) RewriteSpecifiedHeaders(keys []string, headers http.Header, mappings []HostMapping) {
	for _, key := range keys {

		if value := headers.Get(key); value != "" {
			for _, mapping := range mappings {
				if strings.Contains(value, mapping.remote) {
					value = strings.Replace(
						value, mapping.remote, mapping.local, -1)
				}
			}

			headers.Set(key, value)
		}
	}
}

func (*GenericHeaderRewriter) RewriteSpecifiedIncomingHeaders(keys []string, headers http.Header, mappings []HostMapping) {
	for _, key := range keys {

		if value := headers.Get(key); value != "" {
			for _, mapping := range mappings {
				if strings.Contains(value, mapping.local) {
					value = strings.Replace(
						value, mapping.local, mapping.remote, -1)
				}
			}

			headers.Set(key, value)
		}
	}
}
