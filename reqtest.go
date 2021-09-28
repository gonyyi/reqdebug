package reqtest

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

const VERSION = "ReqTest v1.1.0"

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
	ReqNo          int
	ReqList        []string
	Param          string
}

func debugHandler(idx int, param string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		var reqList []string
		currListSize := len(lastDataIndex.List()) // current list size.. it can be 0 to reqsKeep
		for i := 0; i < currListSize; i++ {
			reqList = append(reqList, strconv.Itoa(i))
		}

		// fmt.Printf("ReqList: %v\n", reqList)
		if idx+1 > len(lastData) {
			respTmpl.Execute(w, Data{
				Mode:           "::debug",
				ServiceName:    serviceName,
				ServiceVersion: VERSION,
				Error:          "Index outside the range",
				ReqList:        reqList,
				ReqNo:          idx,
				Param:          param,
			})
			return
		}
		if idx+1 > currListSize || currListSize == 0 {
			respTmpl.Execute(w, Data{
				Mode:           "::debug",
				ServiceName:    serviceName,
				ServiceVersion: VERSION,
				Error:          "No data",
				ReqList:        reqList,
				ReqNo:          idx,
				Param:          param,
			})
			return
		}

		{
			tmpData := lastData[lastDataIndex.Curr()-idx]
			tmpData.Request = strings.TrimSpace(tmpData.Request)
			tmpData.Mode = "::debug"
			tmpData.DebugURL = ""
			tmpData.ReqList = reqList
			tmpData.ReqNo = idx
			tmpData.Param = param
			respTmpl.Execute(w, tmpData)
		}
	}
}

func newHandler(name string, param string) http.HandlerFunc {
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
			lastData[lastDataIndex.Curr()].DebugURL = scheme + "://" + r.Host + r.URL.Path + "?" + param + "=0"
		}

		reqOut, err := httputil.DumpRequest(r, true)
		if err != nil {
			lastData[lastDataIndex.Curr()].Error = err.Error()
		} else {
			lastData[lastDataIndex.Curr()].Error = ""
		}
		lastData[lastDataIndex.Curr()].Request = strings.TrimSpace(string(reqOut))
		if err := respTmpl.Execute(w, lastData[lastDataIndex.Curr()]); err != nil {
			println(err.Error())
		}
	}
}

func Run(addr string, name string, param string, reqsKeep int, ignoreURIs []string, customHandler map[string]http.HandlerFunc) (err error) {
	param = strings.TrimSpace(param)
	if param == "" {
		param = "q"
	}

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
	defaultHandler := newHandler("Default", param)

	router := func(w http.ResponseWriter, r *http.Request) {
		// if URL has a query key "reqtest" with a value, then it's a debug mode.
		// if request URI is in ignoreURLs, do not respond
		if idx, ok := r.URL.Query()[param]; ok {
			intIdx := 0
			if len(idx) > 0 {
				intIdx, _ = strconv.Atoi(idx[0])
			}
			debugHandler(intIdx, param)(w, r)
		} else if _, ok := ignores[r.URL.RequestURI()]; !ok { // don't do anything for ignore URI lists
			// see if customHandler is exist
			url := strings.SplitN(r.Host, ":", 2)[0]
			if h, ok := customHandler[url]; ok && h != nil {
				h(w, r)
			} else {
				defaultHandler(w, r)
			}
		}
	}

	http.HandleFunc("/", router)
	println(VERSION + " <https://gonyyi.com/copyright>")
	println("Starting at <" + addr + ">")
	return http.ListenAndServe(addr, nil)
}
