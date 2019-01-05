package httpheader

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var decoderType = reflect.TypeOf(new(Decoder)).Elem()

// Decoder is an interface implemented by any type that wishes to decode
// itself from Header fields in a non-standard way.
type Decoder interface {
	ParseHeader(header http.Header, v interface{}) error
}

// ParseHeader expects to be passed an http.Header and a struct, and parses
// header into the struct recursively using the same rules as Header (see above)
func ParseHeader(header http.Header, v interface{}) error {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
		fmt.Println(val.Kind())
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("val is not a struct %+v", val.Kind())
	}
	return parseValue(header, val)
}

func parseValue(header http.Header, val reflect.Value) error {
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
				continue
			}
			name = sf.Name
		}

		if opts.Contains("omitempty") && header.Get(name) == "" {
			continue
		}

		if sv.Type().Implements(decoderType) {
			if !reflect.Indirect(sv).IsValid() {
				sv = reflect.New(sv.Type().Elem())
			}

			m := sv.Interface().(Decoder)
			if err := m.ParseHeader(header, &val); err != nil {
				return err
			}
			continue
		}

		// TODO: implement iterating over multiple Headers with the same name such as `Cooke:` or `Server:`
		// if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
		// 	for k, v := range header[name] {

		// 	}
		// 	continue
		// }

		for sv.Kind() == reflect.Ptr {
			sv = sv.Elem()
		}

		if sv.Type() == timeType {
			h := header.Get(name)
			t, err := time.Parse(time.RFC1123, h)
			if err != nil {
				return err
			}
			sv.Set(reflect.ValueOf(t))
			continue
		}

		if sv.Kind() == reflect.Struct {
			parseValue(header, sv)
			continue
		}

		if sv.Kind() == reflect.Int {
			j, err := strconv.Atoi(header.Get(name))
			if err != nil {
				return err
			}
			sv.SetInt(int64(j))
		}

		if sv.Kind() == reflect.String {
			sv.SetString(header.Get(name))
		}
	}
	return nil
}
