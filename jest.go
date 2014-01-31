// Package jest provides a minimal wrapper to net/http for building JSON in/out
// API's. Jest assumes all routes require authorization unless stated otherwise.
// Conveniences such as auto-generating the `OPTIONS` resposne and enforcing the
// use of SSL are offered via jest.Handler(). The Jest package plays nicely with
// `net/http` so routes can be delcared with both and served as a single app.
//
// URI's passed to the `Get`, `Post`, `Patch`, `Put` and `Delete` methods
// support variables:
//
//  package main
//
//  import (
//  	"github.com/daneharrigan/jest"
//	"net/http"
//  )
//
//  func main() {
//  	jest.Get("/resource/:id", handleGetResource)
//	http.ListenAndServe(":5000", jest.Handler())
//  }
//
//  func handleGetResource(w http.ResponseWriter, r *http.Request) *jest.Status {
//  	params := jest.Params(r)
//	// => map[string]string
//	// params["id"] = "<value-of-id>"
//  }
//
// JSON API's that offer OAuth logins can be handled with a mix of Jest routes
// and standard `net/http` routes. Below is an example:
//
//  package main
//
//  import (
//  	"net/http"
//	"github.com/daneharrigan/jest"
//	"encoding/json"
//  )
//
//  func main() {
//  	jest.Auth(handleAuth)
//  	jest.Get("/resources", handleGetResources) // jest route
//	http.HandleFunc("/oauth/authorize", handleOAuthAuthorize) // standard route
//  	http.ListenAndServe(":5000", jest.Handler())
//  }
//
//  func handleAuth(w http.ResponseWriter, r *http.Request) *jest.Status {
//	// determine authorization
//	if authorized(r) {
//		return jest.OK
//	}
//	return jest.Forbidden
//  }
//
//  func handleGetResources(w http.ResponseWriter, r *http.Request) *jest.Status {
//	json.NewEncoder(w).Encode(getResources())
//	return jest.OK
//  }
//
//  func handleOAuthAuthorize(w http.ResponseWriter, r *http.Request) {
//	// render html view
//  }
//
package jest

import (
	"encoding/json"
	"github.com/kr/secureheader"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	config      *secureheader.Config
	routes      []route
	authorize   func(http.ResponseWriter, *http.Request) *Status
	contentType string
)

func init() {
	contentType = "application/json"
	config = secureheader.DefaultConfig
	http.HandleFunc("/", serveResponses)
}

/* public */

// Handler returns an http handler for http.ListenAndServe. The handler is
// provided by github.com/kr/secureheader and decorates HTTP response with
// a series of secure header information.
func Handler() *secureheader.Config {
	https := os.Getenv("HTTPS")
	if https == "" {
		https = os.Getenv("JEST_HTTPS")
	}

	config.HTTPSRedirect = https != "false"
	return config
}

// Auth accepts a method for granting/denying access to protected routes.
// Returning a jest.Status struct is used to determine whether authorization
// should be allowed or denied.
func Auth(fn func(http.ResponseWriter, *http.Request) *Status) {
	authorize = fn
}

// Get accepts a URI string and a method for handling each request. The method
// mimics the net/http package by accepting http.ResponseWriter and
// *http.Request. The method differs though by expecting a *jest.Status to be
// returned.
func Get(uri string, f func(http.ResponseWriter, *http.Request) *Status) *response {
	return request("GET", uri, f)
}

// Post accepts a URI string and a method for handling each request. The method
// mimics the net/http package by accepting http.ResponseWriter and
// *http.Request. The method differs though by expecting a *jest.Status to be
// returned.
func Post(uri string, f func(http.ResponseWriter, *http.Request) *Status) *response {
	return request("POST", uri, f)
}

// Put accepts a URI string and a method for handling each request. The method
// mimics the net/http package by accepting http.ResponseWriter and
// *http.Request. The method differs though by expecting a *jest.Status to be
// returned.
func Put(uri string, f func(http.ResponseWriter, *http.Request) *Status) *response {
	return request("PUT", uri, f)
}

// Patch accepts a URI string and a method for handling each request. The method
// mimics the net/http package by accepting http.ResponseWriter and
// *http.Request. The method differs though by expecting a *jest.Status to be
// returned.
func Patch(uri string, f func(http.ResponseWriter, *http.Request) *Status) *response {
	return request("PATCH", uri, f)
}

// Delete accepts a URI string and a method for handling each request. The method
// mimics the net/http package by accepting http.ResponseWriter and
// *http.Request. The method differs though by expecting a *jest.Status to be
// returned.
func Delete(uri string, f func(http.ResponseWriter, *http.Request) *Status) *response {
	return request("DELETE", uri, f)
}

/* private */

func request(m, u string, fn func(http.ResponseWriter, *http.Request) *Status) *response {
	rs := &response{fn: fn}
	for _, r := range routes {
		if r.URI == u {
			r.Responses[m] = rs
			return rs
		}
	}

	v := `([^/]*)`
	rx := regexp.MustCompile(":" + v)
	r := route{
		URI:       u,
		Responses: make(map[string]*response),
	}

	mr := regexp.MustCompile("^" + rx.ReplaceAllString(u, v) + "$")
	pr := regexp.MustCompile("^" + rx.ReplaceAllString(u, ":"+v) + "$")

	r.Responses[m] = rs
	r.URIMatcher = mr
	r.ParamMatcher = pr

	routes = append(routes, r)
	return rs
}

func serveResponses(ow http.ResponseWriter, r *http.Request) {
	headers := "Authorization, Accept, Range, Content-Type, Host, Origin"
	w := &responseWriter{rw: ow}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Headers", headers)

	defer r.Body.Close()

	if header := r.Header.Get("Content-Type"); header != "" {
		size := len(contentType)
		if len(header) < size || header[:size] != contentType {
			badRequest := &Status{Code: 400, Message: "Bad Request"}
			badRequest.Errors = append(badRequest.Errors, "Invalid Content-Type")
			write(w, badRequest)
			return
		}
	}

	for _, route := range routes {
		if !route.URIMatcher.MatchString(r.URL.Path) {
			continue
		}

		m, mOK := route.Responses[r.Method]
		switch {
		default:
			write(w, MethodNotAllowed)
			return
		case r.Method == "OPTIONS":
			var methods []string
			for m, _ := range route.Responses {
				methods = append(methods, m)
			}

			methods = append(methods, "OPTIONS")
			allowed := strings.Join(methods, ", ")

			w.Header().Set("Content-Length", "0")
			w.Header().Set("Allow", allowed)
			w.Header().Set("Access-Control-Allow-Methods", allowed)
			return
		case mOK:
			if !m.public {
				if authorize == nil {
					write(w, Forbidden)
					return
				}

				s := authorize(w, r)
				if s != nil && (s.Code < 200 || s.Code > 299) {
					write(w, s)
					return
				}
			}

			defer internalError(w)
			write(w, m.fn(w, r))
			return
		}
	}

	write(w, NotFound)
}

func internalError(w *responseWriter) {
	if recover() != nil {
		write(w, InternalServerError)
	}
}

func write(w *responseWriter, s *Status) {
	if s == nil {
		s = OK
	}

	if !w.written {
		if !w.writtenHeader {
			w.WriteHeader(s.Code)
		}

		if s != NoContent {
			json.NewEncoder(w).Encode(s)
		}
	}
}
