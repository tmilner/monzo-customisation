package monzoclient

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type MonzoClient struct {
	url    string
	client *http.Client
}

func CreateMonzoClient(url string, client *http.Client) *MonzoClient {
	return &MonzoClient{url: url, client: client}
}

func (a *MonzoClient) processGetRequest(path string, authToken string) ([]byte, error) {
	req, err := http.NewRequest("GET", a.url+path, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+authToken)
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.Status != "200 OK" {
		log.Printf("Result is not Sucess. Its actually: %s", resp.Status)
		return nil, errors.New("not 200")
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
