package lib

import (
	"errors"
	"fmt"
	"net/url"
)

type Mapping struct {
	local  string
	remote string
}

func newMapping(local, remote string) (m Mapping, err error) {
	u, err := url.Parse(remote)

	if u.Scheme == "" {
		err = errors.New(fmt.Sprintf("%s: missing scheme", remote))
	} else if u.Scheme != "http" && u.Scheme != "https" {
		err = errors.New(fmt.Sprintf("%s: unsupported scheme", remote))
	}

	if err != nil {
		return
	}

	m = Mapping{
		local:  local,
		remote: remote,
	}

	return
}
