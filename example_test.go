package httpheader_test

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/mozillazg/go-httpheader"
)

func ExampleHeader() {
	type Options struct {
		ContentType  string `header:"Content-Type"`
		Length       int
		Bool         bool
		BoolInt      bool      `header:"Bool-Int,int"`
		XArray       []string  `header:"X-Array"`
		TestHide     string    `header:"-"`
		IgnoreEmpty  string    `header:"X-Empty,omitempty"`
		IgnoreEmptyN string    `header:"X-Empty-N,omitempty"`
		CreatedAt    time.Time `header:"Created-At"`
		UpdatedAt    time.Time `header:"Update-At,unix"`
		CustomHeader http.Header
	}

	opt := Options{
		ContentType:  "application/json",
		Length:       2,
		Bool:         true,
		BoolInt:      true,
		XArray:       []string{"test1", "test2"},
		TestHide:     "hide",
		IgnoreEmptyN: "n",
		CreatedAt:    time.Date(2000, 1, 1, 12, 34, 56, 0, time.UTC),
		UpdatedAt:    time.Date(2001, 1, 1, 12, 34, 56, 0, time.UTC),
		CustomHeader: http.Header{
			"X-Test-1": []string{"233"},
			"X-Test-2": []string{"666"},
		},
	}
	h, _ := httpheader.Header(opt)
	var keys []string
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %#v\n", k, h.Values(k))
	}
	// Output:
	// Bool: []string{"true"}
	// Bool-Int: []string{"1"}
	// Content-Type: []string{"application/json"}
	// Created-At: []string{"Sat, 01 Jan 2000 12:34:56 GMT"}
	// Length: []string{"2"}
	// Update-At: []string{"978352496"}
	// X-Array: []string{"test1", "test2"}
	// X-Empty-N: []string{"n"}
	// X-Test-1: []string{"233"}
	// X-Test-2: []string{"666"}
}
