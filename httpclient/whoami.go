package httpclient

import (
	"encoding/json"
	"github.com/tmilner/monzo-customisation/configuration"
	"io/ioutil"
	"log"
	"net/http"
)

type WhoAmIResponse struct {
	Authenticated bool   `json:"authenticated"`
	ClientId      string `json:"client_id"`
	UserId        string `json:"user_id"`
}

func WhoAmI(client *http.Client, config *configuration.Configuration) (*WhoAmIResponse, error) {
	req, err := http.NewRequest("GET", monzoapi+"/ping/whoami", nil)
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

	var result WhoAmIResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}
