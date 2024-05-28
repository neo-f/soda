package soda_test

import (
	"encoding/json"
	"math"
	"net"
	"reflect"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/neo-f/soda/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type case4 struct {
	X string `json:"x"`
}

func (c case4) JSONSchema(t *openapi3.T) *openapi3.SchemaRef {
	return openapi3.NewObjectSchema().
		WithProperty("x", openapi3.NewStringSchema().WithEnum("a", "b")).
		NewRef()
}

var _ = Describe("Generator", func() {
	Describe("NewGenerator", func() {
		It("should not be nil", func() {
			g := soda.NewGenerator()
			Expect(g).ShouldNot(BeNil())
		})
	})

	Describe("GenerateSchemaRef", func() {
		It("should return the correct schema for string", func() {
			schema := soda.GenerateSchemaRef("", "")
			Expect(schema).To(Equal(openapi3.NewStringSchema().NewRef()))
		})

		DescribeTable("integers",
			func(t any, schema *openapi3.SchemaRef) {
				Expect(soda.GenerateSchemaRef(t, "")).To(Equal(schema))
			},

			Entry("should return the correct schema for int", int(0), openapi3.NewIntegerSchema().NewRef()),
			Entry("should return the correct schema for uint", uint(0), openapi3.NewIntegerSchema().WithMin(0).NewRef()),
			Entry("should return the correct schema for int8", int8(0), openapi3.NewIntegerSchema().WithMin(math.MinInt8).WithMax(math.MaxInt8).NewRef()),
			Entry("should return the correct schema for uint8", uint8(0), openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint8).NewRef()),
			Entry("should return the correct schema for int16", int16(0), openapi3.NewIntegerSchema().WithMin(math.MinInt16).WithMax(math.MaxInt16).NewRef()),
			Entry("should return the correct schema for uint16", uint16(0), openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint16).NewRef()),
			Entry("should return the correct schema for int32", int32(0), openapi3.NewInt32Schema().WithMin(math.MinInt32).WithMax(math.MaxInt32).NewRef()),
			Entry("should return the correct schema for uint32", uint32(0), openapi3.NewInt32Schema().WithMin(0).WithMax(math.MaxUint32).NewRef()),
			Entry("should return the correct schema for int64", int64(0), openapi3.NewInt64Schema().WithMin(math.MinInt64).WithMax(math.MaxInt64).NewRef()),
			Entry("should return the correct schema for uint64", uint64(0), openapi3.NewInt64Schema().WithMin(0).WithMax(math.MaxUint64).NewRef()),
		)

		It("should return the correct schema for float", func() {
			Expect(soda.GenerateSchemaRef(float32(0), "")).To(Equal(openapi3.NewFloat64Schema().NewRef()))
			Expect(soda.GenerateSchemaRef(float64(0), "")).To(Equal(openapi3.NewFloat64Schema().NewRef()))
		})

		It("should return the correct schema for boolean", func() {
			schema := soda.GenerateSchemaRef(true, "")
			Expect(schema).To(Equal(openapi3.NewBoolSchema().NewRef()))
		})

		It("should return the correct schema for map[string]any", func() {
			schema := soda.GenerateSchemaRef(map[string]any{}, "")
			Expect(schema).To(Equal(openapi3.NewObjectSchema().WithAnyAdditionalProperties().NewRef()))
		})

		It("should return the correct schema for time.Time", func() {
			schema := soda.GenerateSchemaRef(time.Time{}, "")
			Expect(schema).To(Equal(openapi3.NewStringSchema().WithFormat("date-time").NewRef()))
		})

		It("should return the correct schema for net.IP", func() {
			schema := soda.GenerateSchemaRef(net.IP{}, "")
			Expect(schema).To(Equal(openapi3.NewStringSchema().WithFormat("ipv4").NewRef()))
		})

		It("should return the correct schema for json.RawMessage", func() {
			schema := soda.GenerateSchemaRef(json.RawMessage{}, "")
			Expect(schema).To(Equal(openapi3.NewStringSchema().WithFormat("json").NewRef()))
		})

		It("should return the correct schema for []byte", func() {
			schema := soda.GenerateSchemaRef([]byte{}, "")
			Expect(schema).To(Equal(openapi3.NewBytesSchema().NewRef()))
		})

		It("should return the correct schema for array", func() {
			schema := soda.GenerateSchemaRef([2]int{}, "")
			expected := openapi3.NewArraySchema().
				WithMaxItems(2).
				WithMinItems(2).
				WithItems(openapi3.NewIntegerSchema()).NewRef()
			Expect(schema).To(Equal(expected))
		})

		It("should return the correct schema for slice", func() {
			schema := soda.GenerateSchemaRef([]int{}, "")
			Expect(schema).To(Equal(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).NewRef()))
		})

		It("should return the correct schema for map[string]int", func() {
			schema := soda.GenerateSchemaRef(map[string]int{}, "")
			Expect(schema).To(Equal(openapi3.NewObjectSchema().WithAdditionalProperties(openapi3.NewIntegerSchema()).NewRef()))
		})

		It("should return the correct schema for a basic struct", func() {
			type TestCase struct {
				A string
				B int
			}
			schema := soda.GenerateSchemaRef(TestCase{}, "")
			expected := openapi3.NewObjectSchema().
				WithProperty("A", openapi3.NewStringSchema()).
				WithProperty("B", openapi3.NewIntegerSchema()).
				WithRequired([]string{"A", "B"})
			Expect(schema.Value).To(Equal(expected))
			Expect(schema.Ref).To(Equal("#/components/schemas/soda_test.TestCase"))
		})

		It("should return the correct schema for a pointer struct", func() {
			type TestCase struct {
				A string
				B int
			}
			schema := soda.GenerateSchemaRef(&TestCase{}, "")
			expected := openapi3.NewObjectSchema().
				WithProperty("A", openapi3.NewStringSchema()).
				WithProperty("B", openapi3.NewIntegerSchema()).
				WithRequired([]string{"A", "B"})
			Expect(schema.Value).To(Equal(expected))
			Expect(schema.Ref).To(Equal("#/components/schemas/soda_test.TestCase"))
		})

		It("should return the correct schema for a complex struct", func() {
			type TestCase struct {
				String1 string     `json:"string_1"`
				String2 *string    `json:"string_2"`
				String3 []string   `json:"string_3"`
				String4 *[]string  `json:"string_4"`
				String5 []*string  `json:"string_5"`
				String6 *[]*string `json:"string_6"`
			}
			schema := soda.GenerateSchemaRef(TestCase{}, "json", "lol")
			expected := openapi3.NewObjectSchema().
				WithProperty("string_1", openapi3.NewStringSchema()).
				WithProperty("string_2", openapi3.NewStringSchema()).
				WithProperty("string_3", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithProperty("string_4", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithProperty("string_5", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithProperty("string_6", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithRequired([]string{"string_1", "string_3", "string_5"})
			Expect(schema.Value).To(Equal(expected))
			Expect(schema.Ref).To(Equal("#/components/schemas/lol"))
		})

		It("should return the correct schema for a struct with JSONSchema method", func() {
			schema := soda.GenerateSchemaRef(case4{}, "json")
			expect := openapi3.NewObjectSchema().
				WithProperty("x", openapi3.NewStringSchema().WithEnum("a", "b"))
			Expect(schema.Value).To(Equal(expect))
		})

		It("should return the correct schema for a recursive struct", func() {
			type Node struct {
				Parent   *Node   `json:"parent" oai:"description=recursive node"`
				Children []*Node `json:"children"`
			}
			schema := soda.GenerateSchemaRef(Node{}, "json")
			Expect(schema.Ref).To(Equal("#/components/schemas/soda_test.Node"))
		})

		It("should panic for a anonymous struct", func() {
			Expect(func() { soda.GenerateSchemaRef(struct{}{}, "") }).To(Panic())
		})

		It("should return the correct schema for a struct with embedded struct", func() {
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
			Expect(schema.Value).To(Equal(expected))
			Expect(schema.Ref).To(Equal("#/components/schemas/soda_test.embeddedStruct"))
		})

		// list of structs
		It("should return the correct schema for a list of structs", func() {
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
			Expect(schema.Value).To(Equal(expected))
		})

		It("should ignore the field with ignore tag", func() {
			type ignoreStruct struct {
				A string
				B string `oai:"-"`
			}
			schema := soda.GenerateSchemaRef(ignoreStruct{}, "")
			expected := openapi3.NewObjectSchema().
				WithProperty("A", openapi3.NewStringSchema()).
				WithRequired([]string{"A"})
			Expect(schema.Value).To(Equal(expected))
		})

		It("should panic for unsupported types", func() {
			Expect(func() {
				soda.GenerateSchemaRef(nil, "")
			}).To(Panic())

			Expect(func() {
				soda.GenerateSchemaRef(make(chan int), "")
			}).To(Panic())
		})
	})

	Describe("GenerateParameters", func() {
		g := soda.NewGenerator()
		When("provide a struct", func() {
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
			It("should generated 8 parameters", func() {
				Expect(parameters).To(HaveLen(8))
			})
			It("should have correct parameter in the list", func() {
				Expect(parameters[0].Value).To(Equal(
					openapi3.
						NewQueryParameter("a").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				))

				Expect(parameters[1].Value).To(Equal(
					openapi3.
						NewQueryParameter("ap").
						WithSchema(openapi3.NewStringSchema()),
				))

				Expect(parameters[2].Value).To(Equal(
					openapi3.
						NewHeaderParameter("b").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				))

				Expect(parameters[3].Value).To(Equal(
					openapi3.
						NewHeaderParameter("bp").
						WithSchema(openapi3.NewStringSchema()),
				))

				Expect(parameters[4].Value).To(Equal(
					openapi3.
						NewCookieParameter("c").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				))

				Expect(parameters[5].Value).To(Equal(
					openapi3.
						NewCookieParameter("cp").
						WithSchema(openapi3.NewStringSchema()),
				))
				Expect(parameters[6].Value).To(Equal(
					openapi3.
						NewPathParameter("d").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				))

				Expect(parameters[7].Value).To(Equal(
					openapi3.
						NewPathParameter("dp").
						WithSchema(openapi3.NewStringSchema()).
						WithRequired(true),
				))
			})
		})

		When("provide a struct with a nested struct", func() {
			type TestCase1 struct {
				A string `query:"a"`
			}
			type TestCase2 struct {
				B string `query:"b"`
				TestCase1
			}
			parameters := g.GenerateParameters(reflect.TypeOf(TestCase2{}))
			It("should generated 2 parameters", func() {
				Expect(parameters).To(HaveLen(2))
			})
			It("should have correct parameter in the list", func() {
				b := parameters[0].Value
				a := parameters[1].Value

				Expect(b).To(Equal(openapi3.NewQueryParameter("b").
					WithRequired(true).
					WithSchema(openapi3.NewStringSchema()),
				))
				Expect(a).To(Equal(openapi3.NewQueryParameter("a").
					WithRequired(true).
					WithSchema(openapi3.NewStringSchema()),
				))
			})
		})

		It("should return nil for unsupported types", func() {
			parameters := g.GenerateParameters(reflect.TypeOf([]int{}))
			Expect(parameters).To(BeEmpty())
		})

		It("should ignore some fields", func() {
			type schema struct {
				A string `query:"a"`
				B string `oai:"-"`
				C string
			}
			parameters := g.GenerateParameters(reflect.TypeOf(schema{}))
			Expect(parameters).To(HaveLen(1))
		})

		It("should generate sliced parameters", func() {
			type schema struct {
				A []string `query:"a" oai:"description=This is a;explode"`
				B []string `query:"b" oai:"description=This is b;style=deepObject"`
			}
			parameters := g.GenerateParameters(reflect.TypeOf(schema{}))
			Expect(parameters).To(HaveLen(2))
		})

		It("should panic while invalid parameters", func() {
			type schema struct {
				A []string `query:"a"`
				B []string `query:"a"`
			}
			// duplicate parameter name should be meaningless
			Expect(func() { g.GenerateParameters(reflect.TypeOf(schema{})) }).To(Panic())
		})
	})

	Describe("GenerateRequestBody", func() {
		It("should not be nil", func() {
			g := soda.NewGenerator()
			operationID := "testOperation"
			nameTag := "testNameTag"
			model := reflect.TypeOf(time.Time{})
			reqBody := g.GenerateRequestBody(operationID, nameTag, model)
			Expect(reqBody).ShouldNot(BeNil())
		})
	})

	Describe("GenerateResponse", func() {
		g := soda.NewGenerator()
		When("provide a struct", func() {
			It("should generate correct response ", func() {
				type test struct {
					A string `json:"a"`
					B int    `json:"b"`
				}

				mt := "application/json"
				resp := g.GenerateResponse(200, test{}, mt, "testing")
				Expect(resp).To(Equal(
					openapi3.NewResponse().
						WithDescription("testing").
						WithJSONSchemaRef(soda.GenerateSchemaRef(test{}, "json")),
				))
			})
		})

		When("provide nil", func() {
			It("should generate correct response ", func() {
				resp := g.GenerateResponse(200, nil, "application/json", "testing")
				Expect(resp).To(Equal(
					openapi3.NewResponse().WithDescription("testing"),
				))
			})
		})

		When("provide a unsupported media-type", func() {
			It("should panic", func() {
				type test struct {
					A string `json:"a"`
					B int    `json:"b"`
				}

				Expect(func() {
					g.GenerateResponse(200, test{}, "application/json??", "testing")
				}).To(Panic())
			})
		})
	})
})
