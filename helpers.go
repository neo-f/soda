package soda

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

func GenStaticJSONSchema(model interface{}) *openapi3.Schema {
	rf, _ := newGenerator(&openapi3.Info{}).genSchema(nil, reflect.TypeOf(model), "json")
	return rf.Value
}
