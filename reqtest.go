// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/01/2021

package reqtest

import (
	"embed"
	"fmt"
	"github.com/gonyyi/gosl"
	"html/template"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

const Version gosl.Ver = "ReqTest v1.2.0"

var (
	//go:embed template.html
	tmpl embed.FS
)

type data struct {
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
	ReqNo          int
	ReqList        []string
}

// New creates a request tester
// This panics when parsing handler is failed
func New(name string, debugURI string, reqsKeep int, ignoreURIs []string) *reqTest {
	rt := &reqTest{}
	rt.serviceName = name
	rt.lastData = make([]data, reqsKeep)
	rt.lastDataIndex = newRollingIndex(reqsKeep)
	rt.debugURI = debugURI

	// IGNORE SOME URIs
	rt.ignores = make(map[string]struct{})
	for _, v := range ignoreURIs {
		rt.ignores[v] = struct{}{}
	}
	var err error
	if rt.respTmpl, err = template.ParseFS(tmpl, "template.html"); err != nil {
		panic(err)
		return nil
	}
	return rt
}

type reqTest struct {
	serviceName   string
	lastData      []data
	lastDataIndex rollingIndex
	ignores       map[string]struct{}
	debugURI      string
	respTmpl      *template.Template
}

// DefaultHandler shows a page where request can be view as a response to the request
func (rt *reqTest) DefaultHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		curr := 0
		if _, ok := rt.ignores[r.URL.RequestURI()]; !ok { // don't do anything for ignore URI lists
			curr = rt.tracer("Default Handler", r)
		}
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := rt.respTmpl.Execute(w, rt.lastData[curr]); err != nil {
			println(err.Error())
		}
	}
}

// debugHandler traces
func (rt *reqTest) debugHandler(idx int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		var reqList []string
		currListSize := len(rt.lastDataIndex.List()) // current list size.. it can be 0 to reqsKeep
		for i := 0; i < currListSize; i++ {
			reqList = append(reqList, strconv.Itoa(i))
		}

		// fmt.Printf("ReqList: %v\n", reqList)
		if idx+1 > len(rt.lastData) {
			rt.respTmpl.Execute(w, data{
				Mode:           "::debug",
				ServiceName:    rt.serviceName,
				ServiceVersion: Version.String(),
				Error:          "Index outside the range",
				ReqList:        reqList,
				ReqNo:          idx,
				DebugURL:       rt.debugURI,
			})
			return
		}
		if idx+1 > currListSize || currListSize == 0 {
			rt.respTmpl.Execute(w, data{
				Mode:           "::debug",
				ServiceName:    rt.serviceName,
				ServiceVersion: Version.String(),
				Error:          "No data",
				ReqList:        reqList,
				ReqNo:          idx,
				DebugURL:       rt.debugURI,
			})
			return
		}

		{
			tmpData := rt.lastData[rt.lastDataIndex.Curr()-idx]
			tmpData.Request = strings.TrimSpace(tmpData.Request)
			tmpData.Mode = "::debug"
			tmpData.DebugURL = rt.debugURI
			tmpData.ReqList = reqList
			tmpData.ReqNo = idx
			rt.respTmpl.Execute(w, tmpData)
		}
	}
}

// DebugHandler is to be used for debug page where requests can be viewed
func (rt *reqTest) DebugHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var idx int
		// Get debug number..
		if r.URL != nil && strings.HasPrefix(r.URL.Path, rt.debugURI) {
			println(1, r.URL.Path)
			tmp := strings.Split(strings.TrimPrefix(r.URL.Path, rt.debugURI), "/") // /debug/123/ --> ["", "123", ""] or ["123", ""] depend on debugURI is `/debug` or `/debug/`
			for _, v := range tmp {
				if v != "" {
					idx = gosl.MustAtoi(v, 0)
					break
				}
			}
		}
		rt.debugHandler(idx)(w, r)
	}
}

// tracer records the trace
func (rt *reqTest) tracer(name string, r *http.Request) (curr int) {
	rt.lastDataIndex = rt.lastDataIndex.Next()
	curr = rt.lastDataIndex.Curr()
	rt.lastData[curr] = data{
		ServiceName:    rt.serviceName,
		ServiceVersion: Version.String(),
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
		var scheme string
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
		rt.lastData[curr].DebugURL = fmt.Sprintf("%s://%s/%s/%d", scheme, r.Host, rt.debugURI, 0)
		// rt.lastData[curr].DebugURL = scheme + "://" + r.Host + "/" + rt.debugURI + "/"
		// rt.lastData[curr].DebugURL = scheme + "://" + r.Host + r.URL.Path + "?" + param + "=0"
	}
	reqOut, err := httputil.DumpRequest(r, true)
	if err != nil {
		rt.lastData[curr].Error = err.Error()
	} else {
		rt.lastData[curr].Error = ""
	}
	rt.lastData[curr].Request = strings.TrimSpace(string(reqOut))
	return curr
}

// TraceWrapper wraps handler func with a tracer
func (rt *reqTest) TraceWrapper(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.ignores[r.URL.RequestURI()]; !ok { // don't do anything for ignore URI lists
			rt.tracer("Custom Handler", r)
			hf(w, r)
		}
	}
}
