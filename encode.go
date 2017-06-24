// Package httpheader is a Go library for encoding structs into Header fields.
package httpheader

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const tagName = "header"

// Version ...
const Version = "0.2.0"

var timeType = reflect.TypeOf(time.Time{})
var headerType = reflect.TypeOf(http.Header{})

var encoderType = reflect.TypeOf(new(Encoder)).Elem()

// Encoder ...
type Encoder interface {
	EncodeHeader(key string, v *http.Header) error
}

// Header returns the http.Header encoding of v.
func Header(v interface{}) (http.Header, error) {
	h := make(http.Header)
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return h, nil
		}
		val = val.Elem()
	}

	if v == nil {
		return h, nil
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("httpheader: Header() expects struct input. Got %v", val.Kind())
	}

	err := reflectValue(h, val)
	return h, err
}

// reflectValue populates the values parameter from the struct fields in val.
// Embedded structs are followed recursively (using the rules defined in the
// Values function documentation) breadth-first.
func reflectValue(header http.Header, val reflect.Value) error {
	var embedded []reflect.Value

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}

		sv := val.Field(i)
		tag := sf.Tag.Get(tagName)
		if tag == "-" {
			continue
		}
		name, opts := parseTag(tag)
		if name == "" {
			if sf.Anonymous && sv.Kind() == reflect.Struct {
				// save embedded struct for later processing
				embedded = append(embedded, sv)
				continue
			}

			name = sf.Name
		}

		if opts.Contains("omitempty") && isEmptyValue(sv) {
			continue
		}

		if sv.Type().Implements(encoderType) {
			if !reflect.Indirect(sv).IsValid() {
				sv = reflect.New(sv.Type().Elem())
			}

			m := sv.Interface().(Encoder)
			if err := m.EncodeHeader(name, &header); err != nil {
				return err
			}
			continue
		}

		if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
			for i := 0; i < sv.Len(); i++ {
				k := name
				header.Add(k, valueString(sv.Index(i), opts))
			}
			continue
		}

		for sv.Kind() == reflect.Ptr {
			if sv.IsNil() {
				break
			}
			sv = sv.Elem()
		}

		if sv.Type() == timeType {
			header.Add(name, valueString(sv, opts))
			continue
		}
		if sv.Type() == headerType {
			h := sv.Interface().(http.Header)
			for k, vs := range h {
				for _, v := range vs {
					header.Add(k, v)
				}
			}
			continue
		}

		if sv.Kind() == reflect.Struct {
			reflectValue(header, sv)
			continue
		}

		header.Add(name, valueString(sv, opts))
	}

	for _, f := range embedded {
		if err := reflectValue(header, f); err != nil {
			return err
		}
	}

	return nil
}

// valueString returns the string representation of a value.
func valueString(v reflect.Value, opts tagOptions) string {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if v.Kind() == reflect.Bool && opts.Contains("int") {
		if v.Bool() {
			return "1"
		}
		return "0"
	}

	if v.Type() == timeType {
		t := v.Interface().(time.Time)
		if opts.Contains("unix") {
			return strconv.FormatInt(t.Unix(), 10)
		}
		return t.Format(http.TimeFormat)
	}

	return fmt.Sprint(v.Interface())
}

// isEmptyValue checks if a value should be considered empty for the purposes
// of omitting fields with the "omitempty" option.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	if v.Type() == timeType {
		return v.Interface().(time.Time).IsZero()
	}

	return false
}

// tagOptions is the string following a comma in a struct field's "header" tag, or
// the empty string. It does not include the leading comma.
type tagOptions []string

// parseTag splits a struct field's url tag into its name and comma-separated
// options.
func parseTag(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	return s[0], s[1:]
}

// Contains checks whether the tagOptions contains the specified option.
func (o tagOptions) Contains(option string) bool {
	for _, s := range o {
		if s == option {
			return true
		}
	}
	return false
}
