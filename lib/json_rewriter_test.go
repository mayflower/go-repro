package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"testing"
)

type MockContext []HostMapping

func (m MockContext) IncomingRequest() *http.Request {
	return nil
}

func (m MockContext) UpstreamResponse() *http.Response {
	return nil
}

func (m MockContext) RequestUrl() string {
	return ""
}

func (m MockContext) Log(message string) {}

func (m MockContext) HostMappings() []HostMapping {
	return m
}

func newMockContext() (ctx MockContext, err error) {
	mappings := make([]Mapping, 0, 2)

	var mapping Mapping

	mapping, err = NewMapping("1.2.3.4:8888", "http://foo.bar")

	if err != nil {
		return
	}

	mappings = append(mappings, mapping)

	mapping, err = NewMapping("4.3.2.1:9999", "https://bar.baz")

	if err != nil {
		return
	}

	mappings = append(mappings, mapping)

	ctx = buildHostMappings(mappings, "localhost")

	return
}

func assertRewritesTo(originalJson string, rewrittenJson string) (err error) {
	rewriter := NewJsonRewriter(make([]*regexp.Regexp, 0, 0))

	var ctx MockContext
	ctx, err = newMockContext()

	if err != nil {
		return
	}

	rewritten := rewriter.RewriteResponse([]byte(originalJson), ctx)

	var parsed, parsedFixture interface{}

	err = json.Unmarshal(rewritten, &parsed)

	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(rewrittenJson), &parsedFixture)

	if err != nil {
		return
	}

	if !reflect.DeepEqual(parsed, parsedFixture) {
		err = errors.New(fmt.Sprintf("rewrite failed, got %s, expected %s", string(rewritten), rewrittenJson))
	}

	return
}

func TestRewriteSimple(t *testing.T) {
	jsonData := `
        {
            "foo": 1,
            "hanni": true,
            "bar": "http://bar.foo",
            "baz": "http://foo.bar"
        }
    `

	fixture := `
        {
            "foo": 1,
            "hanni": true,
            "bar": "http://bar.foo",
            "baz": "http://1.2.3.4:8888"
        }
    `

	if err := assertRewritesTo(jsonData, fixture); err != nil {
		t.Fatal(err)
	}
}

func TestRewriteString(t *testing.T) {
	jsonData := `"http://foo.bar"`

	fixture := `"http://1.2.3.4:8888"`

	if err := assertRewritesTo(jsonData, fixture); err != nil {
		t.Fatal(err)
	}
}

func TestRewriteArray(t *testing.T) {
	jsonData := `
        [
            "http://foo.bar",
            "https://bar.baz",
            "http://bar.foo",
            true,
            42
        ]
    `

	fixture := `
        [
            "http://1.2.3.4:8888",
            "http://4.3.2.1:9999",
            "http://bar.foo",
            true,
            42
        ]
    `

	if err := assertRewritesTo(jsonData, fixture); err != nil {
		t.Fatal(err)
	}
}

func TestNested(t *testing.T) {
	jsonData := `{
        "foo": {
            "bar": "http://foo.bar",
            "baz": [
                1,
                "https://bar.baz",
                {
                    "hanni": "http://foo.bar"
                }
            ]
        }
    }`

	fixture := `{
            "foo": {
                "bar": "http://1.2.3.4:8888",
                "baz": [
                    1,
                    "http://4.3.2.1:9999",
                    {
                        "hanni": "http://1.2.3.4:8888"
                    }
                ]
            }
        }`

	if err := assertRewritesTo(jsonData, fixture); err != nil {
		t.Fatal(err)
	}
}

func testRewriteKeys(t *testing.T) {
	jsonData := `{
        "http://foo.bar": 42,
        "foo": {
            "https://bar.baz": 24,
            "http://4.3.2.1:9999": "huppe",
            "bar": {
                "https://bar.baz": "huppe",
            }
        }
    }`

	fixture := `{
        "http://1.2.3.4:8888": 42,
        "foo": {
            "https://bar.baz": 24,
            "http://4.3.2.1:9999": "huppe",
            "bar": {
                "http://4.3.2.1:9999": "huppe",
            }
        }
    }`

	if err := assertRewritesTo(jsonData, fixture); err != nil {
		t.Fatal(err)
	}
}
