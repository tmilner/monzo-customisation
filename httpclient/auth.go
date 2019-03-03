package httpclient

import (
	"encoding/json"
	"errors"
	"github.com/twinj/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var state = uuid.NewV4().String()
var ticker = time.NewTicker(10 * time.Minute)

func extendAuth(api *MonzoApi) {
	for {
		select {
		case <-ticker.C:
			api.RefreshAuth()
		}
	}
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	ClientId     string `json:"client_id"`
	Expiry       int32  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	UserId       string `json:"user_id"`
}

func (a *MonzoApi) AuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Auth Reuqest recieved: %+v", r)
	uri := "https://auth.monzo.com/?client_id=" + a.ClientConfig.ClientId + "&redirect_uri=" + a.ClientConfig.RedirectUri + "&response_type=code&state=" + state

	http.Redirect(w, r, uri, 303)
}

func (a *MonzoApi) AuthReturnHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Auth_Return received!")
	code := r.URL.Query().Get("code")
	stateReturned := r.URL.Query().Get("state")

	log.Printf("Got code: %s", code)

	if stateReturned != state {
		log.Println("State token is not correct!")
		_, _ = io.WriteString(w, "Fuck Off.")
		return
	}

	client := &http.Client{}

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", a.ClientConfig.ClientId)
	form.Add("client_secret", a.ClientConfig.ClientSecret)
	form.Add("code", code)
	form.Add("redirect_uri", a.ClientConfig.RedirectUri)

	res, err := client.PostForm("https://api.monzo.com/oauth2/token", form)
	if err != nil {
		log.Println("Error posting for token")
		_, _ = io.WriteString(w, "Error")
		return
	}

	if res.Status != "200 OK" {
		log.Printf("Auth response is not 200. Is %+v", res.Status)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println("Error in auth")
		return
	}

	var result *AuthResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Unable to unmarshal auth response: %+v", string(body))
		return
	}

	a.Auth = result
	go a.runBasicInfo()
	go extendAuth(a)

	_, _ = io.WriteString(w, "Suck it.")
}

func (a *MonzoApi) RefreshAuth() error {
	log.Println("Refreshing Auth")

	client := &http.Client{}

	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("client_id", a.ClientConfig.ClientId)
	form.Add("client_secret", a.ClientConfig.ClientSecret)
	form.Add("refresh_token", a.Auth.RefreshToken)

	res, err := client.PostForm("https://api.monzo.com/oauth2/token", form)
	if err != nil {
		log.Printf("Error posting for token %+v", err)
		return err
	}

	if res.Status != "200 OK" {
		log.Printf("Auth response is not 200. Is %+v", res.Status)
		return errors.New("response is not 200")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println("Error in auth")
		return err
	}

	var result *AuthResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("Unable to unmarshal auth response: %+v", string(body))
		return err
	}

	a.Auth = result
	return nil
}
