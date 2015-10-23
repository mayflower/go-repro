package lib

import (
	"net/http"
)

type HeaderRewriter interface {
	Rewrite(http.Header)
}

type Rewriter interface {
}
