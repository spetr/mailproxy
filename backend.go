package main

import (
	"bytes"
	"log"
	"os/exec"
)

func getBackend(username string) (string, error) {
	var (
		err    error
		outBuf bytes.Buffer
	)
	cmd := exec.Command("/usr/share/nginx/html/getserver.sh", username)
	cmd.Stdout = &outBuf
	if err = cmd.Start(); err != nil {
		log.Printf("Get backend error: %s\n", err.Error())
		return "", err
	}
	err = cmd.Wait()
	if err == nil {
		if outBuf.String() == "200" {
			return "icewarp", nil
		} else {
			return "other", nil
		}
	}
	return "", err
}
