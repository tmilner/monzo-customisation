package httpclient

import (
	"encoding/json"
	"github.com/tmilner/monzo-customisation/configuration"
	"io/ioutil"
	"net/http"
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

func GetPots(client *http.Client, config *configuration.Configuration) (*PotsResponse, error) {
	req, err := http.NewRequest("GET", monzoapi+"/pots", nil)
	req.Header.Add("Authorization", "Bearer "+config.Authorization)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result PotsResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}
