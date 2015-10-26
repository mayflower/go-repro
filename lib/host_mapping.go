package lib

import (
	"errors"
	"strings"
)

type HostMapping struct {
	local  string
	remote string
}

func buildHostMappings(mappings []Mapping, requestHostvar string) (hostMappings []HostMapping) {
	hostMappings = make([]HostMapping, 0, len(mappings))

	requestHost, _, requestErr := splitHostPort(requestHostvar)

	for _, mapping := range mappings {
		h := HostMapping{
			remote: mapping.remote,
		}

		localHost, localPort, localErr := splitHostPort(mapping.local)

		if localErr == nil && requestErr == nil && localHost == "0.0.0.0" {
			h.local = "http://" + requestHost + ":" + localPort
		} else {
			h.local = "http://" + mapping.local
		}

		hostMappings = append(hostMappings, h)
	}

	return
}

func splitHostPort(addr string) (host, port string, err error) {
	parts := strings.Split(addr, ":")

	if len(parts) != 2 {
		err = errors.New("")
		return
	}

	host = parts[0]
	port = parts[1]

	return
}
