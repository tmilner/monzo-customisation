package monzorestclient

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type MonzoRestClient struct {
	url    string
	client *http.Client
}

func CreateMonzoRestClient(url string, client *http.Client) *MonzoRestClient {
	return &MonzoRestClient{url: url, client: client}
}

func (a *MonzoRestClient) processGetRequest(path string, authToken string) ([]byte, error) {
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

func (a *MonzoRestClient) processPatchRequest(path string, authToken string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest("PATCH", a.url+path, body)

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
