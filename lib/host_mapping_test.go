package lib

import (
	"testing"
)

func TestLocalip(t *testing.T) {
	mappings := []Mapping{
		{
			local:  "127.0.0.1:8080",
			remote: "foo.bar.com",
		},
	}

	hostMappings := buildHostMappings(mappings, "192.168.0.1:8090")

	if hostMappings[0].local != "http://127.0.0.1:8080" {
		t.Fatalf("expected http://127.0.0.1:8080, got %s", hostMappings[0].local)
	}
}

func TestLocalhost(t *testing.T) {
	mappings := []Mapping{
		{
			local:  "localhost:8080",
			remote: "foo.bar.com",
		},
	}

	hostMappings := buildHostMappings(mappings, "192.168.0.1:8090")

	if hostMappings[0].local != "http://localhost:8080" {
		t.Fatalf("expected http://localhost:8080, got %s", hostMappings[0].local)
	}
}

func TestAllInterfaces(t *testing.T) {
	mappings := []Mapping{
		{
			local:  "0.0.0.0:8080",
			remote: "foo.bar.com",
		},
	}

	hostMappings := buildHostMappings(mappings, "192.168.0.1:8090")

	if hostMappings[0].local != "http://192.168.0.1:8080" {
		t.Fatalf("expected http://192.168.0.1:8080, got %s", hostMappings[0].local)
	}
}
