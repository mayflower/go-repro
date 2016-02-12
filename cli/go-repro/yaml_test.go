package main

import (
	"testing"
)

func TestScalars(t *testing.T) {
	fixture1 := `
        allow-insecure: true
        disable-logging: false
    `

	fixture2 := `
        allow-insecure: false
        disable-logging: true
    `

	parsed1, err := UnmarshalYamlConfigBuffer([]byte(fixture1))

	if err != nil {
		t.Fatal(err)
	}

	parsed2, er := UnmarshalYamlConfigBuffer([]byte(fixture2))

	if err != nil {
		t.Fatal(er)
	}

	if !parsed1.AllowInsecure {
		t.Fatalf("allow-insecure failed to parse: %v", parsed1)
	}

	if parsed2.AllowInsecure {
		t.Fatalf("allow-insecure failed to parse: %v", parsed2)
	}

	if parsed1.NoLogging {
		t.Fatalf("disable-logging failed to parse: %v", parsed1)
	}

	if !parsed2.NoLogging {
		t.Fatalf("disable-logging failed to parse: %v", parsed2)
	}

	cfg1, err := parsed1.createReproConfig()

	if err != nil {
		t.Fatal(err)
	}

	cfg2, err := parsed2.createReproConfig()

	if err != nil {
		t.Fatal(err)
	}

	if !cfg1.SSLAllowInsecure() || cfg2.SSLAllowInsecure() {
		t.Fatal("SSL setup failed to propagate")
	}

	if cfg1.NoLogging() || !cfg2.NoLogging() {
		t.Fatal("Logging settings failed to propagate")
	}
}

func TestMappings(t *testing.T) {
	fixture := `
        mappings:
            - local: localhost:1234
              remote: http://foo.bar
            - local: 0.0.0.0:4321
              remote: https://bar.foo
    `

	parsed, err := UnmarshalYamlConfigBuffer([]byte(fixture))

	if err != nil {
		t.Fatal(err)
	}

	if len(parsed.Mappings) != 2 ||
		parsed.Mappings[0].Local != "localhost:1234" ||
		parsed.Mappings[0].Remote != "http://foo.bar" ||
		parsed.Mappings[1].Local != "0.0.0.0:4321" ||
		parsed.Mappings[1].Remote != "https://bar.foo" {

		t.Fatalf("mappings failed to parse: %v", parsed)
	}

	cfg, err := parsed.createReproConfig()

	if err != nil {
		t.Fatal(err)
	}

	if cfg.CountMappings() != 2 {
		t.Fatalf("mappings failed to propagate")
	}
}

func TestRewritePatterns(t *testing.T) {
	fixture := `
        rewrites:
            - foo
            - .*bar[123]+
    `

	badFixture := `
        rewrites:
            - bar
            - (
    `

	parsed, err := UnmarshalYamlConfigBuffer([]byte(fixture))

	if err != nil {
		t.Fatal(err)
	}

	parsedBad, err := UnmarshalYamlConfigBuffer([]byte(badFixture))

	if err != nil {
		t.Fatal(err)
	}

	if len(parsed.Rewrites) != 2 ||
		parsed.Rewrites[0] != "foo" ||
		parsed.Rewrites[1] != ".*bar[123]+" {

		t.Fatalf("rewrites failed to parse: %v", parsed)
	}

	cfg, err := parsed.createReproConfig()

	if err != nil {
		t.Fatal(err)
	}

	if cfg.CountRewriteRoutes() != 2 {
		t.Fatal("rewrite patterns failed to propagate")
	}

	_, err = parsedBad.createReproConfig()

	if err == nil {
		t.Fatal("bad pattern should not have been accepted")
	}
}

func TestBadYaml(t *testing.T) {
	fixture := `
        @@
    `

	_, err := UnmarshalYamlConfigBuffer([]byte(fixture))

	if err == nil {
		t.Fatal("bad YAML should not parse")
	}
}
