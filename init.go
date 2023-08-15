package main

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func init() {
	var err error
	remoteIcewarpParsed, err = url.Parse(remoteIcewarp)
	if err != nil {
		panic(err)
	}

	remoteOtherParsed, err = url.Parse(remoteOther)
	if err != nil {
		panic(err)
	}

	proxyIcewarp = httputil.NewSingleHostReverseProxy(remoteIcewarpParsed)
	proxyOther = httputil.NewSingleHostReverseProxy(remoteOtherParsed)

	proxyIcewarp.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxyOther.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	proxyIcewarp.Director = func(r *http.Request) {
		r.URL.Scheme = remoteIcewarpParsed.Scheme
		r.URL.Host = remoteIcewarpParsed.Host
		r.Host = externalAddress
	}
	proxyOther.Director = func(r *http.Request) {
		r.URL.Scheme = remoteOtherParsed.Scheme
		r.URL.Host = remoteOtherParsed.Host
		r.Host = externalAddress
	}
}
