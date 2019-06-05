package monzorestclient

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func (a *MonzoRestClient) RegisterWebhook(accountId string, accessToken string, uri string) error {
	form := url.Values{}
	form.Add("account_id", accountId)
	form.Add("url", uri)

	req, err := http.NewRequest("POST", a.url+"/webhooks", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	res, lastErr := a.client.Do(req)

	if lastErr != nil || (res.Status != "200 OK" && res.Status != "201 Created") {
		if res != nil {
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Println("Failed in many ways :'( ")
				return err
			}

			log.Printf("Registering a web hook... response was not 200 or 201! [%s, %+v]", res.Status, string(body))
		} else {
			log.Printf("An error occured registersing a webhook: [%+v]", lastErr)
		}
		return errors.New("not 200 or 201")
	}

	log.Println("Registered webhook")

	return nil
}
