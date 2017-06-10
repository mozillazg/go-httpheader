package httpheader_test

import (
	"bitbucket.org/mozillazg/go-httpheader"
	"fmt"
)

func ExampleHeader() {
	type Options struct {
		ContentType  string   `header:"Content-Type"`
		Length       int      `header:"Length"`
		XArray       []string `header:"X-Array"`
		TestHide     string   `header:"-"`
		IgnoreEmpty  string   `header:"X-Empty,omitempty"`
		IgnoreEmptyN string   `header:"X-Empty-N,omitempty"`
	}

	opt := Options{
		ContentType:  "application/json",
		Length:       2,
		XArray:       []string{"test1", "test2"},
		TestHide:     "hide",
		IgnoreEmptyN: "n",
	}
	h, _ := httpheader.Header(opt)
	fmt.Println(h["Content-Type"])
	fmt.Println(h["Length"])
	fmt.Println(h["X-Array"])
	_, ok := h["TestHide"]
	fmt.Println(ok)
	_, ok = h["X-Empty"]
	fmt.Println(ok)
	fmt.Println(h["X-Empty-N"])
	// Output:
	// [application/json]
	// [2]
	// [test1 test2]
	// false
	// false
	// [n]
}
