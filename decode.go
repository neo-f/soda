package soda

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

type parserFunc func(*http.Request, interface{}) error

var parameterParsers = map[string]parserFunc{
	"query":  parseQuery,
	"header": parseHeader,
	"path":   parsePath,
	"cookie": parseCookie,
}

var (
	queryDecoder  = newDecoder("query")
	headerDecoder = newDecoder("header")
	pathDecoder   = newDecoder("path")
	cookieDecoder = newDecoder("cookie")
)

// parseQuery parses query parameters into a struct.
func parseQuery(r *http.Request, out interface{}) error {
	return queryDecoder.Decode(out, r.URL.Query())
}

// parseHeader parses header parameters into a struct.
func parseHeader(r *http.Request, out interface{}) error {
	return headerDecoder.Decode(out, r.Header)
}

// parsePath parses path parameters into a struct.
func parsePath(r *http.Request, out interface{}) error {
	data := make(map[string][]string)
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		for i := range rctx.URLParams.Keys {
			data[rctx.URLParams.Keys[i]] = []string{rctx.URLParams.Values[i]}
		}
	}
	return pathDecoder.Decode(out, data)
}

// parseCookie parses cookie parameters into a struct.
func parseCookie(r *http.Request, out interface{}) error {
	data := make(map[string][]string)
	for _, cookie := range r.Cookies() {
		data[cookie.Name] = []string{cookie.Value}
	}
	return cookieDecoder.Decode(out, data)
}

func newDecoder(aliasTag string) *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag(aliasTag)
	decoder.IgnoreUnknownKeys(true)
	return decoder
}
