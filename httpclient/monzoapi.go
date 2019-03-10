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
	url          string
	client       *http.Client
	clientConfig *ClientConfig
	users        map[string]*User
	usersLock    sync.RWMutex
	accounts     map[string]*Account
	accountsLock sync.RWMutex
}

type ClientConfig struct {
	ClientId     string
	ClientSecret string
	URI          string
	WebhookURI   string
	RedirectUri  string
}

type User struct {
	id       string
	auth     *AuthResponse
	accounts []*Account
}

type Account struct {
	id                    string
	processedTransactions sync.Map
	dailyTotal            sync.Map
	closed                bool
	description           string
	created               string
	type_                 string
	accountNumber         string
	sortCode              string
	owners                []OwnersResponse
	user                  *User
}

func CreateMonzoApi(config *ClientConfig) *MonzoApi {
	client := &http.Client{}

	api := &MonzoApi{
		clientConfig: config,
		client:       client,
		url:          monzoapi,
		users:        map[string]*User{},
		usersLock:    sync.RWMutex{},
		accounts:     map[string]*Account{},
		accountsLock: sync.RWMutex{},
	}
	go extendAuth(api)

	return api

}

func (a *MonzoApi) findUserForAccount(accountId string) (*User, error) {
	if acc := a.accounts[accountId]; acc != nil {
		return acc.user, nil
	} else {
		return nil, errors.New("cannot find Account")
	}
}

func (a *MonzoApi) processGetRequest(path string, userId string) ([]byte, error) {
	req, err := http.NewRequest("GET", a.url+path, nil)
	user := a.users[userId]
	if user == nil {
		log.Printf("Could not find user for request to %s.", path)
		return nil, errors.New("cound not find user")
	}

	req.Header.Add("Authorization", "Bearer "+user.auth.AccessToken)
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

func (a *MonzoApi) runBasicInfo(userId string) {
	listAccountRes, err := a.ListAccounts(userId)
	if err != nil {
		log.Fatalln("ListAccount error", err)
	}

	log.Println("Retrieving pots:")
	pots, err := a.GetPots(userId)
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
			balance, err := a.GetBalance(account.Id, userId)
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
