# jest todo

* Automatically assign `Retry-After` header on 503 responses
* Paginate with `Range` header (prototype)
* Should/can `HEAD` requests be automatically generated
* `BasicAuth()` helper
* `BearerToken()` helper
* Consider a testing package

#### Testing Concept

```golang
package concept_test

import "github.com/daneharrigan/jest/testing"

func TestConcept(t *testing.T) {
	t.Request.Header.Add("Authorization", "Bearer X")
	r := t.Get("/concepts")

	t.AssertEqual(200, r.StatusCode)
	t.Assert(r.Payload["Boolean"])
	t.AssertNil(r.Payload["Nil"])
}
```
