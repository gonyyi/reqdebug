// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/01/2021

package reqtest

import "net/http"

// SimpleRun starts only with the address like `:8080`.
// It's debug URI is `/debug` also keeps up to 20 records,
// and ignores request for `favicon.ico`
func SimpleRun(addr string) error {
	rt := New("SimpleRun", "/debug", 20, []string{"/favicon.ico"}) // New creates a testing service
	http.HandleFunc("/debug/", rt.DebugHandler())           // DebugHandler shows all requests
	http.HandleFunc("/", rt.DefaultHandler())               // DefaultHandler returns debug info of the request -eg. for web browser
	return http.ListenAndServe(addr, nil)
}
