package httpclient

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

var monzoapi = "https://api.monzo.com/"

type MonzoApi struct {
	Auth         *AuthResponse
	URL          string
	Client       *http.Client
	ClientConfig *ClientConfig
}

type ClientConfig struct {
	ClientId     string
	ClientSecret string
	RedirectUri  string
}

func CreateMonzoApi(config *ClientConfig) *MonzoApi {
	client := &http.Client{}

	return &MonzoApi{
		ClientConfig: config,
		Client:       client,
		URL:          monzoapi,
	}
}

func (a *MonzoApi) processGetRequest(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", a.URL+path, nil)
	req.Header.Add("Authorization", "Bearer "+a.Auth.AccessToken)
	resp, err := a.Client.Do(req)
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
	//whoAmIRes, err := httpclient.WhoAmI(client, config)
	//if err != nil {
	//	log.Println("WhoAmI error", err)
	//}

	listAccountRes, err := a.ListAccounts()
	if err != nil {
		log.Fatalln("ListAccount error", err)
	}

	log.Println("Retrieving account balances:")
	for index, account := range listAccountRes.Accounts {
		balance, err := a.GetBalance(account.Id)
		if err != nil {
			log.Fatalln("Error getting balance", err)
		}
		log.Printf("%d GetBalance for account %s is %d", index, account.Type, balance.Balance)
	}

	pots, err := a.GetPots()
	if err != nil {
		log.Fatalln("GetPots error", err)
	}

	log.Println("Retrieving pots:")
	for index, pot := range pots.Pots {
		var deleted = "active"
		if pot.Deleted {
			deleted = "deleted"
		}
		log.Printf("%d: Found a %s pot called %s, its got a balence of %d and is currently using the style %s", index, deleted, pot.Name, pot.Balance, pot.Style)
	}

	log.Println("Lets get the transactions for all accounts:")

	for _, account := range listAccountRes.Accounts {
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
			Title:    "Starting Service",
			Body:     "Service is starting",
			ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
		}

		feedItem := &FeedItem{
			TypeParam: "basic",
			AccountId: account.Id,
			Url:       "http://tmilner.co.uk",
			Params:    params,
		}
		log.Printf("Creating a feed item %+v", feedItem)
		feedErr := a.CreateFeedItem(feedItem)
		if feedErr != nil {
			log.Fatalf("Feed error!! %+v", feedErr)
		}
	}
}
