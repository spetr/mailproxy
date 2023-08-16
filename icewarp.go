package main

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const iceWarpLoginXML = `
<iq format="text/xml">
	<query xmlns="admin:iq:rpc" >
		<commandname>getauthtoken</commandname>
		<commandparams>
			<email>%s</email>
			<authtype>0</authtype>
			<persistentlogin>%d</persistentlogin>
			<password>%s</password>
		</commandparams>
	</query>
</iq>`

type (
	tIceWarpAPIResponse struct {
		XMLName xml.Name `xml:"iq"`
		Query   struct {
			XMLName xml.Name `xml:"query"`
			Result  struct {
				XMLName   xml.Name `xml:"result"`
				AuthToken string   `xml:"authtoken"`
			} `xml:"result"`
		} `xml:"query"`
	}
)

func getIceWarpToken(username string, password string, persistent bool) (string, error) {
	var err error
	persistentInt := 0
	if persistent {
		persistentInt = 1
	}
	requestXml := fmt.Sprintf(iceWarpLoginXML, username, persistentInt, password)

	apiURL := ""
	for i := range config.Backends {
		if config.Backends[i].Type == "icewarp" {
			apiURL = fmt.Sprintf("https://%s:%d", config.Backends[i].Address, config.Backends[i].Http.Port)
			break
		}
	}
	apiURL, _ = url.JoinPath(apiURL, "icewarpapi/")

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte(requestXml)))
	req.Header.Set("Content-Type", "text/xml")

	// TODO use transport from proxy
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			TLSHandshakeTimeout:   5 * time.Second,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("IceWarp API error: %s\n", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("IceWarp API error: %s\n", err.Error())
		return "", err
	}
	IceWarpAPIResponse := tIceWarpAPIResponse{}
	err = xml.Unmarshal(body, &IceWarpAPIResponse)
	if err != nil {
		log.Printf("IceWarp API error (XML parser): %s\n", err.Error())
		return "", err
	}

	return IceWarpAPIResponse.Query.Result.AuthToken, nil
}
