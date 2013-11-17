package jest_test

import (
	"github.com/daneharrigan/jest"
	"net/http"
	"net/url"
	"testing"
)

func TestParams(t *testing.T) {
	jest.Get("/foo/:foo_id/bar/:id", serveExample)
	r := &http.Request{
		Method: "GET",
		URL: &url.URL{ Path: "/foo/1/bar/example-2" },
	}

	params := jest.Params(r)
	assertEqual(t, params["foo_id"], "1")
	assertEqual(t, params["id"], "example-2")
}

func serveExample(w http.ResponseWriter, r *http.Request) *jest.Status {
	return jest.OK
}

func assertEqual(t *testing.T, a, b string) {
	if a != b {
		t.Fatalf("Equal failure! '%s' was not '%s'", a, b)
	}
}
