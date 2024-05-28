package soda_test

import (
	"encoding/json"
	"math"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/neo-f/soda/v3"
	. "github.com/smartystreets/goconvey/convey"
)

type case4 struct {
	X string `json:"x"`
}

func (c case4) JSONSchema(t *openapi3.T) *openapi3.SchemaRef {
	return openapi3.NewObjectSchema().
		WithProperty("x", openapi3.NewStringSchema().WithEnum("a", "b")).
		NewRef()
}

func TestGenerator(t *testing.T) {
	Convey("Given a soda generator", t, func() {
		g := soda.NewGenerator()

		Convey("When the generator is created", func() {
			So(g, ShouldNotBeNil)
		})

		Convey("When GenerateSchemaRef is called", func() {
			Convey("It should return the correct schema for string", func() {
				schema := soda.GenerateSchemaRef("", "")
				So(schema, ShouldResemble, openapi3.NewStringSchema().NewRef())
			})

			Convey("It should return the correct schema for integers", func() {
				type testCase struct {
					Name     string
					Actual   any
					Expected *openapi3.SchemaRef
				}
				cases := []testCase{
					{"int", int(0), openapi3.NewIntegerSchema().NewRef()},
					{"uint", uint(0), openapi3.NewIntegerSchema().WithMin(0).NewRef()},
					{"int8", int8(0), openapi3.NewIntegerSchema().WithMin(math.MinInt8).WithMax(math.MaxInt8).NewRef()},
					{"uint8", uint8(0), openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint8).NewRef()},
					{"int16", int16(0), openapi3.NewIntegerSchema().WithMin(math.MinInt16).WithMax(math.MaxInt16).NewRef()},
					{"uint16", uint16(0), openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint16).NewRef()},
					{"int32", int32(0), openapi3.NewInt32Schema().WithMin(math.MinInt32).WithMax(math.MaxInt32).NewRef()},
					{"uint32", uint32(0), openapi3.NewInt32Schema().WithMin(0).WithMax(math.MaxUint32).NewRef()},
					{"int64", int64(0), openapi3.NewInt64Schema().WithMin(math.MinInt64).WithMax(math.MaxInt64).NewRef()},
					{"uint64", uint64(0), openapi3.NewInt64Schema().WithMin(0).WithMax(math.MaxUint64).NewRef()},
				}
				for _, c := range cases {
					Convey("It should return the correct schema for "+c.Name, func() {
						So(soda.GenerateSchemaRef(c.Actual, ""), ShouldResemble, c.Expected)
					})
				}
			})

			Convey("It should return the correct schema for float", func() {
				So(soda.GenerateSchemaRef(float32(0), ""), ShouldResemble, openapi3.NewFloat64Schema().NewRef())
				So(soda.GenerateSchemaRef(float64(0), ""), ShouldResemble, openapi3.NewFloat64Schema().NewRef())
			})

			Convey("It should return the correct schema for boolean", func() {
				schema := soda.GenerateSchemaRef(true, "")
				So(schema, ShouldResemble, openapi3.NewBoolSchema().NewRef())
			})

			Convey("It should return the correct schema for map[string]any", func() {
				schema := soda.GenerateSchemaRef(map[string]any{}, "")
				So(schema, ShouldResemble, openapi3.NewObjectSchema().WithAnyAdditionalProperties().NewRef())
			})

			Convey("It should return the correct schema for time.Time", func() {
				schema := soda.GenerateSchemaRef(time.Time{}, "")
				So(schema, ShouldResemble, openapi3.NewStringSchema().WithFormat("date-time").NewRef())
			})

			Convey("It should return the correct schema for net.IP", func() {
				schema := soda.GenerateSchemaRef(net.IP{}, "")
				So(schema, ShouldResemble, openapi3.NewStringSchema().WithFormat("ipv4").NewRef())
			})

			Convey("It should return the correct schema for json.RawMessage", func() {
				schema := soda.GenerateSchemaRef(json.RawMessage{}, "")
				So(schema, ShouldResemble, openapi3.NewStringSchema().WithFormat("json").NewRef())
			})

			Convey("It should return the correct schema for []byte", func() {
				schema := soda.GenerateSchemaRef([]byte{}, "")
				So(schema, ShouldResemble, openapi3.NewBytesSchema().NewRef())
			})

			Convey("It should return the correct schema for array", func() {
				schema := soda.GenerateSchemaRef([2]int{}, "")
				expected := openapi3.NewArraySchema().
					WithMaxItems(2).
					WithMinItems(2).
					WithItems(openapi3.NewIntegerSchema()).NewRef()
				So(schema, ShouldResemble, expected)
			})

			Convey("It should return the correct schema for slice", func() {
				schema := soda.GenerateSchemaRef([]int{}, "")
				So(schema, ShouldResemble, openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).NewRef())
			})

			Convey("It should return the correct schema for map[string]int", func() {
				schema := soda.GenerateSchemaRef(map[string]int{}, "")
				So(schema, ShouldResemble, openapi3.NewObjectSchema().WithAdditionalProperties(openapi3.NewIntegerSchema()).NewRef())
			})

			Convey("It should return the correct schema for a basic struct", func() {
				type TestCase struct {
					A string
					B int
				}
				schema := soda.GenerateSchemaRef(TestCase{}, "")
				expected := openapi3.NewObjectSchema().
					WithProperty("A", openapi3.NewStringSchema()).
					WithProperty("B", openapi3.NewIntegerSchema()).
					WithRequired([]string{"A", "B"})
				So(schema.Value, ShouldResemble, expected)
				So(schema.Ref, ShouldEqual, "#/components/schemas/soda_test.TestCase")
			})

			Convey("It should return the correct schema for a pointer struct", func() {
				type TestCase struct {
					A string
					B int
				}
				schema := soda.GenerateSchemaRef(&TestCase{}, "")
				expected := openapi3.NewObjectSchema().
					WithProperty("A", openapi3.NewStringSchema()).
					WithProperty("B", openapi3.NewIntegerSchema()).
					WithRequired([]string{"A", "B"})
				So(schema.Value, ShouldResemble, expected)
				So(schema.Ref, ShouldEqual, "#/components/schemas/soda_test.TestCase")
			})

			Convey("It should return the correct schema for a generic struct", func() {
				type Container[T any] struct {
					Items []T `json:"items"`
					Total int `json:"total"`
				}
				schema := soda.GenerateSchemaRef(Container[string]{}, "json")
				expected := openapi3.NewObjectSchema().
					WithProperty("items", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
					WithProperty("total", openapi3.NewIntegerSchema()).
					WithRequired([]string{"items", "total"})
				So(schema.Value, ShouldResemble, expected)
			})

			Convey("It should return the correct schema for a struct slices", func() {
				type TestCase struct {
					A string `json:"a"`
				}
				schema := soda.GenerateSchemaRef([]TestCase{}, "json")
				expected := openapi3.NewArraySchema()
				expected.Items = openapi3.NewSchemaRef(
					"#/components/schemas/soda_test.TestCase",
					openapi3.NewObjectSchema().
						WithProperty("a", openapi3.NewStringSchema()).
						WithRequired([]string{"a"}),
				)
				So(schema.Value, ShouldResemble, expected)
			})

			Convey("It should return the correct schema for a complex struct", func() {
				type TestCase struct {
					String1 string     `json:"string1"`
					String2 *string    `json:"string2"`
					String3 []string   `json:"string3"`
					String4 *[]string  `json:"string4"`
					String5 []*string  `json:"string5"`
					String6 *[]*string `json:"string6"`
				}
				schema := soda.GenerateSchemaRef(TestCase{}, "json", "lol")
				expected := openapi3.NewObjectSchema().
					WithProperty("string1", openapi3.NewStringSchema()).
					WithProperty("string2", openapi3.NewStringSchema()).
					WithProperty("string3", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
					WithProperty("string4", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
					WithProperty("string5", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
					WithProperty("string6", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
					WithRequired([]string{"string1", "string3", "string5"})
				So(schema.Value, ShouldResemble, expected)
				So(schema.Ref, ShouldEqual, "#/components/schemas/lol")
			})

			Convey("It should return the correct schema for a struct with JSONSchema method", func() {
				schema := soda.GenerateSchemaRef(case4{}, "json")
				expect := openapi3.NewObjectSchema().
					WithProperty("x", openapi3.NewStringSchema().WithEnum("a", "b"))
				So(schema.Value, ShouldResemble, expect)
			})

			Convey("It should return the correct schema for a recursive struct", func() {
				type Node struct {
					Parent   *Node   `json:"parent"   oai:"description=recursive node"`
					Children []*Node `json:"children"`
				}
				schema := soda.GenerateSchemaRef(Node{}, "json")
				So(schema.Ref, ShouldEqual, "#/components/schemas/soda_test.Node")
			})

			Convey("It should panic for an anonymous struct", func() {
				So(func() { soda.GenerateSchemaRef(struct{}{}, "") }, ShouldPanic)
			})

			Convey("It should return the correct schema for a struct with embedded struct", func() {
				type Embedded struct {
					A string
				}
				type embeddedStruct struct {
					*Embedded
					B int
				}
				schema := soda.GenerateSchemaRef(embeddedStruct{}, "")
				expected := openapi3.NewObjectSchema().
					WithProperty("A", openapi3.NewStringSchema()).
					WithProperty("B", openapi3.NewIntegerSchema()).
					WithRequired([]string{"A", "B"})
				So(schema.Value, ShouldResemble, expected)
				So(schema.Ref, ShouldEqual, "#/components/schemas/soda_test.embeddedStruct")
			})

			Convey("It should return the correct schema for a list of structs", func() {
				type TestCase struct {
					A string
					B int
				}
				schema := soda.GenerateSchemaRef([]TestCase{}, "")
				itemsSchema := openapi3.NewObjectSchema().
					WithProperty("A", openapi3.NewStringSchema()).
					WithProperty("B", openapi3.NewIntegerSchema()).
					WithRequired([]string{"A", "B"})
				expected := openapi3.NewArraySchema()
				expected.Items = openapi3.NewSchemaRef("#/components/schemas/soda_test.TestCase", itemsSchema)
				So(schema.Value, ShouldEqual, expected)
			})

			Convey("It should ignore the field with ignore tag", func() {
				type ignoreStruct struct {
					A string
					B string `oai:"-"`
				}
				schema := soda.GenerateSchemaRef(ignoreStruct{}, "")
				expected := openapi3.NewObjectSchema().
					WithProperty("A", openapi3.NewStringSchema()).
					WithRequired([]string{"A"})
				So(schema.Value, ShouldEqual, expected)
			})

			Convey("It should panic for unsupported types", func() {
				So(func() { soda.GenerateSchemaRef(nil, "") }, ShouldPanic)
				So(func() { soda.GenerateSchemaRef(make(chan int), "") }, ShouldPanic)
			})
		})
	})

	Convey("Given parameters generation", t, func() {
		g := soda.NewGenerator()

		Convey("When providing a struct", func() {
			type testCase struct {
				A  string  `query:"a"`
				AP *string `query:"ap"`
				B  string  `header:"b"`
				BP *string `header:"bp"`
				C  string  `cookie:"c"`
				CP *string `cookie:"cp"`
				D  string  `path:"d"`
				DP *string `path:"dp"`
			}
			parameters := g.GenerateParameters(reflect.TypeOf(testCase{}))
			Convey("It should generate 8 parameters", func() {
				So(parameters, ShouldHaveLength, 8)
			})
			Convey("It should have correct parameter in the list", func() {
				So(parameters[0].Value, ShouldEqual,
					openapi3.
						NewQueryParameter("a").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				)

				So(parameters[1].Value, ShouldEqual,
					openapi3.
						NewQueryParameter("ap").
						WithSchema(openapi3.NewStringSchema()),
				)

				So(parameters[2].Value, ShouldEqual,
					openapi3.
						NewHeaderParameter("b").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				)

				So(parameters[3].Value, ShouldEqual,
					openapi3.
						NewHeaderParameter("bp").
						WithSchema(openapi3.NewStringSchema()),
				)

				So(parameters[4].Value, ShouldEqual,
					openapi3.
						NewCookieParameter("c").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				)

				So(parameters[5].Value, ShouldEqual,
					openapi3.
						NewCookieParameter("cp").
						WithSchema(openapi3.NewStringSchema()),
				)

				So(parameters[6].Value, ShouldEqual,
					openapi3.
						NewPathParameter("d").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				)

				So(parameters[7].Value, ShouldEqual,
					openapi3.
						NewPathParameter("dp").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				)
			})

			Convey("When providing a struct with a nested struct", func() {
				type TestCase1 struct {
					A string `query:"a"`
				}
				type TestCase2 struct {
					B string `query:"b"`
					TestCase1
				}
				parameters := g.GenerateParameters(reflect.TypeOf(TestCase2{}))
				Convey("It should generate 2 parameters", func() {
					So(parameters, ShouldHaveLength, 2)
				})
				Convey("It should have correct parameter in the list", func() {
					b := parameters[0].Value
					a := parameters[1].Value

					So(b, ShouldEqual, openapi3.NewQueryParameter("b").
						WithRequired(true).
						WithSchema(openapi3.NewStringSchema()),
					)
					So(a, ShouldEqual, openapi3.NewQueryParameter("a").
						WithRequired(true).
						WithSchema(openapi3.NewStringSchema()),
					)
				})
			})

			Convey("It should return nil for unsupported types", func() {
				parameters := g.GenerateParameters(reflect.TypeOf([]int{}))
				So(parameters, ShouldBeEmpty)
			})

			Convey("It should ignore some fields", func() {
				type schema struct {
					A string `query:"a"`
					B string `oai:"-"`
					C string
				}
				parameters := g.GenerateParameters(reflect.TypeOf(schema{}))
				So(parameters, ShouldHaveLength, 1)
			})

			Convey("It should generate sliced parameters", func() {
				type schema struct {
					A []string `oai:"description=This is a;explode"          query:"a"`
					B []string `oai:"description=This is b;style=deepObject" query:"b"`
				}
				parameters := g.GenerateParameters(reflect.TypeOf(schema{}))
				So(parameters, ShouldHaveLength, 2)
			})

			Convey("It should panic while invalid parameters", func() {
				type schema struct {
					A []string `query:"a"`
					B []string `query:"a"`
				}
				// duplicate parameter name should be meaningless
				So(func() { g.GenerateParameters(reflect.TypeOf(schema{})) }, ShouldPanic)
			})
		})
	})

	Convey("Given request body generation", t, func() {
		Convey("It should not be nil", func() {
			g := soda.NewGenerator()
			operationID := "testOperation"
			nameTag := "testNameTag"
			model := reflect.TypeOf(time.Time{})
			reqBody := g.GenerateRequestBody(operationID, nameTag, model)
			So(reqBody, ShouldNotBeNil)
		})
	})

	Convey("Given response generation", t, func() {
		g := soda.NewGenerator()
		Convey("It should generate correct response", func() {
			type test struct {
				A string `json:"a"`
				B int    `json:"b"`
			}

			mt := "application/json"
			resp := g.GenerateResponse(200, test{}, mt, "testing")
			So(resp, ShouldEqual,
				openapi3.NewResponse().
					WithDescription("testing").
					WithJSONSchemaRef(soda.GenerateSchemaRef(test{}, "json")),
			)
		})

		Convey("Providing nil should generate correct response", func() {
			resp := g.GenerateResponse(200, nil, "application/json", "testing")
			So(resp, ShouldEqual,
				openapi3.NewResponse().WithDescription("testing"),
			)
		})

		Convey("Providing an unsupported media-type should panic", func() {
			type test struct {
				A string `json:"a"`
				B int    `json:"b"`
			}

			So(func() {
				g.GenerateResponse(200, test{}, "application/json??", "testing")
			}, ShouldPanic)
		})
	})
}
