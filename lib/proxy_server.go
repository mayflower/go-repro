package lib

import (
	"fmt"
	"io"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

type proxyServer struct {
	local  string
	remote string
	log    io.Writer

	server http.Server
	client http.Client
}

func init() {
	_ = spew.UnsafeDisabled
}

func (p *proxyServer) ServeHTTP(outgoing http.ResponseWriter, incoming *http.Request) {
	var err error

	proxyRequest, err := p.buildProxyRequest(incoming)
	var response *http.Response

	if err == nil {
		response, err = p.client.Do(proxyRequest)
	}

	if err != nil {
		http.Error(outgoing, err.Error(), http.StatusBadRequest)
	} else {
		p.proxyResponse(response, outgoing)
	}
}

func (p *proxyServer) buildProxyRequest(incoming *http.Request) (outgoing *http.Request, err error) {
	outgoing, err = http.NewRequest(incoming.Method, p.remote+"/"+incoming.RequestURI, incoming.Body)

	for key, values := range incoming.Header {
		key = http.CanonicalHeaderKey(key)

		for _, value := range values {
			outgoing.Header.Add(key, value)
		}
	}

	return
}

func (p *proxyServer) proxyResponse(response *http.Response, outgoing http.ResponseWriter) {
	outgoingHeaders := outgoing.Header()

	for key, values := range response.Header {
		key = http.CanonicalHeaderKey(key)

		for _, value := range values {
			outgoingHeaders.Add(key, value)
		}
	}

	outgoing.WriteHeader(response.StatusCode)

	io.Copy(outgoing, response.Body)

	response.Body.Close()
}

func (p *proxyServer) Start() <-chan error {
	c := make(chan error, 1)

	go func() {
		c <- p.server.ListenAndServe()
	}()

	fmt.Fprintf(p.log, "proxying requests for %s to %s\n", p.local, p.remote)

	return c
}

func newProxyServer(m Mapping, log io.Writer) (p *proxyServer, err error) {
	p = &proxyServer{
		local:  m.local,
		remote: m.remote,
		log:    log,
	}

	p.server = http.Server{
		Addr:    p.local,
		Handler: p,
	}

	p.client = http.Client{}

	return
}
