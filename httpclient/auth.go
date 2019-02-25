package httpclient

import (
	"encoding/json"
	"github.com/twinj/uuid"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var state = uuid.NewV4().String()

func (a *MonzoApi) AuthHandler(w http.ResponseWriter, r *http.Request) {
	uri := "https://auth.monzo.com/?client_id=" + a.ClientConfig.ClientId + "&redirect_uri=" + a.ClientConfig.RedirectUri + "&response_type=code&state=" + state

	http.Redirect(w, r, uri, 303)
}

func (a *MonzoApi) AuthReturnHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Auth_Return received!")
	code := r.URL.Query().Get("code")
	stateReturned := r.URL.Query().Get("state")

	if stateReturned != state {
		log.Println("State token is not correct!")
		io.WriteString(w, "Fuck Off.")
	}

	client := &http.Client{}

	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", a.ClientConfig.ClientId)
	form.Add("client_secret", a.ClientConfig.ClientSecret)
	form.Add("code", code)

	res, err := client.PostForm("https://api.monzo.com/oauth2/token", form)
	if err != nil {
		log.Println("Error posting for token")
		io.WriteString(w, "Error")
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
		log.Println("Unable to unmarshal auth response")
		return
	}

	a.Auth = result
	a.runBasicInfo()
}
