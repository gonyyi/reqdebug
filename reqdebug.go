package reqdebug

import (
	"embed"
	"github.com/gonyyi/common"
	"html/template"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

const VERSION = "ReqDebug v1.1.0"

var (
	respTmpl      *template.Template
	lastData      []Data
	lastDataIndex common.RollingIndex
	serviceName   string
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
	Path           string
	URI            string
	IP             string
	Error          string
	Request        string
	DebugURL       string
}

func debugHandler(idx int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		currListSize := len(lastDataIndex.List()) // current list size.. it can be 0 to reqsKeep
		if idx+1 > len(lastData) {
			respTmpl.Execute(w, Data{
				Mode: "::debug",
				ServiceName:    serviceName,
				ServiceVersion: VERSION,
				Error: "Index outside the range",
			})
			return
		}
		if idx+1 > currListSize || currListSize == 0 {
			respTmpl.Execute(w, Data{
				Mode: "::debug",
				ServiceName:    serviceName,
				ServiceVersion: VERSION,
				Error: "No data",
			})
			return
		}

		{
			tmpData := lastData[lastDataIndex.Curr()-idx]
			tmpData.Mode = "::debug"
			tmpData.DebugURL = ""
			respTmpl.Execute(w, tmpData)
		}
	}
}

func newHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		lastDataIndex = lastDataIndex.Next()
		lastData[lastDataIndex.Curr()] = Data{
			ServiceName:    serviceName,
			ServiceVersion: VERSION,
			Mode:           "",
			Time:           time.Now().Format("2006/01/02 15:04:05.000"),
			Name:           name,
			Host:           r.Host,
			Path:           r.URL.Path,
			URI:            r.RequestURI,
			IP:             r.RemoteAddr,
			DebugURL:       "",
		}

		{
			scheme := "https"
			if r.TLS == nil {
				scheme = "http"
			}
			lastData[lastDataIndex.Curr()].DebugURL = scheme + "://" + r.Host + r.URL.Path + "?reqdebug=0"
		}

		reqOut, err := httputil.DumpRequest(r, true)
		if err != nil {
			lastData[lastDataIndex.Curr()].Error = err.Error()
		} else {
			lastData[lastDataIndex.Curr()].Error = ""
		}
		lastData[lastDataIndex.Curr()].Request = string(reqOut)
		if err := respTmpl.Execute(w, lastData[lastDataIndex.Curr()]); err != nil {
			println(err.Error())
		}
	}
}

func Run(addr string, name string, reqsKeep int, ignoreURIs []string, customHandler map[string]http.HandlerFunc) (err error) {
	serviceName = name
	lastData = make([]Data, reqsKeep)
	lastDataIndex = common.NewRollingIndex(reqsKeep)


	// IGNORE SOME URIs
	ignores := make(map[string]struct{})
	for _, v := range ignoreURIs {
		ignores[v] = struct{}{}
	}

	// LOAD TEMPLATE
	respTmpl, err = template.ParseFS(tmpl, "template.html")
	if err != nil {
		return err
	}
	if !strings.Contains(addr, ":") {
		return ERR_BAD_ADDRESS
	}

	// HANDLERS
	if customHandler == nil {
		customHandler = make(map[string]http.HandlerFunc)
	}
	defaultHandler := newHandler("Default")

	router := func(w http.ResponseWriter, r *http.Request) {
		if idx := r.URL.Query().Get("reqdebug"); idx != "" {
			intIdx, _ := strconv.Atoi(idx)

			debugHandler(intIdx)(w,r)

			// if dstIdx := (len(idxList) - 1) - intIdx; dstIdx > -1 {
			// 	debugHandler(idxList[dstIdx])(w, r)
			// } else {
			// 	debugHandler(-1)(w, r)
			// }

		} else if _, ok := ignores[r.URL.RequestURI()]; !ok {
			// don't do anything for ignore URI lists
			url := strings.SplitN(r.Host, ":", 2)[0]
			if h, ok := customHandler[url]; ok && h != nil {
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
