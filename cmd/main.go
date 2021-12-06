// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/06/2021

package main

import (
	"github.com/gonyyi/gosl"
	"github.com/gonyyi/reqtest"
	"os"
)

func main() {
	println(reqtest.Version + " " + "<https://github.com/gonyyi/reqtest>")

	if len(os.Args) < 2 {
		println("Usage: "+os.Args[0]+" <PORT>")
		return
	}
	// if not number, then report an error
	if !gosl.IsNumber(os.Args[1]) {
		println("Error: <PORT> must be an integer (1-65535)")
		return
	}
	if err := reqtest.SimpleRun(":" + os.Args[1]); err != nil {
		println(err.Error())
	}
}

func Example() {
	rt := reqtest.New("t1", "/debug", 10, []string{"/favicon.ico"}) // New creates a testing service
	hello := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("hello world test")) }
	http.HandleFunc("/test/", rt.TraceWrapper(hello)) // TraceWrapper takes a handler function and run trace
	http.HandleFunc("/debug/", rt.DebugHandler())     // DebugHandler shows all requests
	http.HandleFunc("/", rt.DefaultHandler())         // DefaultHandler returns debug info of the request -eg. for web browser
	http.ListenAndServe(":8080", nil)
}
