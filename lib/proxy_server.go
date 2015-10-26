package lib

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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

func (p *proxyServer) ServeHTTP(outgoing http.ResponseWriter, incoming *http.Request) {
	var err error

	hostMappings := buildHostMappings(p.mappings, incoming.Host)

	upstreamRequest, err := p.buildUpstreamRequest(incoming, hostMappings)
	var response *http.Response

	if err == nil {
		response, err = p.client.Do(upstreamRequest)

		if isRedirectError(err) {
			err = nil
		}
	}

	if err != nil {
		fmt.Fprintf(p.log, "error during proxy request: %v\n", err)
		http.Error(outgoing, err.Error(), http.StatusBadGateway)
	} else {
		p.sendResponse(incoming, response, outgoing, hostMappings)
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

func (p *proxyServer) buildUpstreamRequest(incoming *http.Request, mappings []HostMapping) (outgoing *http.Request, err error) {
	outgoing, err = http.NewRequest(incoming.Method, p.remote+incoming.RequestURI, incoming.Body)

	for key, values := range incoming.Header {
		for _, value := range values {
			outgoing.Header.Add(key, value)
		}
	}

	for _, rewriter := range p.rewriters {
		if rewriter, ok := rewriter.(IncomingHeaderRewriter); ok {
			rewriter.RewriteIncomingHeaders(outgoing.Header, mappings)
		}
	}

	// We also try to compress upstream communication
	outgoing.Header.Set("accept-encoding", "gzip")

	return
}

func (p *proxyServer) sendResponse(request *http.Request, response *http.Response, outgoing http.ResponseWriter, mappings []HostMapping) {
	defer response.Body.Close()

	// Transfer headers
	outgoingHeaders := outgoing.Header()
	for key, values := range response.Header {
		for _, value := range values {
			outgoingHeaders.Add(key, value)
		}
	}

	// Content length will be recalculated as content may be rewritten
	outgoingHeaders.Del("content-length")

	// Reset transfer-encoding
	outgoingHeaders.Del("transfer-encoding")

	// We need to remove any domain information set on the cookies
	outgoingHeaders.Del("set-cookie")
	for _, cookie := range response.Cookies() {
		cookie.Domain = ""

		http.SetCookie(outgoing, cookie)
	}

	// First rewrite pass: identify matching content rewriters and process headers
	responseRewriters := make([]ResponseRewriter, 0, len(p.rewriters))

	for _, rewriter := range p.rewriters {
		if r, ok := rewriter.(ResponseRewriter); ok {
			if r.Matches(request, response) {
				responseRewriters = append(responseRewriters, r)
			}
		}

		if r, ok := rewriter.(HeaderRewriter); ok {
			r.RewriteHeaders(outgoingHeaders, mappings)
		}
	}

	// Handle compression
	bodyWriter := outgoing.(io.Writer)
	bodyReader := response.Body.(io.Reader)

	// We are ignoring any q-value here, so this is wrong for the case q=0
	clientAcceptsGzip := strings.Contains(request.Header.Get("accept-encoding"), "gzip")

	if response.Header.Get("content-encoding") == "gzip" &&
		(len(responseRewriters) > 0 || !clientAcceptsGzip) {

		var err error
		bodyReader, err = gzip.NewReader(bodyReader)
		if err != nil {
			http.Error(outgoing, err.Error(), http.StatusBadGateway)
			return
		}

		if clientAcceptsGzip {
			bodyWriter = gzip.NewWriter(bodyWriter)
			outgoingHeaders.Set("content-encoding", "gzip")
		} else {
			outgoingHeaders.Del("content-encoding")
		}
	}

	// If there are any body rewriters, we read the response into memory and process it
	if len(responseRewriters) > 0 {
		bodyData, err := ioutil.ReadAll(bodyReader)

		if err == nil {
			for _, rewriter := range responseRewriters {
				bodyData = rewriter.RewriteResponse(bodyData, mappings)
			}
		} else {
			bodyData = make([]byte, 0)
		}

		bodyReader = bytes.NewBuffer(bodyData)
	}

	// Send headers
	outgoing.WriteHeader(response.StatusCode)

	// Send body
	io.Copy(bodyWriter, bodyReader)

	// Close writer if applicable (looking at you, gzip)
	if bodyWriter, ok := bodyWriter.(io.Closer); ok {
		bodyWriter.Close()
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
			// We rather handle compression ourselves
			DisableCompression: true,
			TLSClientConfig:    tlsConfig,
		},
	}

	return
}
