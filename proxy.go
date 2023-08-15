package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

type MyProxy struct {
	Proxy *httputil.ReverseProxy
}

func (s *MyProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request:", r.URL.Path)

	// Websocket
	if r.Method == "GET" && strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		log.Printf("IceWarp websocket: %s\n", r.URL)
		proxyIcewarp.ServeHTTP(w, r)
		return
	}

	// Login
	if r.URL.Path == "/" && r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading body in login request:", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 - Bad request"))
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // Assign back original body
		if r.PostFormValue("username") != "" && r.PostFormValue("password") != "" {
			log.Printf("Login request: %s\n", r.PostFormValue("username"))
			backend, _ := getBackend(r.PostFormValue("username"))
			log.Printf("Backend: %s\n", backend)
			switch backend {
			case "icewarp":
				log.Println("IceWarp login")
				authtoken, _ := getIceWarpToken(r.PostFormValue("username"), r.PostFormValue("password"), true)
				http.Redirect(w, r, fmt.Sprintf("/webmail/?atoken=%s&language=en", authtoken), http.StatusFound)
				return
			case "other":
				log.Println("Other login")
				r.Body = io.NopCloser(bytes.NewBuffer(body)) // Assign back original body
				proxyOther.ServeHTTP(w, r)
				return
			default:
				log.Println("Unknown backend (serving as other login)")
				r.Body = io.NopCloser(bytes.NewBuffer(body)) // Assign back original body
				proxyOther.ServeHTTP(w, r)
				return
			}
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // Assign back original body
	}

	// Other get requests
	if strings.HasPrefix(r.URL.Path, "/webmail") ||
		strings.HasPrefix(r.URL.Path, "/webdav") ||
		strings.HasPrefix(r.URL.Path, "/icewarpapi") ||
		strings.HasPrefix(r.URL.Path, "/admin") ||
		strings.HasPrefix(r.URL.Path, "/autodiscover") ||
		strings.HasPrefix(r.URL.Path, "/collaboration") ||
		strings.HasPrefix(r.URL.Path, "/conference") ||
		strings.HasPrefix(r.URL.Path, "/downloads") ||
		strings.HasPrefix(r.URL.Path, "/favicon") ||
		strings.HasPrefix(r.URL.Path, "/files") ||
		strings.HasPrefix(r.URL.Path, "/geoip") ||
		strings.HasPrefix(r.URL.Path, "/geoserver") ||
		strings.HasPrefix(r.URL.Path, "/images") ||
		strings.HasPrefix(r.URL.Path, "/teamchat") ||
		strings.HasPrefix(r.URL.Path, "/teamchatapi") ||
		strings.HasPrefix(r.URL.Path, "/wcs") ||
		strings.HasPrefix(r.URL.Path, "/calendar") ||
		strings.HasPrefix(r.URL.Path, "/freebusy") ||
		strings.HasPrefix(r.URL.Path, "/install") ||
		strings.HasPrefix(r.URL.Path, "/ischedule") ||
		strings.HasPrefix(r.URL.Path, "/reports") ||
		strings.HasPrefix(r.URL.Path, "/-.._._.--.._") {
		log.Println("IceWarp request:", r.URL)
		proxyIcewarp.ServeHTTP(w, r)
		return
	}

	log.Println("Other request:", r.URL)
	proxyOther.ServeHTTP(w, r)
	return

}
