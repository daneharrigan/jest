# jest

A Jest is a JSON in/out REST API written in Go. Jest only accepts
`application/json` content types and responds with JSON only.

* All routes in Jest require authorization unless marked as public
* Responses are always JSON
* Request payloads are always JSON
* Any request with a Content-Type not `application/json` will be rejected
* All requests that do not resolve to a route will receive a 404 response
* Requests for OPTIONS are automatcally generated based on existing routes
* If a status code is not specified 200 is assumed

### Example

A simple `GET` and `POST` example with authentication:

```go
package main

import (
	"encoding/json"
	"github.com/daneharrigan/jest"
	"net/http"
)

func main() {
	jest.Auth(handleAuthorization)
	jest.Get("/examples", serveGetExamples)
	jest.Post("/examples", servePostExamples)
	http.ListenAndServe(":5000", jest.Handler())
}

func handleAuthorization(w http.ResponseWriter, r *http.Request) *jest.Status {
	if r.Header.Get("Authorization") != "Bearer X" {
		return jest.Forbidden
	}

	return jest.OK
}

func serveGetExamples(w http.ResponseWriter, r *http.Request) *jest.Status {
	e := getExamples() // get many example resources
	json.NewEncoder(w).Encode(e)
	return jest.OK
}

func servePostExamples(w http.ResponseWriter, r *http.Request) *jest.Status {
	e := new(Example)
	json.NewDecoder(r.Body).Decode(e)
	saveExample(e) // save the new example resource

	json.NewEncoder(w).Encode(e)
	return jest.Created
}
```

An example reading parameters from the URL:

```go
package main

import (
	"encoding/json"
	"github.com/daneharrigan/jest"
	"net/http"
)

func main() {
	jest.Get("/examples/:id", serveGetExample).Public()
	http.ListenAndServe(":5000", jest.Handler())
}

func serveGetExample(w http.ResponseWriter, r *http.Request) *jest.Status {
	params := jest.Params(r)
	e := getExample(params["id"])
	json.NewEncoder(w).Encode(e)
	return jest.OK
}
```

Using jest and `net/http` together:

```go
package main

import (
	"github.com/daneharrigan/jest"
	"net/http"
)

func main() {
	jest.Auth(handleAuthorization)
	jest.Get("/examples", serveGetExamples)

	http.HandleFunc("/login", serveLogin)
	http.ListenAndServe(":5000", jest.Handler())
}

func handleAuthorization(w http.ResponseWriter, r *http.Request) *jest.Status {
	if r.Header.Get("Authorization") != "Bearer X" {
		return jest.Forbidden
	}

	return jest.OK
}

func serveGetExamples(w http.ResponseWriter, r *http.Request) *jest.Status {
	e := getExamples() // get many example resources
	json.NewEncoder(w).Encode(e)
	return jest.OK
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "login.html")
}
```
