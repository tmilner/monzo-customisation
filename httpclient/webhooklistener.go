package httpclient

import (
	"encoding/json"
	"errors"
	"fmt"
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
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "Suck it.")
}

func (a *MonzoApi) handleWebhook(w *WebhookResponse) {
	a.accountsLock.Lock()
	if account, found := a.accounts[w.Data.AccountId]; found {
		if _, found := account.processedTransactions.Load(w.Data.Id); !found {

			account.processedTransactions.Store(w.Data.Id, w.Data)

			dailyTotal, found := account.dailyTotal.Load(w.Data.Created)
			if found {
				dailyTotal = w.Data.AccountBalance
			} else {
				dailyTotal = dailyTotal.(int64) + w.Data.AccountBalance
			}
			account.dailyTotal.Store(w.Data.Created, dailyTotal)

			var params *Params

			if dailyTotal.(int64) < -5000 {
				params = &Params{
					Title:    "Spending a bit much aren't we?",
					Body:     fmt.Sprintf("Daily spend is at %d! Chill your spending!", dailyTotal.(int64)),
					ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
				}
			} else if w.Data.Amount < -10000 {
				params = &Params{
					Title:    "What the fuck is this Mr Big Spender!",
					Body:     fmt.Sprintf("Daily spend is at %d! Chill your spending!", dailyTotal.(int64)),
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
		a.accounts[w.Data.AccountId] = account
	}
	a.accountsLock.Unlock()
}

func (a *MonzoApi) RegisterWebhook(accountId string) error {
	a.accountsLock.RLock()
	account, found := a.accounts[accountId]
	a.accountsLock.RUnlock()

	if !found {
		return errors.New("account not found")
	}

	form := url.Values{}
	form.Add("account_id", accountId)
	form.Add("url", a.clientConfig.WebhookURI)

	req, err := http.NewRequest("POST", a.url+"/webhooks", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	a.usersLock.RLock()
	req.Header.Add("Authorization", "Bearer "+account.user.auth.AccessToken)
	a.usersLock.RUnlock()

	res, lastErr := a.client.Do(req)

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
