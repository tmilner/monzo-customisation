package monzoclient

import (
	"encoding/json"
)

type WhoAmIResponse struct {
	Authenticated bool   `json:"authenticated"`
	ClientId      string `json:"client_id"`
	UserId        string `json:"user_id"`
}

func (a *MonzoClient) WhoAmI(authToken string) (*WhoAmIResponse, error) {
	body, err := a.processGetRequest("/ping/whoami", authToken)
	if err != nil {
		return nil, err
	}

	var result WhoAmIResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}
