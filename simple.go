// (c) Gon Y. Yi 2021 <https://gonyyi.com/copyright>
// Last Update: 12/07/2021

package reqtest

import "net/http"

// SimpleRun starts only with the address like `:8080`.
// It's debug URI is `/debug` also keeps up to 20 records,
// and ignores request for `favicon.ico`
func SimpleRun(serviceAddr, viewerAddr string) error {
	// New creates a testing service
	rt := New("ReqTest", viewerAddr, 20, "/favicon.ico")
	http.HandleFunc("/", rt.DefaultHandler()) // DefaultHandler returns debug info of the request -eg. for web browser
	println("ServiceAddr: <" + serviceAddr + ">")
	println("ViewerAddr:  <" + rt.ViewerURL() + ">")
	return http.ListenAndServe(serviceAddr, nil)
}
