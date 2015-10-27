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

func NewMapping(local, remote string) (m Mapping, err error) {
	if len(remote) > 0 && remote[len(remote)-1] == '/' {
		remote = remote[:len(remote)-1]
	}

	err = validateRemote(remote)

	if err != nil {
		return
	}

	m = Mapping{
		local:  local,
		remote: remote,
	}

	return
}

func validateRemote(remote string) (err error) {
	u, err := url.Parse(remote)

	if u.Scheme == "" {
		err = errors.New(fmt.Sprintf("%s: missing scheme", remote))
	} else if u.Scheme != "http" && u.Scheme != "https" {
		err = errors.New(fmt.Sprintf("%s: unsupported scheme", remote))
	}

	if u.Path != "" {
		err = errors.New(fmt.Sprintf("%s: must not have a nontrivial path", remote))
	}

	return
}
