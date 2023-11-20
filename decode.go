package soda

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
)

// parserFunc is a function type that takes an http.Request and an any to parse the request into.
type parserFunc func(*http.Request, any) error

// parameterParsers is a map of parser functions for different parameter types.
var parameterParsers = map[string]parserFunc{
	"query":  parseQuery,  // Function to parse query parameters
	"header": parseHeader, // Function to parse header parameters
	"path":   parsePath,   // Function to parse path parameters
	"cookie": parseCookie, // Function to parse cookie parameters
}

// Predefined decoders for different parameter types.
var (
	queryDecoder  = newDecoder("query")
	headerDecoder = newDecoder("header")
	pathDecoder   = newDecoder("path")
	cookieDecoder = newDecoder("cookie")
)

// parseQuery parses query parameters from the request into the provided interface.
func parseQuery(r *http.Request, out any) error {
	return queryDecoder.Decode(out, r.URL.Query())
}

// parseHeader parses header parameters from the request into the provided interface.
func parseHeader(r *http.Request, out any) error {
	return headerDecoder.Decode(out, r.Header)
}

// parsePath parses path parameters from the request into the provided interface.
func parsePath(r *http.Request, out any) error {
	data := make(map[string][]string)
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		for i := range rctx.URLParams.Keys {
			data[rctx.URLParams.Keys[i]] = []string{rctx.URLParams.Values[i]}
		}
	}
	return pathDecoder.Decode(out, data)
}

// parseCookie parses cookie parameters from the request into the provided interface.
func parseCookie(r *http.Request, out any) error {
	data := make(map[string][]string)
	for _, cookie := range r.Cookies() {
		data[cookie.Name] = []string{cookie.Value}
	}
	return cookieDecoder.Decode(out, data)
}

// newDecoder creates a new gorilla/schema decoder with the provided alias tag.
// It ignores unknown keys.
func newDecoder(aliasTag string) *schema.Decoder {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag(aliasTag)
	decoder.IgnoreUnknownKeys(true)
	return decoder
}
