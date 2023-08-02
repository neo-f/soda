package soda

import (
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestNewFieldResolver(t *testing.T) {
	type testStruct struct {
		Name string `json:"name" oai:"title=Name;description=The name of the test struct"`
	}

	field := reflect.TypeOf(testStruct{}).Field(0)
	resolver := newFieldResolver(&field)

	if resolver.t != &field {
		t.Errorf("Expected resolver.t to be %v, but got %v", &field, resolver.t)
	}

	if resolver.ignored {
		t.Errorf("Expected resolver.ignored to be false, but got true")
	}

	if resolver.tagPairs == nil {
		t.Errorf("Expected resolver.tagPairs to be initialized, but got nil")
	}

	if resolver.tagPairs["title"] != "Name" {
		t.Errorf("Expected resolver.tagPairs[\"title\"] to be \"Name\", but got %v", resolver.tagPairs["title"])
	}

	if resolver.tagPairs["description"] != "The name of the test struct" {
		t.Errorf("Expected resolver.tagPairs[\"description\"] to be \"The name of the test struct\", but got %v", resolver.tagPairs["description"])
	}
}

func TestFieldResolver_InjectOAITags(t *testing.T) {
	schema := &openapi3.Schema{
		Type: openapi3.TypeString,
	}

	resolver := &fieldResolver{
		tagPairs: map[string]string{
			propTitle:           "Test Title",
			propDescription:     "Test Description",
			propDeprecated:      "true",
			propAllowEmptyValue: "false",
			propNullable:        "true",
			propWriteOnly:       "false",
			propReadOnly:        "true",
		},
	}

	resolver.injectOAITags(schema)

	if schema.Title != "Test Title" {
		t.Errorf("Expected schema.Title to be \"Test Title\", but got %v", schema.Title)
	}

	if schema.Description != "Test Description" {
		t.Errorf("Expected schema.Description to be \"Test Description\", but got %v", schema.Description)
	}

	if !schema.Deprecated {
		t.Errorf("Expected schema.Deprecated to be true, but got false")
	}

	if schema.AllowEmptyValue {
		t.Errorf("Expected schema.AllowEmptyValue to be false, but got true")
	}

	if !schema.Nullable {
		t.Errorf("Expected schema.Nullable to be true, but got false")
	}

	if schema.WriteOnly {
		t.Errorf("Expected schema.WriteOnly to be false, but got true")
	}

	if !schema.ReadOnly {
		t.Errorf("Expected schema.ReadOnly to be true, but got false")
	}
}

func TestFieldResolver_Required(t *testing.T) {
	type testStruct struct {
		Name string `json:"name" oai:"required=true;nullable=false"`
		Age  *int   `json:"age" oai:"required=false;nullable=true"`
	}

	nameField := reflect.TypeOf(testStruct{}).Field(0)
	nameResolver := newFieldResolver(&nameField)

	if !nameResolver.required() {
		t.Errorf("Expected nameResolver.required() to be true, but got false")
	}

	ageField := reflect.TypeOf(testStruct{}).Field(1)
	ageResolver := newFieldResolver(&ageField)

	if ageResolver.required() {
		t.Errorf("Expected ageResolver.required() to be false, but got true")
	}
}

func TestFieldResolver_Name(t *testing.T) {
	type testStruct struct {
		Name string `json:"name" oai:"title=Test Name"`
	}

	nameField := reflect.TypeOf(testStruct{}).Field(0)
	nameResolver := newFieldResolver(&nameField)

	if nameResolver.name() != "Name" {
		t.Errorf("Expected nameResolver.name() to be \"Name\", but got %v", nameResolver.name())
	}

	if nameResolver.name("json") != "name" {
		t.Errorf("Expected nameResolver.name(\"json\") to be \"name\", but got %v", nameResolver.name("json"))
	}
}

func TestFieldResolver_ShouldEmbed(t *testing.T) {
	type testStruct struct {
		Name string `json:"name" oai:"title=Test Name"`
	}

	nameField := reflect.TypeOf(testStruct{}).Field(0)
	nameResolver := newFieldResolver(&nameField)

	if nameResolver.shouldEmbed() {
		t.Errorf("Expected nameResolver.shouldEmbed() to be false, but got true")
	}

	anonField := reflect.TypeOf(struct {
		testStruct
	}{}).Field(0)
	anonResolver := newFieldResolver(&anonField)

	if !anonResolver.shouldEmbed() {
		t.Errorf("Expected anonResolver.shouldEmbed() to be true, but got false")
	}
}

func TestFieldResolver_InjectOAIString(t *testing.T) {
	schema := &openapi3.Schema{
		Type: openapi3.TypeString,
	}

	resolver := &fieldResolver{
		tagPairs: map[string]string{
			propMinLength: "5",
			propMaxLength: "10",
			propPattern:   "^[a-z]+$",
			propFormat:    "email",
			propEnum:      "foo,bar,baz",
			propDefault:   "default",
			propExample:   "example",
		},
	}

	resolver.injectOAIString(schema)

	if schema.MinLength != 5 {
		t.Errorf("Expected schema.MinLength to be 5, but got %v", schema.MinLength)
	}

	if *schema.MaxLength != 10 {
		t.Errorf("Expected schema.MaxLength to be 10, but got %v", *schema.MaxLength)
	}

	if schema.Pattern != "^[a-z]+$" {
		t.Errorf("Expected schema.Pattern to be \"^[a-z]+$\", but got %v", schema.Pattern)
	}

	if schema.Format != "email" {
		t.Errorf("Expected schema.Format to be \"email\", but got %v", schema.Format)
	}

	if len(schema.Enum) != 3 || schema.Enum[0] != "foo" || schema.Enum[1] != "bar" || schema.Enum[2] != "baz" {
		t.Errorf("Expected schema.Enum to be [\"foo\", \"bar\", \"baz\"], but got %v", schema.Enum)
	}

	if schema.Default != "default" {
		t.Errorf("Expected schema.Default to be \"default\", but got %v", schema.Default)
	}

	if schema.Example != "example" {
		t.Errorf("Expected schema.Example to be \"example\", but got %v", schema.Example)
	}
}

func TestFieldResolver_InjectOAINumeric(t *testing.T) {
	schema := &openapi3.Schema{
		Type: openapi3.TypeNumber,
	}

	resolver := &fieldResolver{
		tagPairs: map[string]string{
			propMultipleOf:       "2",
			propMinimum:          "0",
			propMaximum:          "10",
			propExclusiveMaximum: "true",
			propExclusiveMinimum: "false",
			propDefault:          "5",
			propExample:          "7",
			propEnum:             "1,2,3",
		},
	}

	resolver.injectOAINumeric(schema)

	if *schema.MultipleOf != 2 {
		t.Errorf("Expected schema.MultipleOf to be 2, but got %v", *schema.MultipleOf)
	}

	if *schema.Min != 0 {
		t.Errorf("Expected schema.Min to be 0, but got %v", *schema.Min)
	}

	if *schema.Max != 10 {
		t.Errorf("Expected schema.Max to be 10, but got %v", *schema.Max)
	}

	if !schema.ExclusiveMax {
		t.Errorf("Expected schema.ExclusiveMax to be true, but got false")
	}

	if schema.ExclusiveMin {
		t.Errorf("Expected schema.ExclusiveMin to be false, but got true")
	}

	if schema.Default != 5.0 {
		t.Errorf("Expected schema.Default to be 5.0, but got %v", schema.Default)
	}

	if schema.Example != 7.0 {
		t.Errorf("Expected schema.Example to be 7.0, but got %v", schema.Example)
	}

	if len(schema.Enum) != 3 || schema.Enum[0] != 1.0 || schema.Enum[1] != 2.0 || schema.Enum[2] != 3.0 {
		t.Errorf("Expected schema.Enum to be [1.0, 2.0, 3.0], but got %v", schema.Enum)
	}
}

func TestFieldResolver_InjectOAIArray(t *testing.T) {
	schema := &openapi3.Schema{
		Type: openapi3.TypeArray,
		Items: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: openapi3.TypeString,
			},
		},
	}

	resolver := &fieldResolver{
		tagPairs: map[string]string{
			propMinItems:    "2",
			propMaxItems:    "5",
			propUniqueItems: "true",
			propDefault:     "foo,bar,baz",
			propExample:     "baz,bar,foo",
			propEnum:        "foo,bar,baz",
		},
	}

	resolver.injectOAIArray(schema)

	if schema.MinItems != 2 {
		t.Errorf("Expected schema.MinItems to be 2, but got %v", schema.MinItems)
	}

	if *schema.MaxItems != 5 {
		t.Errorf("Expected schema.MaxItems to be 5, but got %v", *schema.MaxItems)
	}

	if !schema.UniqueItems {
		t.Errorf("Expected schema.UniqueItems to be true, but got false")
	}
	if len(schema.Default.([]interface{})) != 3 || schema.Default.([]interface{})[0] != "foo" || schema.Default.([]interface{})[1] != "bar" || schema.Default.([]interface{})[2] != "baz" {
		t.Errorf("Expected schema.Default to be [\"foo\", \"bar\", \"baz\"], but got %v", schema.Default)
	}

	if len(schema.Example.([]interface{})) != 3 || schema.Example.([]interface{})[0] != "baz" || schema.Example.([]interface{})[1] != "bar" || schema.Example.([]interface{})[2] != "foo" {
		t.Errorf("Expected schema.Example to be [\"baz\", \"bar\", \"foo\"], but got %v", schema.Example)
	}

	if len(schema.Enum) != 3 || schema.Enum[0].(string) != "foo" || schema.Enum[1].(string) != "bar" || schema.Enum[2].(string) != "baz" {
		t.Errorf("Expected schema.Enum to be [[\"foo\"], [\"bar\"], [\"baz\"]], but got %v", schema.Enum)
	}
}

func TestFieldResolver_InjectOAIBoolean(t *testing.T) {
	schema := &openapi3.Schema{
		Type: openapi3.TypeBoolean,
	}

	resolver := &fieldResolver{
		tagPairs: map[string]string{
			propDefault: "true",
			propExample: "false",
		},
	}

	resolver.injectOAIBoolean(schema)

	if !schema.Default.(bool) {
		t.Errorf("Expected schema.Default to be true, but got false")
	}

	if schema.Example.(bool) {
		t.Errorf("Expected schema.Example to be false, but got true")
	}
}
