package httpclient

import (
	"encoding/json"
)

type WhoAmIResponse struct {
	Authenticated bool   `json:"authenticated"`
	ClientId      string `json:"client_id"`
	UserId        string `json:"user_id"`
}

func (a *MonzoApi) WhoAmI() (*WhoAmIResponse, error) {
	body, err := a.processGetRequest("/ping/whoami")
	if err != nil {
		return nil, err
	}

	var result WhoAmIResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}
