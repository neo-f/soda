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

type case1 struct {
	A string
	B int
}
type case2 struct {
	String1 string     `json:"string_1"`
	String2 *string    `json:"string_2"`
	String3 []string   `json:"string_3"`
	String4 *[]string  `json:"string_4"`
	String5 []*string  `json:"string_5"`
	String6 *[]*string `json:"string_6"`
}

type case3 struct {
	Node []*case3 `json:"node" oai:"description=recursive node"`
}

type case4 struct {
	X string `json:"x"`
}

func (c case4) JSONSchema(t *openapi3.T) *openapi3.SchemaRef {
	return openapi3.NewObjectSchema().
		WithProperty("x", openapi3.NewStringSchema().WithEnum("a", "b")).
		NewRef()
}

var _ = Describe("Soda", func() {
	Describe("NewGenerator", func() {
		It("should not be nil", func() {
			g := soda.NewGenerator()
			Expect(g).ShouldNot(BeNil())
		})
	})

	Describe("GenerateParameters", func() {
		It("should not be nil", func() {
			g := soda.NewGenerator()
			model := reflect.TypeOf(time.Time{})
			params := g.GenerateParameters(model)
			Expect(params).ShouldNot(BeNil())
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
		It("should not be nil", func() {
			g := soda.NewGenerator()
			code := 200
			model := reflect.TypeOf(time.Time{})
			mt := "application/json"
			description := "test description"
			resp := g.GenerateResponse(code, model, mt, description)
			Expect(resp).ShouldNot(BeNil())
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
			schema := soda.GenerateSchemaRef(case1{}, "")
			expected := openapi3.NewObjectSchema().
				WithProperty("A", openapi3.NewStringSchema()).
				WithProperty("B", openapi3.NewIntegerSchema()).
				WithRequired([]string{"A", "B"})
			Expect(schema.Value).To(Equal(expected))
			Expect(schema.Ref).To(Equal("#/components/schemas/soda_test.case1"))
		})

		It("should return the correct schema for a complex struct", func() {
			schema := soda.GenerateSchemaRef(case2{}, "json")
			expected := openapi3.NewObjectSchema().
				WithProperty("string_1", openapi3.NewStringSchema()).
				WithProperty("string_2", openapi3.NewStringSchema()).
				WithProperty("string_3", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithProperty("string_4", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithProperty("string_5", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithProperty("string_6", openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())).
				WithRequired([]string{"string_1", "string_3", "string_5"})
			Expect(schema.Value).To(Equal(expected))
			Expect(schema.Ref).To(Equal("#/components/schemas/soda_test.case2"))
		})

		It("should return the correct schema for a recursive struct", func() {
			schema := soda.GenerateSchemaRef(case3{}, "json")
			Expect(schema.Ref).To(Equal("#/components/schemas/soda_test.case3"))
		})

		It("should return the correct schema for a struct with JSONSchema method", func() {
			schema := soda.GenerateSchemaRef(case4{}, "json")
			expect := openapi3.NewObjectSchema().
				WithProperty("x", openapi3.NewStringSchema().WithEnum("a", "b"))
			Expect(schema.Value).To(Equal(expect))
		})
	})
})
