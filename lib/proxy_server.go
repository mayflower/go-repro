package lib

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/davecgh/go-spew/spew"
)

type redirectCaughtError struct{}

func (c redirectCaughtError) Error() string {
	return "redirect caught"
}

type proxyServer struct {
	local     string
	remote    string
	log       io.Writer
	rewriters []Rewriter
	mappings  []Mapping

	server http.Server
	client http.Client
}

func init() {
	_ = spew.UnsafeDisabled
}

func (p *proxyServer) ServeHTTP(outgoing http.ResponseWriter, incoming *http.Request) {
	var err error

	hostMappings := buildHostMappings(p.mappings, incoming.Host)
	for _, rewriter := range p.rewriters {
		rewriter.SetMappings(hostMappings)
	}

	proxyRequest, err := p.buildProxyRequest(incoming)
	var response *http.Response

	if err == nil {
		response, err = p.client.Do(proxyRequest)

		if isRedirectError(err) {
			err = nil
		}
	}

	if err != nil {
		fmt.Fprintf(p.log, "error during proxy request: %v\n", err)
		http.Error(outgoing, err.Error(), http.StatusBadGateway)
	} else {
		p.sendProxyResponse(incoming, response, outgoing)
	}
}

func isRedirectError(err error) (q bool) {
	urlError, q := err.(*url.Error)
	if !q {
		return
	}

	_, q = urlError.Err.(redirectCaughtError)

	return
}

func (p *proxyServer) buildProxyRequest(incoming *http.Request) (outgoing *http.Request, err error) {
	outgoing, err = http.NewRequest(incoming.Method, p.remote+incoming.RequestURI, incoming.Body)

	for key, values := range incoming.Header {
		for _, value := range values {
			outgoing.Header.Add(key, value)
		}
	}

	// We save compression for later
	outgoing.Header.Del("accept-encoding")

	for _, rewriter := range p.rewriters {
		if rewriter, ok := rewriter.(IncomingHeaderRewriter); ok {
			rewriter.RewriteIncomingHeaders(outgoing.Header)
		}
	}

	return
}

func (p *proxyServer) sendProxyResponse(request *http.Request, response *http.Response, outgoing http.ResponseWriter) {
	defer response.Body.Close()

	outgoingHeaders := outgoing.Header()

	for key, values := range response.Header {
		for _, value := range values {
			outgoingHeaders.Add(key, value)
		}
	}

	outgoingHeaders.Del("content-length")
	outgoingHeaders.Del("set-cookie")

	// We need to remove any domain information set on the cookies
	for _, cookie := range response.Cookies() {
		cookie.Domain = ""

		http.SetCookie(outgoing, cookie)
	}

	responseRewriters := make([]ResponseRewriter, 0, len(p.rewriters))

	for _, rewriter := range p.rewriters {
		if r, ok := rewriter.(ResponseRewriter); ok {
			if r.Matches(request, response) {
				responseRewriters = append(responseRewriters, r)
			}
		}

		if r, ok := rewriter.(HeaderRewriter); ok {
			r.RewriteHeaders(outgoingHeaders)
		}
	}

	outgoing.WriteHeader(response.StatusCode)

	if len(responseRewriters) > 0 {
		bodyData, err := ioutil.ReadAll(response.Body)

		if err == nil {
			for _, rewriter := range responseRewriters {
				bodyData = rewriter.RewriteResponse(bodyData)
			}
		} else {
			bodyData = make([]byte, 0)
		}

		io.Copy(outgoing, bytes.NewBuffer(bodyData))
	} else {
		io.Copy(outgoing, response.Body)
	}
}

func (p *proxyServer) Start() <-chan error {
	c := make(chan error, 1)

	go func() {
		c <- p.server.ListenAndServe()
	}()

	fmt.Fprintf(p.log, "proxying requests for %s to %s\n", p.local, p.remote)

	return c
}

func (p *proxyServer) AddRewriter(r Rewriter) {
	p.rewriters = append(p.rewriters, r)
}

func newProxyServer(m Mapping, mappings []Mapping, log io.Writer, sslAllowInsecure bool) (p *proxyServer, err error) {
	p = &proxyServer{
		local:     m.local,
		remote:    m.remote,
		log:       log,
		rewriters: make([]Rewriter, 0),
		mappings:  mappings,
	}

	p.server = http.Server{
		Addr:    p.local,
		Handler: p,
	}

	var tlsConfig *tls.Config
	if sslAllowInsecure {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	p.client = http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return redirectCaughtError{}
		},
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return
}
