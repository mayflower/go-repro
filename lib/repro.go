package lib

import (
	"io"
)

type Repro struct {
	proxies []*ProxyServer
	log     io.Writer
}

func (r *Repro) Start() (err <-chan error) {
	c := make(chan error, 1)

	for _, p := range r.proxies {
		go func(proxy *ProxyServer) {
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
	refererRewriter := NewRefererRewriter()
	corsRewriter := NewCorsRewriter()
	genericResponseRewriter := NewGenericResponseRewriter(cfg.rewriteRoutes)
	jsonRewriter := NewJsonRewriter(cfg.rewriteRoutes)

	for _, m := range cfg.mappings {
		proxyServer, e := NewProxyServer(m, cfg.mappings, r.log, cfg.sslAllowInsecure)

		if e != nil {
			err = e
			return
		}

		proxyServer.AddRewriter(locationRewriter)
		proxyServer.AddRewriter(refererRewriter)
		proxyServer.AddRewriter(corsRewriter)
		proxyServer.AddRewriter(genericResponseRewriter)
		proxyServer.AddRewriter(jsonRewriter)

		proxyServer.SetNoLogging(cfg.noLogging)

		r.proxies = append(r.proxies, proxyServer)
	}

	return
}
