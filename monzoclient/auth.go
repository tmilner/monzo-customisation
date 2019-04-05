package monzoclient

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/url"
)

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	ClientId     string `json:"client_id"`
	Expiry       int32  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	UserId       string `json:"user_id"`
}

func (a *MonzoClient) Authenticate(code string, clientId string, clientSecret string, redirectUri string) (*AuthResponse, error) {
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", clientId)
	form.Add("client_secret", clientSecret)
	form.Add("code", code)
	form.Add("redirect_uri", redirectUri)

	return a.authRequest(form)
}

func (a *MonzoClient) RefreshAuth(auth string, clientId string, clientSecret string) (*AuthResponse, error) {
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("client_id", clientId)
	form.Add("client_secret", clientSecret)
	form.Add("refresh_token", auth)

	return a.authRequest(form)
}

func (a *MonzoClient) authRequest(form map[string][]string) (*AuthResponse, error) {
	res, err := a.client.PostForm("https://api.monzo.com/oauth2/token", form)
	if err != nil {
		log.Printf("Error posting for token %+v", err)
		return nil, err
	}

	if res.Status != "200 OK" {
		log.Printf("auth response is not 200. Is %+v", res.Status)
		return nil, errors.New("response is not 200")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println("Error in auth")
		return nil, err
	}

	var result *AuthResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Unable to unmarshal auth response: %+v", string(body))
		return nil, err
	}

	return result, nil
}
