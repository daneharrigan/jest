# jest

A Jest is a JSON in/out REST API written in Go. Jest only accepts
`application/json` content types and responds with JSON only.

All routes in Jest are assumed to be private


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
