// (c) Gon Y. Yi 2021-2022 <https://gonyyi.com/copyright>
// Last Update: 01/15/2022

package main

import (
	"flag"
	"github.com/gonyyi/reqtest"
	"net/http"
)

var (
	ViewerAddr  = ":8080"
	ServiceAddr = ":80"
	Name        = "gonyyi.int"
	NoReqKeep   = 20
	showVersion = false

	buildDate = "2000-0101-0000"
	buildNo   = "1"
	Version   = "gonyyi.int v1.3.1-" + buildNo
)

func main() {
	flag.BoolVar(&showVersion, "version", showVersion, "show version")
	flag.StringVar(&ViewerAddr, "v", ViewerAddr, "viewer addr")
	flag.StringVar(&ServiceAddr, "s", ServiceAddr, "service addr")
	flag.StringVar(&Name, "name", Name, "service name")
	flag.IntVar(&NoReqKeep, "keep", NoReqKeep, "number of requests to keep")
	flag.Parse()

	if showVersion {
		println(Version + " (" + buildDate + ")")
		return
	}

	println("Starting " + Version)

	rt := reqtest.New(Name, ViewerAddr, NoReqKeep, "/favicon.ico")
	println("ServiceAddr: <" + ServiceAddr + ">")
	println("ViewerAddr:  <" + rt.ViewerURL() + ">")

	http.HandleFunc("/", rt.TraceWrapper(helloHandler(nil)))

	http.ListenAndServe(ServiceAddr, nil)
}
