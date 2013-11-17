package jest

import "regexp"

type route struct {
	URI          string
	URIMatcher   *regexp.Regexp
	ParamMatcher *regexp.Regexp
	Responses    map[string]*response
}
