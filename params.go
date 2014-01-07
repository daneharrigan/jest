package jest

import "net/http"

// Params accepts the *http.Request and parses parameters found in the URL. A
// map is returned where the keys are parameter names and the values are the
// values found in the URL.
func Params(r *http.Request) map[string]string {
	params := make(map[string]string)

	for _, route := range routes {
		if !route.URIMatcher.MatchString(r.URL.Path) {
			continue
		}

		keys := route.ParamMatcher.FindAllStringSubmatch(route.URI, -1)
		values := route.URIMatcher.FindAllStringSubmatch(r.URL.Path, -1)

		for i := 1; i < len(keys[0]); i++ {
			k := keys[0][i]
			v := values[0][i]
			params[k] = v
		}
	}

	return params
}
