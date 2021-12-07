// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/07/2021

package main

import (
	"flag"
	"github.com/gonyyi/gosl"
	"github.com/gonyyi/reqtest"
	"net/http"
)

var (
	ViewerAddr  = ":8089"
	ServiceAddr = ":8080"
	Name        = "ReqTest"
	NoReqKeep   = 20
	showVersion = false

	buildDate = "2000-0101-0000"
	buildNo   = "1"
	Version   = gosl.Ver("ReqTest-CMD v1.3.0-" + buildNo)
)

func main() {
	flag.BoolVar(&showVersion, "version", showVersion, "show version")
	flag.StringVar(&ViewerAddr, "v", ViewerAddr, "viewer addr")
	flag.StringVar(&ServiceAddr, "s", ServiceAddr, "service addr")
	flag.StringVar(&Name, "name", Name, "service name")
	flag.IntVar(&NoReqKeep, "keep", NoReqKeep, "number of requests to keep")
	flag.Parse()

	if showVersion {
		println(Version.String() + " (" + buildDate + ")")
		return
	}

	println("Starting " + Version)

	rt := reqtest.New(Name, ViewerAddr, NoReqKeep, "/favicon.ico")
	println("ServiceAddr: <" + ServiceAddr + ">")
	println("ViewerAddr:  <" + rt.ViewerURL() + ">")

	hello := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(Version.String() + ":OK"))
	}

	http.HandleFunc("/", rt.TraceWrapper(hello))
	http.ListenAndServe(ServiceAddr, nil)
}
