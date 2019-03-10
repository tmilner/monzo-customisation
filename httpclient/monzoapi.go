package httpclient

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var monzoapi = "https://api.monzo.com/"

type MonzoApi struct {
	auth                  *AuthResponse
	url                   string
	client                *http.Client
	clientConfig          *ClientConfig
	processedTransactions sync.Map
}

type ClientConfig struct {
	ClientId     string
	ClientSecret string
	URI          string
	WebhookURI   string
	RedirectUri  string
}

func CreateMonzoApi(config *ClientConfig) *MonzoApi {
	client := &http.Client{}

	return &MonzoApi{
		clientConfig: config,
		client:       client,
		url:          monzoapi,
		processedTransactions: sync.Map{},
	}
}

func (a *MonzoApi) processGetRequest(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", a.url+path, nil)
	req.Header.Add("Authorization", "Bearer "+a.auth.AccessToken)
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

func (a *MonzoApi) runBasicInfo() {
	listAccountRes, err := a.ListAccounts()
	if err != nil {
		log.Fatalln("ListAccount error", err)
	}

	log.Println("Retrieving pots:")
	pots, err := a.GetPots()
	if err != nil {
		log.Fatalln("GetPots error", err)
	}

	for _, pot := range pots.Pots {
		if !pot.Deleted {
			log.Printf("Found a pot called %s, its got a balence of %d", pot.Name, pot.Balance)
		}
	}

	log.Println("Running through accounts:")

	for _, account := range listAccountRes.Accounts {
		if !account.Closed {
			balance, err := a.GetBalance(account.Id)
			if err != nil {
				log.Printf("Error getting balance: %+v", err)
			}
			log.Printf("Balance for account %s is %d", account.Type, balance.Balance)

			//transactions, err := a.GetTransactions(account.Id)
			//if err != nil {
			//	log.Fatalln("Error getting transactions", err)
			//}
			//
			//domain, err := transactions.ToDomain()
			//if err != nil {
			//	log.Fatalln("Error converting to domain type")
			//}
			//RankAndPrintMerchants(domain)

			params := Params{
				Title:    "tmilner.co.uk Authenticated!",
				Body:     "Woop Woop",
				ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
			}

			feedItem := &FeedItem{
				TypeParam: "basic",
				AccountId: account.Id,
				Url:       "http://tmilner.co.uk",
				Params:    params,
			}
			log.Printf("Creating a feed item: %+v", feedItem)
			feedErr := a.CreateFeedItem(feedItem)
			if feedErr != nil {
				log.Printf("Feed error: %+v", feedErr)
			}

			err = a.RegisterWebhook(account.Id)
			if err != nil {
				log.Printf("Error creting webhook: %+v", err)
			}
		}
	}
}
