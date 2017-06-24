# go-httpheader

go-httpheader is a Go library for encoding structs into Header fields.

[![Build Status](https://circleci.com/bb/mozillazg/go-httpheader.svg?style=svg)](https://circleci.com/bb/mozillazg/go-httpheader)
[![Go Report Card](https://goreportcard.com/badge/bitbucket.org/mozillazg/go-httpheader)](https://goreportcard.com/report/bitbucket.org/mozillazg/go-httpheader)
[![GoDoc](https://godoc.org/bitbucket.org/mozillazg/go-httpheader?status.svg)](https://godoc.org/bitbucket.org/mozillazg/go-httpheader)

## install

`go get -u bitbucket.org/mozillazg/go-httpheader`


## usage

```go
package main

import (
	"fmt"
	"net/http"

	"bitbucket.org/mozillazg/go-httpheader"
)

func main() {
	type Options struct {
		hide         string
		ContentType  string `header:"Content-Type"`
		Length       int
		XArray       []string `header:"X-Array"`
		TestHide     string   `header:"-"`
		IgnoreEmpty  string   `header:"X-Empty,omitempty"`
		IgnoreEmptyN string   `header:"X-Empty-N,omitempty"`
		CustomHeader http.Header
	}

	opt := Options{
		hide:         "hide",
		ContentType:  "application/json",
		Length:       2,
		XArray:       []string{"test1", "test2"},
		TestHide:     "hide",
		IgnoreEmptyN: "n",
		CustomHeader: http.Header{
			"X-Test-1": []string{"233"},
			"X-Test-2": []string{"666"},
		},
	}
	h, _ := httpheader.Header(opt)
	fmt.Printf("%#v", h)
}
```
