package lib

import (
	"io"
)

type Repro struct {
	proxies []*proxyServer
	log     io.Writer
}

func (r *Repro) Start() (err <-chan error) {
	c := make(chan error, 1)

	for _, p := range r.proxies {
		go func(proxy *proxyServer) {
			for err := range proxy.Start() {
				c <- err
			}
		}(p)
	}

	return c
}

func NewRepro(cfg Config) (r *Repro, err error) {
	r = &Repro{
		log: cfg.log,
	}

	locationRewriter := NewLocationRewriter()
	CorsRewriter := NewCorsRewriter()
	responseRewriter := NewGenericResponseRewriter(cfg.rewriteRoutes)
	JsonRewriter := NewJsonRewriter(cfg.rewriteRoutes)

	for _, m := range cfg.mappings {
		proxyServer, e := newProxyServer(m, cfg.mappings, r.log)

		if e != nil {
			err = e
			return
		}

		proxyServer.AddRewriter(locationRewriter)
		proxyServer.AddRewriter(responseRewriter)
		proxyServer.AddRewriter(JsonRewriter)
		proxyServer.AddRewriter(CorsRewriter)

		r.proxies = append(r.proxies, proxyServer)
	}

	return
}
