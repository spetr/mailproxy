package main

import (
	"log"
	"net/http"
)

var (
//remoteIcewarp = "https://192.168.94.225"
//remoteOther   = "https://192.168.94.226"

//remoteIcewarpParsed *url.URL
//remoteOtherParsed   *url.URL

// proxyIcewarp *httputil.ReverseProxy
// proxyOther   *httputil.ReverseProxy
)

func main() {

	loadConfig()

	log.Println("[HTTP] Starting proxy server")

	proxy := &MyProxy{}
	http.Handle("/", proxy)
	server := &http.Server{
		Addr:    ":8080",
		Handler: proxy,
	}
	log.Fatal(server.ListenAndServe())

}
