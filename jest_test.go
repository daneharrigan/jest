package jest_test

import (
	"net/http/httptest"
	"net/http"
	"github.com/daneharrigan/jest"
	"testing"
	"bufio"
	"strings"
	"encoding/json"
)

func TestUnauthorizedValidMethod(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := req("GET", s.URL+"/")
	assertErrNil(t, err)
	assertStatus(t, r, 403)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Code":403,"Message":"Forbidden"}`)
}

func TestUnauthorizedInvalidMethod(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := req("POST", s.URL+"/")
	assertErrNil(t, err)
	assertStatus(t, r, 405)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Code":405,"Message":"Method Not Allowed"}`)
}

func TestUnauthorizedOptions(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := req("OPTIONS", s.URL+"/")
	assertErrNil(t, err)
	assertStatus(t, r, 200)
	assertHeader(t, r, "Content-Type", "application/json")
	assertHeader(t, r, "Content-Length", "0")
	assertHeader(t, r, "Allow", "GET, OPTIONS")
	assertEmptyBody(t, r)
}

func TestAuthroziedValidMethod(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := auth("GET", s.URL+"/")
	assertErrNil(t, err)
	assertStatus(t, r, 200)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Code":200,"Message":"OK"}`)
}

func TestAuthorizedInvalidMethod(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := auth("POST", s.URL+"/")
	assertErrNil(t, err)
	assertStatus(t, r, 405)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Code":405,"Message":"Method Not Allowed"}`)
}

func TestAuthorizedOptions(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := auth("OPTIONS", s.URL+"/")
	assertErrNil(t, err)
	assertStatus(t, r, 200)
	assertHeader(t, r, "Content-Type", "application/json")
	assertHeader(t, r, "Content-Length", "0")
	assertHeader(t, r, "Allow", "GET, OPTIONS")
	assertEmptyBody(t, r)
}

func TestAuthorizedCustom(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := auth("GET", s.URL+"/custom")
	assertErrNil(t, err)
	assertStatus(t, r, 200)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Foo":"foo","Bar":1}`)
}

func TestPublic(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := req("GET", s.URL+"/public")
	assertErrNil(t, err)
	assertStatus(t, r, 200)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Foo":"foo","Bar":1}`)
}

func TestNotFound(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := req("GET", s.URL+"/404")
	assertErrNil(t, err)
	assertStatus(t, r, 404)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Code":404,"Message":"Not Found"}`)
}

func TestBadRequest(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := auth("GET", s.URL+"/", "application/xml", "<xml></xml>")
	assertErrNil(t, err)
	assertStatus(t, r, 400)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Code":400,"Message":"Bad Request"}`)
}

func TestContentTypeNoPayload(t *testing.T) {
	s := server()
	defer s.Close()

	r, err := auth("GET", s.URL+"/", "application/xml")
	assertErrNil(t, err)
	assertStatus(t, r, 400)
	assertHeader(t, r, "Content-Type", "application/json")
	assertEqualBody(t, r, `{"Code":400,"Message":"Bad Request"}`)
}

func server() *httptest.Server {
	jest.Auth(handleAuthorization)
	jest.Get("/", serveRoot)
	jest.Get("/custom", serveCustom)
	jest.Get("/public", serveCustom).Public()
	return httptest.NewServer(jest.Handler())
}

func serveRoot(w http.ResponseWriter, r *http.Request) *jest.Status {
	return nil
}

func serveCustom(w http.ResponseWriter, r *http.Request) *jest.Status {
	var s struct {
		Foo string
		Bar int
	}

	s.Foo = "foo"
	s.Bar = 1

	json.NewEncoder(w).Encode(s)
	return jest.OK
}

func handleAuthorization(w http.ResponseWriter, r *http.Request) *jest.Status {
	header := r.Header.Get("Authorization")
	if header != "Bearer X" {
		return jest.Forbidden
	}

	return nil
}

// helpers

func req(m, u string, args ...string) (*http.Response, error) {
	var r *http.Request
	var err error

	if len(args) > 1 {
		r, err = http.NewRequest(m, u, strings.NewReader(args[1]))
	} else {
		r, err = http.NewRequest(m, u, nil)
	}

	if err != nil {
		return nil, err
	}

	if len(args) > 0 {
		r.Header.Add("Content-Type", args[0])
	}

	return http.DefaultClient.Do(r)
}

func auth(m, u string, args ...string) (*http.Response, error) {
	var r *http.Request
	var err error

	if len(args) > 1 {
		r, err = http.NewRequest(m, u, strings.NewReader(args[1]))
	} else {
		r, err = http.NewRequest(m, u, nil)
	}

	if err != nil {
		return nil, err
	}

	if len(args) > 0 {
		r.Header.Add("Content-Type", args[0])
	}

	r.Header.Add("Authorization", "Bearer X")
	return http.DefaultClient.Do(r)
}

func assertErrNil(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Error failure! %s", err)
	}
}

func assertStatus(t *testing.T, r *http.Response, exp int) {
	act := r.StatusCode
	if act != exp {
		t.Fatalf("Status failure! Expected %d, but got %d", exp, act)
	}
}

func assertEqualBody(t *testing.T, r *http.Response, s string) {
	b, _ := bufio.NewReader(r.Body).ReadString('$')
	if strings.TrimSpace(b) != s {
		t.Fatalf("Body failure! '%s' was not the body, but '%s' was", s, b)
	}
}

func assertEmptyBody(t *testing.T, r *http.Response) {
	b, _ := bufio.NewReader(r.Body).ReadString('$')
	if len(b) > 0 {
		t.Fatalf("Body failure! Body is not empty")
	}
}

func assertHeader(t *testing.T, r *http.Response, n, v string) {
	h := r.Header.Get(n)
	if h != v {
		t.Fatalf("Header failure! Expected %s to be %s, but got %s", n, v, h)
	}
}
