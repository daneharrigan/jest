package jest

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

var (
	config      *http.ServeMux
	routes      []route
	authorize   func(http.ResponseWriter, *http.Request) *Status
	contentType string
)

func init() {
	config = http.DefaultServeMux
	config.HandleFunc("/", serveResponses)
	contentType = "application/json"
}

// public

func Handler() *http.ServeMux {
	return config
}

func Auth(fn func(http.ResponseWriter, *http.Request) *Status) {
	authorize = fn
}

func Get(uri string, fn func(http.ResponseWriter, *http.Request) *Status) *response {
	return request("GET", uri, fn)
}

func Post(uri string, fn func(http.ResponseWriter, *http.Request) *Status) *response {
	return request("POST", uri, fn)
}

// private

func request(m, u string, fn func(http.ResponseWriter, *http.Request) *Status) *response {
	rs := &response{fn: fn}
	for _, r := range routes {
		if r.URI == u {
			r.Responses[m] = rs
			return rs
		}
	}

	v := "([a-zA-Z0-9_-]+)"
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
	w := &responseWriter{rw: ow}
	w.Header().Set("Content-Type", contentType)

	header := r.Header.Get("Content-Type")
	if header != "" && header != contentType {
		write(w, BadRequest)
		return
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

			w.Header().Set("Content-Length", "0")
			w.Header().Set("Allow", strings.Join(methods, ", "))
			return
		case mOK:
			if !m.public {
				s := authorize(w, r)
				if s != nil && (s.Code < 200 || s.Code > 299) {
					write(w, s)
					return
				}
			}

			write(w, m.fn(w, r))
			return
		}
	}

	write(w, NotFound)
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
