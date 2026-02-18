package reflection

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// GetTagNameByValue searches struct s (or *s) for a field whose value equals fieldValue
// and returns the tag for tagKey (e.g. "json") for the first match.
// It returns an error if s is not a struct or no matching field is found.
//
// Notes & limitations:
// - It compares field values using reflect.DeepEqual.
// - Unexported fields are skipped (fv.CanInterface() == false).
// - If multiple fields have identical values, the first match is returned (search order is struct field order).
// - This requires passing the originating struct instance; a bare field value alone is insufficient in Go.
func GetTagNameByValue(strukt, fieldValue any, tagKey string) (string, error) {
	v := reflect.ValueOf(strukt)
	if !v.IsValid() {
		return "", fmt.Errorf("nil value")
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", fmt.Errorf("nil pointer to struct")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("GetTagNameByValue: expected struct or pointer to struct, got %strukt", v.Kind())
	}

	t := v.Type()
	var walk func(reflect.Value, reflect.Type) (string, bool)
	walk = func(rv reflect.Value, rt reflect.Type) (string, bool) {
		for i := 0; i < rt.NumField(); i++ {
			sf := rt.Field(i)
			fv := rv.Field(i)

			// if embedded struct, recurse
			if sf.Anonymous {
				// handle pointer-to-embedded
				fvKind := fv.Kind()
				if fvKind == reflect.Ptr && !fv.IsNil() {
					fv = fv.Elem()
					fvKind = fv.Kind()
				}
				if fvKind == reflect.Struct {
					if tag, ok := walk(fv, fv.Type()); ok {
						return tag, true
					}
				}
				continue
			}

			// skip unexported fields we can't interface with
			if !fv.IsValid() || !fv.CanInterface() {
				continue
			}

			if reflect.DeepEqual(fv.Interface(), fieldValue) {
				return sf.Tag.Get(tagKey), true
			}
		}
		return "", false
	}

	if tag, ok := walk(v, t); ok {
		return tag, nil
	}
	return "", fmt.Errorf("no matching field with that value found in struct")
}

// GetMethodName is meant to fetch the name of a given method passed in as an argument.
func GetMethodName(method any) string {
	v := reflect.ValueOf(method)
	if v.Kind() != reflect.Func {
		return ""
	}

	if pc := v.Pointer(); pc != 0 {
		if f := runtime.FuncForPC(pc); f != nil {
			fullName := f.Name()
			parts := strings.Split(fullName, ".")
			if len(parts) > 0 {
				return strings.TrimSuffix(parts[len(parts)-1], "-fm")
			}
		}
	}
	return ""
}

// GetFieldTypes returns a map of field names to their types. For nested structs,
// the value is a map[string]any containing the nested struct's fields.
func GetFieldTypes(strukt any) (map[string]any, error) {
	var t reflect.Type

	switch v := strukt.(type) {
	case reflect.Type:
		t = v
	default:
		rv := reflect.ValueOf(strukt)
		if !rv.IsValid() {
			return nil, fmt.Errorf("nil value")
		}

		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				// For nil pointer, try to get the type from the pointer type
				t = rv.Type().Elem()
			} else {
				t = rv.Elem().Type()
			}
		} else {
			t = rv.Type()
		}
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("GetFieldTypes: expected struct or pointer to struct, got %v", t.Kind())
	}

	result := make(map[string]any)

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		fieldType := sf.Type

		// Skip unexported fields
		if !sf.IsExported() {
			continue
		}

		// Check if it's a pointer type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// If it's a struct, recursively get its fields
		if fieldType.Kind() == reflect.Struct {
			nestedMap, err := GetFieldTypes(fieldType)
			if err != nil {
				return nil, fmt.Errorf("error processing nested struct field %s: %w", sf.Name, err)
			}
			result[sf.Name] = nestedMap
		} else {
			// For non-struct fields, store the type string
			result[sf.Name] = fieldType.String()
		}
	}

	return result, nil
}
