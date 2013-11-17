package jest

import "net/http"

type response struct {
	fn     func(http.ResponseWriter, *http.Request) *Status
	public bool
}

type responseWriter struct {
	rw            http.ResponseWriter
	written       bool
	writtenHeader bool
}

func (r *response) Public() {
	r.public = true
}

func (r *responseWriter) Header() http.Header {
	return r.rw.Header()
}

func (r *responseWriter) Write(b []byte) (int, error) {
	r.written = true
	return r.rw.Write(b)
}

func (r *responseWriter) WriteHeader(c int) {
	r.writtenHeader = true
	r.rw.WriteHeader(c)
}
