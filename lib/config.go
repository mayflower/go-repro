package lib

import (
	"io"
	"os"
	"regexp"
)

type Config struct {
	mappings         []Mapping
	rewriteRoutes    []*regexp.Regexp
	log              io.Writer
	sslAllowInsecure bool
	noLogging        bool
}

func NewConfig() Config {
	return Config{
		log: os.Stdout,
	}
}

func (c *Config) AddMapping(local, remote string) (err error) {
	m, err := NewMapping(local, remote)

	if err == nil {
		c.mappings = append(c.mappings, m)
	}

	return
}

func (c *Config) AddRewriteRoute(pattern string) (err error) {
	r, err := regexp.Compile(pattern)

	if err == nil {
		c.rewriteRoutes = append(c.rewriteRoutes, r)
	}

	return
}

func (c *Config) SetLog(log io.Writer) {
	c.log = log
}

func (c *Config) CountMappings() int {
	return len(c.mappings)
}

func (c *Config) CountRewriteRoutes() int {
	return len(c.rewriteRoutes)
}

func (c *Config) SSLAllowInsecure() bool {
	return c.sslAllowInsecure
}

func (c *Config) SetSSLAllowInsecure(flag bool) {
	c.sslAllowInsecure = flag
}

func (c *Config) NoLogging() bool {
	return c.noLogging
}

func (c *Config) SetNoLogging(flag bool) {
	c.noLogging = flag
}
