package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	remoteIcewarp   = "https://192.168.94.225"
	remoteOther     = "https://192.168.94.226"
	externalAddress = "192.168.94.204"

	remoteIcewarpParsed *url.URL
	remoteOtherParsed   *url.URL

	proxyIcewarp *httputil.ReverseProxy
	proxyOther   *httputil.ReverseProxy
)

func main() {

	proxy := &MyProxy{}
	http.Handle("/", proxy)
	server := &http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}
	log.Fatal(server.ListenAndServe())

}
