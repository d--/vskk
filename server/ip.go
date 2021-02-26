package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type PublicIPResponse struct {
	Ip string `json:"ip"`
}

func GetPublicIP() (string, error) {
	response, err := http.Get("https://api.ipify.org/?format=json")
	if err != nil {
		return "", fmt.Errorf("ip get failed: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		return "", fmt.Errorf("ip response body read failed: %w", err)
	}

	obj := &PublicIPResponse{}
	err = json.Unmarshal(buf.Bytes(), obj)
	if err != nil {
		return "", fmt.Errorf("ip response body unmarhsal failed: %w", err)
	}

	return obj.Ip, nil
}