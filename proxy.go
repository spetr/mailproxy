package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type MyProxy struct {
	Proxy *httputil.ReverseProxy
}

func (s *MyProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Websocket to "/" is handled with first backend of icewarp type
	if r.Method == "GET" && r.URL.Path == "/" && strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		for _, backend := range config.Backends {
			if backend.Type == "icewarp" {
				log.Printf("[IceWarp Websocket] %s %s\n", r.RemoteAddr, r.URL)
				backend.proxy.ServeHTTP(w, r)
				return
			}
		}
	}

	// Typical icewarp URLs are handled with first backend of icewarp type
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
		for _, backend := range config.Backends {
			if backend.Type == "icewarp" {
				log.Printf("[IceWarp URL] %s %s\n", r.RemoteAddr, r.URL)
				backend.proxy.ServeHTTP(w, r)
				return
			}
		}
	}

	// Zimbra login
	if r.URL.Path == "/" && r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("[Zimbra login] Error reading body in login request:", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 - Bad request"))
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // Assign back original body
		if r.PostFormValue("username") != "" && r.PostFormValue("password") != "" {
			backend, _ := getBackend(r.PostFormValue("username"))
			log.Printf("[Zimbra loginform] %s %s %s\n", r.RemoteAddr, r.URL, r.PostFormValue("username"))
			switch backend {
			case "icewarp": // TODO
				authtoken, _ := getIceWarpToken(r.PostFormValue("username"), r.PostFormValue("password"), true)
				authtoken = url.QueryEscape(authtoken)
				log.Printf("[IceWarp token] %s %s\n", r.PostFormValue("username"), authtoken)
				http.Redirect(w, r, fmt.Sprintf("/webmail/?atoken=%s&language=en", authtoken), http.StatusFound)
				return
			case "other": // TODO
				r.Body = io.NopCloser(bytes.NewBuffer(body)) // Assign back original body
				// Handle with first backend of not icewarp type
				for _, backend := range config.Backends {
					if backend.Type != "icewarp" { // TODO
						log.Printf("[Zimbra login] %s %s %s\n", r.RemoteAddr, r.URL, r.PostFormValue("username"))
						backend.proxy.ServeHTTP(w, r)
						return
					}
				}
				return
			default:
				log.Println("Unknown backend (serving as other login)")
				r.Body = io.NopCloser(bytes.NewBuffer(body)) // Assign back original body
				// Handle with default backend
				for _, backend := range config.Backends {
					if backend.Name == config.External.Default {
						log.Println("Login handled with default backend")
						backend.proxy.ServeHTTP(w, r)
						return
					}
				}
				return
			}
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // Assign back original body
	}

	// Handle with default backend
	for i := range config.Backends {
		if config.Backends[i].Name == config.External.Default {
			log.Printf("Request handled with default backend: %s\n", r.URL)
			config.Backends[i].proxy.ServeHTTP(w, r)
			return
		}
	}

	log.Printf("Unhandled request: %s\n", r.URL)
}
