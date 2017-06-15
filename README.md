# go-cors-gate

[![Build Status](https://travis-ci.org/roemerb/go-cors-gate.svg?branch=master)](https://travis-ci.org/roemerb/go-cors-gate)

Server side CORS validation middleware. It's a Golang version of [mixmaxhq's cors-gate package for Node](https://github.com/mixmaxhq).

### Usage

This middleware can be used as follows:

```
package main

import (
	"net/http"
	
	"github.com/roemerb/corsgate"
)

var myHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
})

func main() {
	gate := corsgate.New(corsgate.Options{
		Origin: 	[]string{"example.com"},
		Strict: 	true,
		Failure: 	func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
	})
	
	app := gate.Handler(myHandler)
	http.ListenAndServe("127.0.0.1:8080", app)
}
```
