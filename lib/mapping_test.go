package lib

import (
	"testing"
)

func TestInvalidRemote(t *testing.T) {
	_, err := newMapping("0.0.0.0:8080", "foo.bar.com")

	if err == nil {
		t.Fatal("missing scheme should be an error")
	}
}

func TestInvalidScheme(t *testing.T) {
	_, err := newMapping("0.0.0.0:8080", "foo://bar.com")

	if err == nil {
		t.Fatal("invalid scheme should be an error")
	}
}

func TestTrailingSlash(t *testing.T) {
	m, err := newMapping("0.0.0.0:8080", "http://foo.bar.com/")

	if err != nil {
		t.Fatalf("instantiation failed: %v", err)
	}

	if m.remote != "http://foo.bar.com" {
		t.Fatal("trailing slash should be amputated")
	}
}

func TestNontrivialPath(t *testing.T) {
	_, err := newMapping("0.0.0.0:8080", "http://foo.bar.com/baz")

	if err == nil {
		t.Fatal("nontrivial path should be an error")
	}
}
