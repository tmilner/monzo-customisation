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
	"sync"
	"time"
)

var state = uuid.NewV4().String()
var ticker = time.NewTicker(2 * time.Hour)

func extendAuth(api *MonzoApi) {
	for {
		select {
		case <-ticker.C:
			_ = api.RefreshAuth()
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
	log.Printf("auth Reuqest recieved: %+v", r)
	uri := "https://auth.monzo.com/?client_id=" + a.clientConfig.ClientId + "&redirect_uri=" + a.clientConfig.RedirectUri + "&response_type=code&state=" + state

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
	form.Add("client_id", a.clientConfig.ClientId)
	form.Add("client_secret", a.clientConfig.ClientSecret)
	form.Add("code", code)
	form.Add("redirect_uri", a.clientConfig.RedirectUri)

	res, err := client.PostForm("https://api.monzo.com/oauth2/token", form)
	if err != nil {
		log.Printf("Error posting for token! Error: %+v", err)
		_, _ = io.WriteString(w, "Something is wrong")
		return
	}

	if res.Status != "200 OK" {
		log.Printf("auth response is not 200. Is %+v", res.Status)
		return
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

	err = a.saveUserAndAccounts(result, false)
	if err != nil {
		return
	}

	go a.runBasicInfo(result.UserId)

	_, _ = io.WriteString(w, "Suck it.")
}

func (a *MonzoApi) saveUserAndAccounts(response *AuthResponse, usersLocked bool) error {
	accountRes, err := a.ListAccounts(response.UserId)
	if err != nil {
		log.Printf("Failed to get account info for authorised account %+v", err)
		return errors.New("failed to get account info")
	}
	user := &User{
		id:       response.UserId,
		auth:     response,
		accounts: make([]*Account, 0),
	}

	for _, acc := range accountRes.Accounts {
		account := &Account{
			id: acc.Id,
			processedTransactions: sync.Map{},
			dailyTotal:            sync.Map{},
			closed:                acc.Closed,
			description:           acc.Description,
			created:               acc.Created,
			type_:                 acc.Type,
			accountNumber:         acc.AccountNumber,
			sortCode:              acc.SortCode,
			owners:                acc.Owners,
			user:                  user,
		}

		a.accountsLock.Lock()
		a.accounts[acc.Id] = account
		a.accountsLock.Unlock()

		user = &User{
			id:       user.id,
			auth:     user.auth,
			accounts: append(user.accounts, account),
		}
	}

	if !usersLocked {
		a.usersLock.Lock()
	}
	a.users[response.UserId] = user
	if !usersLocked {
		a.usersLock.Unlock()
	}

	return nil
}

func (a *MonzoApi) RefreshAuth() error {
	log.Println("Refreshing auth")

	a.usersLock.Lock()
	for _, user := range a.users {

		client := &http.Client{}

		form := url.Values{}
		form.Add("grant_type", "refresh_token")
		form.Add("client_id", a.clientConfig.ClientId)
		form.Add("client_secret", a.clientConfig.ClientSecret)
		form.Add("refresh_token", user.auth.RefreshToken)

		res, err := client.PostForm("https://api.monzo.com/oauth2/token", form)
		if err != nil {
			log.Printf("Error posting for token %+v", err)
			return err
		}

		if res.Status != "200 OK" {
			log.Printf("auth response is not 200. Is %+v", res.Status)
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

		err = a.saveUserAndAccounts(result, true)
		if err != nil {
			return err
		}
	}
	a.usersLock.Unlock()

	return nil
}
