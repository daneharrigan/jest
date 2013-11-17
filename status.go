package jest

import "net/http"

var (
	OK		    = NewStatus(200, http.StatusText(200))
	Created             = NewStatus(201, http.StatusText(201))
	NoContent           = NewStatus(204, http.StatusText(204))
	BadRequest	    = NewStatus(400, http.StatusText(400))
	NotFound            = NewStatus(404, http.StatusText(404))
	Forbidden           = NewStatus(403, http.StatusText(403))
	MethodNotAllowed    = NewStatus(405, http.StatusText(405))
	InternalServerError = NewStatus(500, http.StatusText(500))
)

type Status struct {
	Code    int
	Message string
}

func NewStatus(c int, s string) *Status {
	return &Status{Code: c, Message: s}
}
