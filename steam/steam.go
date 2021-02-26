package steam

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type PlayerSummariesResponseWrapper struct {
	Response PlayerSummariesResponse `json:"response"`
}

type PlayerSummariesResponse struct {
	Players []PlayerSummary `json:"players"`
}

type PlayerSummary struct {
	PersonaName string `json:"personaname"`
}

type API struct {
	Key string
}

func (api *API) GetPlayerName(id string) (string, error) {
	endpoint := "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/"

	url := fmt.Sprintf("%s?key=%s&steamids=%s", endpoint, api.Key, id)
	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting response from steam player api: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response from steam player api: %w", err)
	}

	obj := &PlayerSummariesResponseWrapper{}
	err = json.Unmarshal(buf.Bytes(), obj)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response from steam player api: %w", err)
	}

	if len(obj.Response.Players) != 0 {
		return obj.Response.Players[0].PersonaName, nil
	}

	return "", errors.New("response from steam player api contained no players")
}