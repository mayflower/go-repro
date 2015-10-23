package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mayflower/go-repro/lib"
)

func parseCommandline() (cfg lib.Config, err error) {
	var mappingDefs, rewriteDefs string

	flag.StringVar(&mappingDefs, "m", "", "mapping definitions, format: local=remote,[local=remote,...]")
	flag.StringVar(&rewriteDefs, "r", "", "comma-separated list of regexes indetifying routes whose response will be rewritten")

	flag.Usage = func() {
		fmt.Fprint(os.Stdout, "usage: go-repro [options]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	cfg = lib.NewConfig()

	err = addMappings(mappingDefs, &cfg)

	if err == nil {
		err = addRewrites(rewriteDefs, &cfg)
	}

	return
}

func addMappings(def string, cfg *lib.Config) (err error) {
	for _, definition := range strings.Split(def, ",") {
		parts := strings.Split(definition, "=")

		if len(parts) != 2 {
			err = errors.New(fmt.Sprintf("syntax error in mapping: %s", def))
		} else {
			err = cfg.AddMapping(parts[0], parts[1])
		}

		if err != nil {
			return
		}
	}

	return
}

func addRewrites(def string, cfg *lib.Config) (err error) {
	for _, definition := range strings.Split(def, ",") {
		err = cfg.AddRewriteRoute(definition)

		if err != nil {
			return
		}
	}

	return
}

func main() {
	var err error

	cfg, err := parseCommandline()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if cfg.CountMappings() == 0 {
		fmt.Println("nothing to do")
		return
	}

	r, err := lib.NewRepro(cfg)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = <-r.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
