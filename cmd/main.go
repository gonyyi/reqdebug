// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/01/2021

package main

import (
	"github.com/gonyyi/reqtest"
	"net/http"
)

func main() {
	reqtest.SimpleRun(":8080")
}

func main1() {
	rt := reqtest.New("t1", "/debug", 10, []string{"/favicon.ico"}) // New creates a testing service
	hello := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("hello world test")) }
	http.HandleFunc("/test/", rt.TraceWrapper(hello)) // TraceWrapper takes a handler function and run trace
	http.HandleFunc("/debug/", rt.DebugHandler())     // DebugHandler shows all requests
	http.HandleFunc("/", rt.DefaultHandler())         // DefaultHandler returns debug info of the request -eg. for web browser
	http.ListenAndServe(":8080", nil)
}
