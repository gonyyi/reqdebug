package reqdebug

import (
	"embed"
	"html/template"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

const VERSION = "ReqDebug v1.0.0"

var (
	respTmpl    *template.Template
	lastData    Data
	serviceName string
	//go:embed template.html
	tmpl embed.FS
)

type Err string

func (e Err) Error() string {
	return string(e)
}

const (
	ERR_BAD_ADDRESS = Err("unexpected address is given (expected: ':8080')")
)

type Data struct {
	ServiceName    string
	ServiceVersion string
	Mode           string
	Time           string
	Name           string
	Host           string
	URI            string
	IP             string
	Error          string
	Request        string
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	lastData.Mode = "::debug"
	respTmpl.Execute(w, lastData)
}

func newHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		lastData = Data{
			ServiceName:    serviceName,
			ServiceVersion: VERSION,
			Mode: "",
			Time:           time.Now().Format("2006/01/02 15:04:05.000"),
			Name:           name,
			Host:           r.Host,
			URI:            r.URL.String(),
			IP:             r.RemoteAddr,
		}
		reqOut, err := httputil.DumpRequest(r, true)
		if err != nil {
			lastData.Error = err.Error()
		} else {
			lastData.Error = "n/a"
		}
		lastData.Request = string(reqOut)
		respTmpl.Execute(w, lastData)
	}
}

func Run(addr string, name string) (err error) {
	serviceName = name
	lastData.ServiceName = name
	lastData.ServiceVersion = VERSION

	respTmpl, err = template.ParseFS(tmpl, "template.html")
	if err != nil {
		return err
	}

	if !strings.Contains(addr, ":") {
		return ERR_BAD_ADDRESS
	}

	handlers := make(map[string]http.HandlerFunc)
	handlers["int.gonyyi.com"] = newHandler("Internal network")
	handlers["play.gonyyi.com"] = newHandler("Playground")
	handlers["test.play.gonyyi.com"] = newHandler("Playground Test")
	defaultHandler := newHandler("Default")

	router := func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/debug" {
			debugHandler(w, r)
		} else {
			url := strings.SplitN(r.Host, ":", 2)[0]
			if h, ok := handlers[url]; ok && h != nil {
				h(w, r)
			} else {
				defaultHandler(w, r)
			}
		}
	}

	http.HandleFunc("/", router)
	println(VERSION + " / Copyright (c) 2021 Gon Yi <https://gonyyi.com/copyright>")
	println("Starting at <" + addr + ">")
	return http.ListenAndServe(addr, nil)
}
