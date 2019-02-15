package httpclient

import (
	"encoding/json"
	"github.com/tmilner/monzo-customisation/configuration"
	"io/ioutil"
	"log"
	"net/http"
)

type AccountListResponse struct {
	Accounts []AccountResponse
}
type AccountResponse struct {
	Id            string           `json:"id"`
	Closed        bool             `json:"closed"`
	Description   string           `json:"description"`
	Created       string           `json:"created"`
	Type          string           `json:"type"`
	AccountNumber string           `json:"account_number,omitempty"`
	SortCode      string           `json:"sort_code,omitempty"`
	Owners        []OwnersResponse `json:"owners"`
}
type OwnersResponse struct {
	UserId             string `json:"user_id"`
	PreferredName      string `json:"preferred_name"`
	PreferredFirstName string `json:"preferred_first_name"`
}

func ListAccounts(client *http.Client, config *configuration.Configuration) (*AccountListResponse, error) {
	req, err := http.NewRequest("GET", monzoapi+"/accounts", nil)
	req.Header.Add("Authorization", "Bearer "+config.Authorization)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	var result AccountListResponse

	err = json.Unmarshal(body, &result)

	return &result, err
}
