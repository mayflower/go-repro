package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/mayflower/go-repro/lib"
)

type YamlConfig struct {
	Mappings      []YamlMapping `yaml:"mappings"`
	Rewrites      []string      `yaml:"rewrites"`
	AllowInsecure bool          `yaml:"allow-insecure"`
	NoLogging     bool          `yaml:"disable-logging"`
}

type YamlMapping struct {
	Local  string `yaml:"local"`
	Remote string `yaml:"remote"`
}

func UnmarshalYamlConfigBuffer(buffer []byte) (config YamlConfig, err error) {
	err = yaml.Unmarshal(buffer, &config)

	return
}

func UnmarshalYamlConfigFile(fname string) (config YamlConfig, err error) {
	buffer, err := ioutil.ReadFile(fname)

	if err != nil {
		return
	}

	config, err = UnmarshalYamlConfigBuffer(buffer)

	return
}

func (c *YamlConfig) createReproConfig() (cfg lib.Config, err error) {
	cfg = lib.NewConfig()

	for _, mapping := range c.Mappings {
		err = cfg.AddMapping(mapping.Local, mapping.Remote)

		if err != nil {
			return
		}
	}

	for _, rewritePattern := range c.Rewrites {
		err = cfg.AddRewriteRoute(rewritePattern)

		if err != nil {
			return
		}
	}

	cfg.SetSSLAllowInsecure(c.AllowInsecure)
	cfg.SetNoLogging(c.NoLogging)

	return
}
