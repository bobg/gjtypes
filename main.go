// Command gjschema reads JSON data from stdin and writes Go types for parsing that data to stdout.
package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"sort"
	"strconv"

	"github.com/bobg/errors"
	"github.com/iancoleman/strcase"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var val any
	dec := json.NewDecoder(os.Stdin)
	dec.UseNumber()
	if err := dec.Decode(&val); err != nil {
		return errors.Wrap(err, "decoding JSON")
	}

	result := anyType
	if val != nil {
		result = schemaFor(val)
	}

	fmt.Printf("var data %s\n", rendered(result))

	for i := 1; i <= len(structNames); i++ {
		fmt.Println()

		name := fmt.Sprintf("S%03d", i)
		typ := structsByName[name]

		fmt.Printf("type %s struct {\n", name)

		for fieldNum := range typ.NumField() {
			field := typ.Field(fieldNum)
			fmt.Printf("  %s %s `%s`\n", field.Name, rendered(field.Type), field.Tag)
		}

		fmt.Println("}")
	}

	return nil
}

var (
	anyType        = reflect.TypeFor[any]()
	float64Type    = reflect.TypeFor[float64]()
	int64Type      = reflect.TypeFor[int64]()
	jsonNumberType = reflect.TypeFor[json.Number]()
	stringType     = reflect.TypeFor[string]()
	undefinedType  = reflect.TypeFor[undefined]()
)

type undefined struct{}

func schemaFor(inp any) reflect.Type {
	val := reflect.ValueOf(inp)

	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if val.IsNil() {
			return undefinedType
		}
	}

	typ := val.Type()

	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int64Type

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64Type

	case reflect.Float32, reflect.Float64:
		return float64Type

	case reflect.Map:
		if val.Len() == 0 {
			return undefinedType
		}

		fields := make(map[string]reflect.StructField)

		mapRange := val.MapRange()
		for mapRange.Next() {
			key, elem := mapRange.Key(), mapRange.Value()

			if key.Type() != stringType {
				return anyType // xxx or map[x]y ?
			}
			origFieldName := key.String()
			fieldName := strcase.ToCamel(origFieldName)
			if _, ok := fields[fieldName]; ok {
				// Field name collision.
				return anyType
			}

			fields[fieldName] = reflect.StructField{
				Name: fieldName,
				Type: schemaFor(elem.Interface()),
				Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s,omitempty"`, origFieldName)),
			}
		}

		return structOf(fields)

	case reflect.Slice:
		if val.Len() == 0 {
			return reflect.SliceOf(undefinedType)
		}

		result := schemaFor(val.Index(0).Interface())
		if result == anyType {
			return reflect.SliceOf(result)
		}

		for i := 1; i < val.Len(); i++ {
			elem := val.Index(i)
			result = updateSchemaFor(result, elem.Interface())
			if result == anyType {
				return reflect.SliceOf(result)
			}
		}

		return reflect.SliceOf(result)

	case reflect.Bool:
		return typ

	case reflect.String:
		if typ == jsonNumberType {
			if _, err := strconv.ParseInt(val.String(), 10, 64); err == nil { // sic
				return int64Type
			}
			return float64Type
		}
		return stringType

	default:
		return anyType
	}
}

func updateSchemaFor(typ reflect.Type, val any) reflect.Type {
	return unifyTypes(typ, schemaFor(val))
}

func unifyTypes(orig, other reflect.Type) reflect.Type {
	if orig == anyType {
		return anyType
	}

	if orig == other {
		return orig
	}

	if orig == undefinedType {
		return other
	}
	if other == undefinedType {
		return orig
	}

	switch orig {
	case float64Type:
		if other == int64Type {
			return float64Type
		}
		fallthrough

	case int64Type:
		if other == stringType {
			return stringType
		}
		return anyType

	case stringType:
		return anyType
	}

	if orig.Kind() == reflect.Slice {
		if other.Kind() != reflect.Slice {
			return anyType
		}

		elemType := unifyTypes(orig.Elem(), other.Elem())
		return reflect.SliceOf(elemType)
	}

	if orig.Kind() != reflect.Struct || other.Kind() != reflect.Struct {
		return anyType
	}

	fields := make(map[string]reflect.StructField)

	for fieldNum := range orig.NumField() {
		field := orig.Field(fieldNum)
		fields[field.Name] = field
	}
	for fieldNum := range other.NumField() {
		field := other.Field(fieldNum)
		if origField, ok := fields[field.Name]; ok {
			fields[field.Name] = reflect.StructField{
				Name: field.Name,
				Type: unifyTypes(origField.Type, field.Type),
				Tag:  origField.Tag,
			}
		} else {
			fields[field.Name] = field
		}
	}

	return structOf(fields)
}

func structOf(fields map[string]reflect.StructField) reflect.Type {
	fieldSlice := slices.Collect(maps.Values(fields))
	sort.Slice(fieldSlice, func(i, j int) bool { return fieldSlice[i].Name < fieldSlice[j].Name })
	return reflect.StructOf(fieldSlice)
}

var (
	structNames   = make(map[reflect.Type]string)
	structsByName = make(map[string]reflect.Type)
)

func rendered(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Struct:
		name, ok := structNames[typ]
		if !ok {
			name = fmt.Sprintf("S%03d", len(structNames)+1)
			structNames[typ] = name
			structsByName[name] = typ
		}
		return "*" + name

	case reflect.Slice:
		return "[]" + rendered(typ.Elem())
	}

	return typ.String()
}
