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
	"strconv"
	"strings"
)

type redirectCaughtError struct{}

type ProxyServer struct {
	local     string
	remote    string
	log       io.Writer
	rewriters []Rewriter
	mappings  []Mapping
	noLogging bool

	server http.Server
	client http.Client
}

type requestContext struct {
	incomingRequest       *http.Request
	upstreamResponse      *http.Response
	hostMappings          []HostMapping
	outgoingHeaders       http.Header
	logs                  []string
	contentLength         int
	suppressContentLength bool
	requestUrl            string
}

func (c redirectCaughtError) Error() string {
	return "redirect caught"
}

func (r *requestContext) IncomingRequest() *http.Request {
	return r.incomingRequest
}

func (r *requestContext) UpstreamResponse() *http.Response {
	return r.upstreamResponse
}

func (r *requestContext) HostMappings() []HostMapping {
	return r.hostMappings
}

func (r *requestContext) Log(message string) {
	r.logs = append(r.logs, message)
}

func (r *requestContext) RequestUrl() string {
	if r.requestUrl == "" {
		r.requestUrl = "http://" + r.incomingRequest.Host + r.incomingRequest.RequestURI
	}

	return r.requestUrl
}

func newRequestContext() *requestContext {
	return &requestContext{
		logs:          make([]string, 0, 10),
		contentLength: -1,
	}
}

