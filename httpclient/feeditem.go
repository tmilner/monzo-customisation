package httpclient

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type FeedItem struct {
	AccountId string `json:"account_id"`
	TypeParam string `json:"type"`
	Url       string `json:"url"`
	Params    Params `json:"params"`
}

type Params struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	ImageUrl string `json:"image_url"`
}

func (a *MonzoApi) CreateFeedItem(item *FeedItem) error {
	form := url.Values{}
	form.Add("account_id", item.AccountId)
	form.Add("type", "basic")
	form.Add("url", item.Url)
	form.Add("params[title]", item.Params.Title)
	form.Add("params[body]", item.Params.Body)
	form.Add("params[image_url]", item.Params.ImageUrl)

	req, err := http.NewRequest("POST", a.url+"/feed", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	a.accountsLock.RLock()
	account, found := a.accounts[item.AccountId]
	a.accountsLock.RUnlock()
	if !found {
		return errors.New("account not found")
	}

	a.usersLock.RLock()
	req.Header.Add("Authorization", "Bearer "+account.user.auth.AccessToken)
	a.usersLock.RUnlock()

	res, lastErr := a.client.Do(req)

	if res.Status != "200 OK" && res.Status != "201 Created" {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Failed in many ways :'( ")
			return err
		}

		log.Printf("Not 200 or 201! is %s. Response is: %+v", res.Status, string(body))
		return errors.New("not 200 or 201")
	}

	return lastErr
}
