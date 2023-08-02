package soda

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGenerator_GetSchemaRef(t *testing.T) {
	g := NewGenerator()

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	ref := g.getSchemaRef(reflect.TypeOf(testStruct{}), "json", "")
	if ref.Ref != "#/components/schemas/soda.testStruct" {
		t.Errorf("Expected schema reference to have ref '#/components/schemas/soda.testStruct', but got '%s'", ref.Ref)
	}
	if ref.Value.Type != "object" {
		t.Errorf("Expected schema reference to have type 'object', but got '%s'", ref.Value.Type)
	}
	if len(ref.Value.Properties) != 2 {
		t.Errorf("Expected schema reference to have 2 properties, but got %d", len(ref.Value.Properties))
	}
}

func TestGenerator_GenerateCycleSchemaRef(t *testing.T) {
	g := NewGenerator()

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	ref := g.generateCycleSchemaRef(reflect.TypeOf([]testStruct{}), nil)
	if ref.Ref != "" {
		t.Errorf("Expected schema reference to have empty ref, but got '%s'", ref.Ref)
	}
	if ref.Value.Type != "array" {
		t.Errorf("Expected schema reference to have type 'array', but got '%s'", ref.Value.Type)
	}
	if ref.Value.Items == nil {
		t.Error("Expected schema reference to have non-nil items")
	}
	if ref.Value.Items.Ref != "#/components/schemas/soda.testStruct" {
		t.Errorf("Expected schema reference to have items ref '#/components/schemas/soda.testStruct', but got '%s'", ref.Value.Items.Ref)
	}
}

func TestGenerator_GenSchema(t *testing.T) {
	g := NewGenerator()

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	ref, _ := g.genSchema(nil, reflect.TypeOf(testStruct{}), "json")
	if ref.Ref != "" {
		t.Errorf("Expected schema reference to have empty ref, but got '%s'", ref.Ref)
	}
	if ref.Value.Type != "object" {
		t.Errorf("Expected schema reference to have type 'object', but got '%s'", ref.Value.Type)
	}
	if len(ref.Value.Properties) != 2 {
		t.Errorf("Expected schema reference to have 2 properties, but got %d", len(ref.Value.Properties))
	}
}

func TestGenerator_GenerateSchema(t *testing.T) {
	g := NewGenerator()

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	schema := g.GenerateSchema(testStruct{}, "json")
	if schema.Type != "object" {
		t.Errorf("Expected schema to have type 'object', but got '%s'", schema.Type)
	}
	if len(schema.Properties) != 2 {
		t.Errorf("Expected schema to have 2 properties, but got %d", len(schema.Properties))
	}
}

func TestGenerator_GenSchemaName(t *testing.T) {
	g := NewGenerator()

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	name := g.genSchemaName(reflect.TypeOf(testStruct{}))
	if name != "soda.testStruct" {
		t.Errorf("Expected schema name to be 'soda.testStruct', but got '%s'", name)
	}

	name = g.genSchemaName(reflect.TypeOf([]testStruct{}))
	if name != "soda.testStructList" {
		t.Errorf("Expected schema name to be 'testStructList', but got '%s'", name)
	}

	name = g.genSchemaName(reflect.TypeOf(time.Time{}))
	if name != "time.Time" {
		t.Errorf("Expected schema name to be 'time.Time', but got '%s'", name)
	}

	name = g.genSchemaName(reflect.TypeOf(&time.Time{}))
	if name != "time.Time" {
		t.Errorf("Expected schema name to be 'time.Time', but got '%s'", name)
	}

	name = g.genSchemaName(reflect.TypeOf(uint64(math.MaxUint64)))
	if name != "uint64" {
		t.Errorf("Expected schema name to be 'uint64', but got '%s'", name)
	}

	name = g.genSchemaName(reflect.TypeOf(strings.Builder{}))
	if name != "strings.Builder" {
		t.Errorf("Expected schema name to be 'strings.Builder', but got '%s'", name)
	}
}