func (p *ProxyServer) ServeHTTP(outgoing http.ResponseWriter, incoming *http.Request) {
	var err error

	ctx := newRequestContext()
	ctx.hostMappings = buildHostMappings(p.mappings, incoming.Host)
	ctx.incomingRequest = incoming

	upstreamRequest, err := p.buildUpstreamRequest(ctx)

	if err == nil {
		ctx.upstreamResponse, err = p.client.Do(upstreamRequest)

		if isRedirectError(err) {
			err = nil
		}
	}

	if err != nil {
		fmt.Fprintf(p.log, "error during proxy request: %v\n", err)
		http.Error(outgoing, err.Error(), http.StatusBadGateway)
	} else {
		p.sendResponse(outgoing, ctx)
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

func (p *ProxyServer) buildUpstreamRequest(ctx *requestContext) (outgoing *http.Request, err error) {
	outgoing, err = http.NewRequest(
		ctx.incomingRequest.Method,
		p.remote+ctx.incomingRequest.RequestURI,
		ctx.incomingRequest.Body)

	if err != nil {
		return
	}

	for key, values := range ctx.incomingRequest.Header {
		for _, value := range values {
			outgoing.Header.Add(key, value)
		}
	}

	for _, rewriter := range p.rewriters {
		if rewriter, ok := rewriter.(IncomingHeaderRewriter); ok {
			rewriter.RewriteIncomingHeaders(outgoing.Header, ctx)
		}
	}

	// We also try to compress upstream communication
	outgoing.Header.Set("accept-encoding", "gzip")

	return
}

func (p *ProxyServer) sendResponse(outgoing http.ResponseWriter, ctx *requestContext) {
	var err error

	defer ctx.upstreamResponse.Body.Close()

	ctx.outgoingHeaders = p.setupOutgoingHeaders(outgoing, ctx)

	bodyRewriters := p.rewriteOutgoingHeaders(ctx)
	rewriteBody := len(bodyRewriters) > 0

	bodyReader, bodyWriter, err :=
		p.handleCompression(ctx.upstreamResponse.Body, outgoing, rewriteBody, ctx)
	if err != nil {
		http.Error(outgoing, err.Error(), http.StatusBadGateway)
		return
	}

	if rewriteBody {
		bodyReader = p.rewriteBody(bodyReader, bodyRewriters, ctx)
	}

	p.setupContentLength(ctx)

	// Add the log
	if !p.noLogging {
		p.addLog(ctx)
	}

	// Send headers
	outgoing.WriteHeader(ctx.upstreamResponse.StatusCode)

	// Send body
	io.Copy(bodyWriter, bodyReader)

	// Close writer if applicable (looking at you, gzip)
	if bodyWriter, ok := bodyWriter.(io.Closer); ok {
		bodyWriter.Close()
	}
}

func (p *ProxyServer) setupOutgoingHeaders(outgoing http.ResponseWriter, ctx *requestContext) (outgoingHeaders http.Header) {
	// Transfer headers
	outgoingHeaders = outgoing.Header()
	for key, values := range ctx.upstreamResponse.Header {
		for _, value := range values {
			outgoingHeaders.Add(key, value)
		}
	}

	if contentLength, e := strconv.Atoi(ctx.upstreamResponse.Header.Get("content-length")); e == nil {
		ctx.contentLength = contentLength
	}

	// Content length will be recalculated as content may be rewritten
	outgoingHeaders.Del("content-length")

	// Reset transfer-encoding
	outgoingHeaders.Del("transfer-encoding")

	// We need to remove any domain information set on the cookies
	outgoingHeaders.Del("set-cookie")
	for _, cookie := range ctx.upstreamResponse.Cookies() {
		cookie.Domain = ""

		http.SetCookie(outgoing, cookie)
	}

	return
}

func (p *ProxyServer) rewriteOutgoingHeaders(ctx *requestContext) (bodyRewriters []BodyRewriter) {
	// First rewrite pass: identify matching content rewriters and process headers
	bodyRewriters = make([]BodyRewriter, 0, len(p.rewriters))

	for _, rewriter := range p.rewriters {
		if r, ok := rewriter.(BodyRewriter); ok {
			if r.Matches(ctx) {
				bodyRewriters = append(bodyRewriters, r)
			}
		}

		if r, ok := rewriter.(HeaderRewriter); ok {
			r.RewriteHeaders(ctx.outgoingHeaders, ctx)
		}
	}

	return
}

func (p *ProxyServer) handleCompression(readerIn io.Reader, writerIn io.Writer, rewriteResponse bool, ctx *requestContext) (
	readerOut io.Reader, writerOut io.Writer, err error) {

	readerOut = readerIn
	writerOut = writerIn

	// We are ignoring any q-value here, so this is wrong for the case q=0
	clientAcceptsGzip := strings.Contains(ctx.incomingRequest.Header.Get("accept-encoding"), "gzip")

	if ctx.upstreamResponse.Header.Get("content-encoding") == "gzip" &&
		(rewriteResponse || !clientAcceptsGzip) {

		var e error
		readerOut, e = gzip.NewReader(readerIn)
		if e != nil {
			// Work around the closed-body-on-redirect bug in the runtime
			// https://github.com/golang/go/issues/10069
			readerOut = readerIn
			return
		}

		if clientAcceptsGzip {
			writerOut = gzip.NewWriter(writerIn)
			ctx.suppressContentLength = true
			ctx.outgoingHeaders.Set("content-encoding", "gzip")
		} else {
			ctx.outgoingHeaders.Del("content-encoding")
		}
	}

	return
}

func (p *ProxyServer) rewriteBody(reader io.Reader, bodyRewriters []BodyRewriter, ctx *requestContext) io.Reader {
	bodyData, err := ioutil.ReadAll(reader)

	if err == nil {
		for _, rewriter := range bodyRewriters {
			bodyData = rewriter.RewriteResponse(bodyData, ctx)
		}
	} else {
		// Work around the closed-body-on-redirect bug in the runtime
		// https://github.com/golang/go/issues/10069
		bodyData = make([]byte, 0)
	}

	ctx.contentLength = len(bodyData)

	return bytes.NewBuffer(bodyData)
}

func (p *ProxyServer) addLog(ctx *requestContext) {
	for _, entry := range ctx.logs {
		ctx.outgoingHeaders.Add("x-go-repro-log", entry)
	}
}

func (p *ProxyServer) setupContentLength(ctx *requestContext) {
	if ctx.contentLength > 0 && !ctx.suppressContentLength {
		ctx.outgoingHeaders.Set("content-length", strconv.Itoa(ctx.contentLength))
	}
}

func (p *ProxyServer) Start() <-chan error {
	c := make(chan error, 1)

	go func() {
		c <- p.server.ListenAndServe()
	}()

	fmt.Fprintf(p.log, "proxying requests for %s to %s\n", p.local, p.remote)

	return c
}

func (p *ProxyServer) AddRewriter(r Rewriter) {
	p.rewriters = append(p.rewriters, r)
}

func (p *ProxyServer) SetNoLogging(flag bool) {
	p.noLogging = flag
}

func NewProxyServer(m Mapping, mappings []Mapping, log io.Writer, sslAllowInsecure bool) (p *ProxyServer, err error) {
	p = &ProxyServer{
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
