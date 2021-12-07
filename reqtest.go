// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/07/2021

package reqtest

import (
	"embed"
	"github.com/gonyyi/gosl"
	"html/template"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

const Version gosl.Ver = "ReqTest v1.3.0"

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
	ViewerURL      string
	ReqNo          int
	ReqList        []string
	TotalReq       int
}

// New creates a request tester -- This panics when parsing handler is failed
// Example:
//   New("X-Test", ":8089", 50, "test.ico")
// When New creates the reqtest, it will kick off the viewer service and it will be serving
// at localhost:8089.
func New(name string, viewerAddr string, reqsKeep int, ignoreURIs ...string) (rt *reqtest) {
	rt = &reqtest{
		serviceName:   name,
		lastData:      make([]data, reqsKeep),
		lastDataIndex: newRollingIndex(reqsKeep),
		ignores:       make(map[string]struct{}),
		viewerAddr:    viewerAddr,
	}

	// ADD IGNORE URIs
	for _, v := range ignoreURIs {
		rt.ignores[v] = struct{}{}
	}

	// PARSE TEMPLATE
	rt.respTmpl, rt.err = template.ParseFS(tmpl, "template.html")

	// CREATE rt.viewerAddrFull
	// If the viewerAddr starts with :, then append localhost
	// Note that rt.viewerAddrFull WAS ":8080", but NOW it's like "http://1.2.3.4:8080"
	if gosl.HasPrefix(rt.viewerAddr, ":") {
		rt.viewerAddrFull = "http://" + getOutboundIP() + rt.viewerAddr
	} else {
		rt.viewerAddrFull = "http://" + rt.viewerAddr
	}

	rt.startViewer()
	return rt
}

type reqtest struct {
	totalRequest   int
	serviceName    string // reqtest service name
	lastData       []data // lastData is to keep most recent x items.
	lastDataIndex  rollingIndex
	ignores        map[string]struct{}
	respTmpl       *template.Template
	viewerAddrFull string // such as "http://1.2.3.4:8080"
	viewerAddr     string // ":8080"
	err            error
}

// Error will return an error
func (rt *reqtest) Error() error {
	return rt.err
}

// startViewer will start the viewer
func (rt *reqtest) startViewer() *reqtest {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rt.ViewHandler())
	go func() {
		// Note that rt.viewerAddrFull here is ":8080", but later will be modified to like "http://1.2.3.4:8080"
		if err := http.ListenAndServe(rt.viewerAddr, mux); err != nil {
			println(err.Error())
		}
	}()
	return rt
}

// ViewerURL returns viewer URL
func (rt *reqtest) ViewerURL() string {
	return rt.viewerAddrFull
}

// DefaultHandler shows a page where request can be view as a response to the request
func (rt *reqtest) DefaultHandler() http.HandlerFunc {
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

// viewHandler traces
func (rt *reqtest) viewHandler(idx int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		var reqList []string
		currListSize := len(rt.lastDataIndex.List()) // current list size.. it can be 0 to reqsKeep
		for i := 0; i < currListSize; i++ {
			reqList = append(reqList, strconv.Itoa(i))
		}

		if gosl.HasPrefix(r.Header.Get("User-Agent"), "curl") {
			if idx+1 > len(rt.lastData) {
				w.WriteHeader(400)
				w.Write([]byte("Index outside the range\n"))
				return
			}
			if idx+1 > currListSize || currListSize == 0 {
				w.WriteHeader(404)
				w.Write([]byte("No data\n"))
				return
			}
			w.WriteHeader(200)
			w.Write([]byte(strings.TrimSpace(rt.lastData[rt.lastDataIndex.Curr()-idx].Request)+"\n"))
			return
		}

		// fmt.Printf("ReqList: %v\n", reqList)
		if idx+1 > len(rt.lastData) {
			rt.respTmpl.Execute(w, data{
				Mode:           "::view",
				ServiceName:    rt.serviceName,
				ServiceVersion: Version.String(),
				Error:          "Index outside the range",
				ReqList:        reqList,
				ReqNo:          idx,
				ViewerURL:      rt.viewerAddrFull,
				TotalReq:       rt.totalRequest,
			})
			return
		}
		if idx+1 > currListSize || currListSize == 0 {
			rt.respTmpl.Execute(w, data{
				Mode:           "::view",
				ServiceName:    rt.serviceName,
				ServiceVersion: Version.String(),
				Error:          "No data",
				ReqList:        reqList,
				ReqNo:          idx,
				ViewerURL:      rt.viewerAddrFull,
				TotalReq:       rt.totalRequest,
			})
			return
		}

		{
			tmpData := rt.lastData[rt.lastDataIndex.Curr()-idx]
			tmpData.Request = strings.TrimSpace(tmpData.Request)
			tmpData.Mode = "::view"
			tmpData.ViewerURL = rt.viewerAddrFull
			tmpData.ReqList = reqList
			tmpData.ReqNo = idx
			tmpData.TotalReq = rt.totalRequest
			rt.respTmpl.Execute(w, tmpData)
		}
	}
}

// ViewHandler is to be used for debug page where requests can be viewed
func (rt *reqtest) ViewHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var idx int
		// Get debug number..
		if r.URL != nil {
			// if URL exist, get the first number
			tmp := gosl.Split(nil, r.URL.Path, '/') // for /debug/123/ => ["123"]
			for _, v := range tmp {
				if v != "" {
					idx = gosl.MustAtoi(v, 0)
					break
				}
			}
		}
		rt.viewHandler(idx)(w, r)
	}
}

// tracer records the trace
func (rt *reqtest) tracer(name string, r *http.Request) (curr int) {
	rt.totalRequest += 1
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
		ViewerURL:      rt.viewerAddrFull,
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
func (rt *reqtest) TraceWrapper(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := rt.ignores[r.URL.RequestURI()]; !ok { // don't do anything for ignore URI lists
			rt.tracer("Custom Handler", r)
			hf(w, r)
		}
	}
}
