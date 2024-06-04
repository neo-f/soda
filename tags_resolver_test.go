package soda_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/neo-f/soda/v3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTagResolver(t *testing.T) {
	Convey("Given a struct field with a '-' tag", t, func() {
		type testStruct struct {
			A string `json:"a" oai:"-"`
		}

		Convey("It should inject nothing into the schema", func() {
			schema := soda.GenerateSchemaRef(testStruct{}, "json")
			So(schema.Value, ShouldResemble, openapi3.NewObjectSchema())
		})
	})

	Convey("Given a struct field with oai tag", t, func() {
		Convey("It should inject generic tags", func() {
			type testStruct struct {
				A string `json:"a" oai:"type=string;title=Test;description=Test;deprecated;readOnly"`
				B string `json:"b" oai:"writeOnly;nullable"`
			}
			schema := soda.GenerateSchemaRef(testStruct{}, "json")
			expectA := openapi3.NewStringSchema()
			expectA.Title = "Test"
			expectA.Description = "Test"
			expectA.Deprecated = true
			expectA.ReadOnly = true
			So(schema.Value.Properties["a"].Value, ShouldResemble, expectA)

			expectB := openapi3.NewStringSchema()
			expectB.WriteOnly = true
			expectB.Nullable = true
			So(schema.Value.Properties["b"].Value, ShouldResemble, expectB)
		})

		Convey("It should inject string related tags", func() {
			type testStruct struct {
				A string `json:"a" oai:"minLength=1;maxLength=8;pattern=^\\d{1}$;format=number;enum=1,2,3;default=1;example=1;required"`
			}
			schema := soda.GenerateSchemaRef(testStruct{}, "json")

			expectA := openapi3.NewStringSchema().
				WithMinLength(1).
				WithMaxLength(8).
				WithPattern("^\\d{1}$").
				WithFormat("number").
				WithEnum("1", "2", "3").
				WithDefault("1")
			expectA.Example = "1"
			expect := openapi3.NewObjectSchema().
				WithProperty("a", expectA).
				WithRequired([]string{"a"})
			So(schema.Value, ShouldResemble, expect)
		})

		Convey("It should inject number related tags", func() {
			type testStruct struct {
				A int     `json:"a" oai:"multipleOf=1;minimum=1;exclusiveMinimum;enum=1,2,3"`
				B int     `json:"b" oai:"multipleOf=1;maximum=9;exclusiveMaximum;enum=1,2,3"`
				C int     `json:"c" oai:"default=1;example=99"`
				D float64 `json:"d" oai:"default=1.1;example=99"`
				E float64 `json:"e" oai:"enum=1.1,1.2,1.3"`
			}
			schema := soda.GenerateSchemaRef(testStruct{}, "json")

			expectA := openapi3.NewIntegerSchema()
			expectA.MultipleOf = openapi3.Float64Ptr(1)
			expectA.Min = openapi3.Float64Ptr(1)
			expectA.ExclusiveMin = true
			expectA.Enum = []any{1, 2, 3}
			So(schema.Value.Properties["a"].Value, ShouldResemble, expectA)

			expectB := openapi3.NewIntegerSchema()
			expectB.MultipleOf = openapi3.Float64Ptr(1)
			expectB.Max = openapi3.Float64Ptr(9)
			expectB.ExclusiveMax = true
			expectB.Enum = []any{1, 2, 3}
			So(schema.Value.Properties["b"].Value, ShouldResemble, expectB)

			expectC := openapi3.NewIntegerSchema().WithDefault(1)
			expectC.Example = 99
			So(schema.Value.Properties["c"].Value, ShouldResemble, expectC)

			expectD := openapi3.NewFloat64Schema().WithDefault(1.1)
			expectD.Example = 99.0
			So(schema.Value.Properties["d"].Value, ShouldResemble, expectD)

			expectE := openapi3.NewFloat64Schema().WithEnum(1.1, 1.2, 1.3)
			So(schema.Value.Properties["e"].Value, ShouldResemble, expectE)
		})
	})

	Convey("Given a struct field with array related tags", t, func() {
		type testStruct struct {
			A []float64 `json:"a" oai:"minItems=1;maxItems=8;uniqueItems"`
		}

		Convey("It should correctly inject array related tags into the schema", func() {
			schema := soda.GenerateSchemaRef(testStruct{}, "json")
			expectA := openapi3.NewArraySchema().WithItems(openapi3.NewFloat64Schema())
			expectA.MinItems = 1
			expectA.MaxItems = openapi3.Uint64Ptr(8)
			expectA.UniqueItems = true
			expect := openapi3.
				NewObjectSchema().
				WithProperty("a", expectA).
				WithRequired([]string{"a"})

			So(schema.Value, ShouldResemble, expect)
		})
	})

	Convey("Given a struct field with boolean related tags", t, func() {
		type testStruct struct {
			A bool `json:"a" oai:"default=true;example=false"`
		}

		Convey("It should correctly inject boolean related tags into the schema", func() {
			schema := soda.GenerateSchemaRef(testStruct{}, "json")
			expectA := openapi3.NewBoolSchema().WithDefault(true)
			expectA.Example = false

			expect := openapi3.NewObjectSchema().
				WithProperty("a", expectA).
				WithRequired([]string{"a"})

			So(schema.Value, ShouldResemble, expect)
		})
	})
}
