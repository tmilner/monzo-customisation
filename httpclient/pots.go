package httpclient

import (
	"encoding/json"
	"time"
)

type PotsResponse struct {
	Pots []PotResponse
}
type PotResponse struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Style    string    `json:"style"`
	Balance  int64     `json:"balance"`
	Currency string    `json:"currency"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
	Deleted  bool      `json:"deleted"`
}

func (a *MonzoApi) GetPots() (*PotsResponse, error) {
	body, err := a.processGetRequest("/pots")
	if err != nil {
		return nil, err
	}

	var result *PotsResponse
	err = json.Unmarshal(body, &result)

	return result, err
}
