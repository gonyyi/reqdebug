// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/07/2021

package main

import (
	"github.com/gonyyi/reqtest"
	"net/http"
)

func main() {
	normal()
}

func simple() {
	println("Starting " + reqtest.Version)
	if err := reqtest.SimpleRun(":8080", ":8089"); err != nil {
		println(err.Error())
	}
}

func normal() {
	rt := reqtest.New("t1", ":8089", 10, "/favicon.ico")
	hello := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("hello world test")) }

	println("ServiceAddr: <" + ":8080" + ">")
	println("ViewerAddr:  <" + rt.ViewerURL() + ">")

	http.HandleFunc("/test/", rt.TraceWrapper(hello)) // TraceWrapper takes a handler function and run trace
	http.HandleFunc("/debug/", rt.ViewHandler())      // ViewHandler shows all requests
	http.HandleFunc("/", rt.DefaultHandler())         // DefaultHandler returns debug info of the request -eg. for web browser
	http.ListenAndServe(":8080", nil)
}
