package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type (
	tConfig struct {
		External tConfigExternal  `yaml:"external"`
		Backends []tConfigBackend `yaml:"backends"`
	}
	tConfigExternal struct {
		Name    string                        `yaml:"name"`
		Default string                        `yaml:"default"`
		Address string                        `yaml:"address"`
		Http    tConfigExternalHttp           `yaml:"http"`
		Https   tConfigExternalHttps          `yaml:"https"`
		Pop3    tConfigExternalGenericService `yaml:"pop3"`
		Pop3s   tConfigExternalGenericService `yaml:"pop3s"`
		Imap    tConfigExternalGenericService `yaml:"imap"`
		Imaps   tConfigExternalGenericService `yaml:"imaps"`
		Auth    tConfigExternalAuth           `yaml:"auth"`
	}
	tConfigExternalHttp struct {
		Enabled       bool   `yaml:"enabled"`
		Address       string `yaml:"address"`
		Port          int    `yaml:"port"`
		HttpsRedirect bool   `yaml:"https-redirect"`
	}
	tConfigExternalHttps struct {
		Enabled          bool   `yaml:"enabled"`
		Address          string `yaml:"address"`
		Port             int    `yaml:"port"`
		Letsencrypt      bool   `yaml:"letsencrypt"`
		LetsencryptEmail string `yaml:"letsencrypt-email"`
		Cert             string `yaml:"cert"`
		Key              string `yaml:"key"`
	}
	tConfigExternalGenericService struct {
		Enabled bool   `yaml:"enabled"`
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	}
	tConfigExternalAuth struct {
		Type   string `yaml:"type"`
		Script string `yaml:"script"`
	}
	tConfigBackend struct {
		Name      string                        `yaml:"name"`
		Type      string                        `yaml:"type"`
		Address   string                        `yaml:"address"`
		CheckCert bool                          `yaml:"check-cert"`
		Http      tConfigBackendsGenericService `yaml:"http"`
		Pop3      tConfigBackendsGenericService `yaml:"pop3"`
		Imap      tConfigBackendsGenericService `yaml:"imap"`
		proxy     *httputil.ReverseProxy
	}
	tConfigBackendsGenericService struct {
		Enabled    bool `yaml:"enabled"`
		Encryption bool `yaml:"encryption"`
		Port       int  `yaml:"port"`
	}
)

var config *tConfig

func loadConfig() {

	// Initialize config with default values
	config = &tConfig{
		External: tConfigExternal{
			Name:    "",
			Default: "",
			Address: "",
			Http: tConfigExternalHttp{
				Enabled:       false,
				Address:       "",
				Port:          80,
				HttpsRedirect: false,
			},
			Https: tConfigExternalHttps{
				Enabled:          false,
				Address:          "",
				Port:             8443,
				Letsencrypt:      false,
				LetsencryptEmail: "",
				Cert:             "",
				Key:              "",
			},
			Pop3: tConfigExternalGenericService{
				Enabled: false,
				Address: "",
				Port:    110,
			},
			Pop3s: tConfigExternalGenericService{
				Enabled: false,
				Address: "",
				Port:    995,
			},
			Imap: tConfigExternalGenericService{
				Enabled: false,
				Address: "",
				Port:    143,
			},
			Imaps: tConfigExternalGenericService{
				Enabled: false,
				Address: "",
				Port:    993,
			},
			Auth: tConfigExternalAuth{
				Type:   "script",
				Script: "",
			},
		},
	}

	// Load config file
	configFile := []string{"/etc/mailproxy/mailproxy.yml", "/usr/local/etc/mailproxy/mailproxy.yml", "./mailproxy.yml"}
	configFileFound := false
	for _, file := range configFile {
		if _, err := os.Stat(file); err == nil {
			configFileFound = true
			log.Println("[Config] Loading config file", file)
			yamlFile, err := os.ReadFile(file)
			if err != nil {
				log.Printf("[Config] Error reading config file: #%v ", err)
				os.Exit(1)
			}
			err = yaml.Unmarshal(yamlFile, config)
			if err != nil {
				log.Printf("[Config] Error parsing config file: #%v ", err)
				os.Exit(1)
			}
			break
		}
	}

	// Check config
	if !configFileFound {
		log.Printf("[Config] No config file found. %v\n", configFile)
		os.Exit(1)
	}

	if config.External.Name == "" {
		log.Println("[Config] External name is not configured.")
		os.Exit(1)
	}

	if config.External.Default == "" {
		log.Println("[Config] Default backend is not configured.")
		os.Exit(1)
	}

	if config.External.Http.Address == "" {
		config.External.Http.Address = config.External.Address
	}

	if config.External.Https.Address == "" {
		config.External.Https.Address = config.External.Address
	}

	if config.External.Pop3.Address == "" {
		config.External.Pop3.Address = config.External.Address
	}

	if config.External.Pop3s.Address == "" {
		config.External.Pop3s.Address = config.External.Address
	}

	if config.External.Imap.Address == "" {
		config.External.Imap.Address = config.External.Address
	}

	if config.External.Imaps.Address == "" {
		config.External.Imaps.Address = config.External.Address
	}

	if config.External.Auth.Type == "script" {
		if config.External.Auth.Script == "" {
			log.Println("[Config] Auth script is not configured.")
			os.Exit(1)
		}
		if _, err := os.Stat(config.External.Auth.Script); os.IsNotExist(err) {
			log.Println("[Config] Auth script does not exist.")
			os.Exit(1)
		}
	}

	// Check if at least one backend is configured
	if len(config.Backends) == 0 {
		log.Println("[Config] No backends configured.")
		os.Exit(1)
	}

	// Setup backend configuration
	for i := range config.Backends {
		if config.Backends[i].Name == "" {
			log.Println("[Config] Backend name is not configured.")
			os.Exit(1)
		}
		if config.Backends[i].Type == "" {
			log.Println("[Config] Backend type is not configured.")
			os.Exit(1)
		}

		backendUrl := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", config.Backends[i].Address, config.Backends[i].Http.Port),
		}
		if config.Backends[i].Http.Encryption {
			backendUrl.Scheme = "https"
		}
		config.Backends[i].proxy = httputil.NewSingleHostReverseProxy(backendUrl)
		config.Backends[i].proxy.Director = func(r *http.Request) {
			r.URL.Scheme = backendUrl.Scheme
			r.URL.Host = backendUrl.Host
			r.Host = config.External.Name
		}
		config.Backends[i].proxy.Transport = &http.Transport{
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: !config.Backends[i].CheckCert},
			TLSHandshakeTimeout:   5 * time.Second,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		}
	}

}
