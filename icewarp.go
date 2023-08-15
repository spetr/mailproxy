package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

	apiURL, _ := url.JoinPath(remoteIcewarpParsed.String(), "icewarpapi/")

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte(requestXml)))
	req.Header.Set("Content-Type", "text/xml")

	client := &http.Client{
		Transport: proxyIcewarp.Transport,
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
	log.Printf("IceWarp authtoken recieved: %s\n", IceWarpAPIResponse.Query.Result.AuthToken)

	return IceWarpAPIResponse.Query.Result.AuthToken, nil
}
