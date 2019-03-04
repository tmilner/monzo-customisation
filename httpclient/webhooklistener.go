package httpclient

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type WebhookResponse struct {
	TransactionType string                     `json:"type"`
	Data            TransactionDetailsResponse `json:"data"`
}

func (a *MonzoApi) WebhookHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var result WebhookResponse
	err := decoder.Decode(&result)

	if err != nil {
		log.Println("Error decoding webhook")
		_, _ = io.WriteString(w, "Failed")
		return
	}
	log.Printf("Recieved new transaction! %+v", result)
	go a.handleWebhook(&result)
	_, _ = io.WriteString(w, "Suck it.")
}

func (a *MonzoApi) handleWebhook(w *WebhookResponse) {
	var params *Params
	if w.Data.Amount > 50 {
		params = &Params{
			Title:    "Spending a bit much aren't we?",
			Body:     "Tut tut 💸💸💸💸💸",
			ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
		}
	} else if w.Data.AccountBalance < 50 {
		params = &Params{
			Title:    "YOURE POOR STOP SPENDING MONEY",
			Body:     "GOD DAMN IT YOU FOOL 💸💸💸💸💸",
			ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
		}
	}

	if params != nil {
		feedItem := &FeedItem{
			AccountId: w.Data.AccountId,
			TypeParam: "basic",
			Url:       "http://tmilner.co.uk",
			Params:    *params,
		}

		_ = a.CreateFeedItem(feedItem)
	}
}

func (a *MonzoApi) RegisterWebhook(accountId string) error {
	form := url.Values{}
	form.Add("account_id", accountId)
	form.Add("url", a.ClientConfig.WebhookURI)

	req, err := http.NewRequest("POST", a.URL+"/webhooks", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+a.Auth.AccessToken)

	res, lastErr := a.Client.Do(req)

	if (res.Status != "200 OK" && res.Status != "201 Created") || lastErr != nil {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Failed in many ways :'( ")
			return err
		}

		log.Printf("Not 200 or 201! is %s. Response is: %+v", res.Status, string(body))
		return errors.New("not 200 or 201")
	}

	log.Println("Registered webhook")

	return nil
}
