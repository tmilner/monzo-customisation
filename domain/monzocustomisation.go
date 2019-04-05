package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/tmilner/monzo-customisation/monzoclient"
	"github.com/twinj/uuid"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type MonzoCustomisation struct {
	client       *monzoclient.MonzoClient
	config       *Config
	users        map[string]*User
	usersLock    sync.RWMutex
	accounts     map[string]*Account
	accountsLock sync.RWMutex
	ticker       *time.Ticker
	stateToken   string
}

type Config struct {
	ClientId     string
	ClientSecret string
	URI          string
	WebhookURI   string
	RedirectUri  string
}

type User struct {
	id       string
	auth     *Auth
	accounts []*Account
}

type Auth struct {
	AccessToken  string
	ClientId     string
	Expiry       int32
	RefreshToken string
	TokenType    string
	UserId       string
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
	owners                []Owner
	user                  *User
}

type Owner struct {
	UserId             string
	PreferredName      string
	PreferredFirstName string
}

type WebhookResponse struct {
	TransactionType string                                 `json:"type"`
	Data            monzoclient.TransactionDetailsResponse `json:"data"`
}

func CreateMonzoCustomisation(client *monzoclient.MonzoClient, config *Config) *MonzoCustomisation {
	monzo := &MonzoCustomisation{
		client:       client,
		config:       config,
		users:        map[string]*User{},
		usersLock:    sync.RWMutex{},
		accounts:     map[string]*Account{},
		accountsLock: sync.RWMutex{},
		ticker:       time.NewTicker(2 * time.Hour),
		stateToken:   uuid.NewV4().String(),
	}
	go monzo.extendAuth()

	errorChain := alice.New(loggerHandler, recoverHandler, timeoutHandler)

	router := mux.NewRouter()
	router.HandleFunc("/", genericIgnore)
	router.HandleFunc("/webhook", monzo.webhookHandler).Methods("POST")
	router.HandleFunc("/auth_return", monzo.authReturnHandler).Methods("GET")
	router.HandleFunc("/auth_start", monzo.authHandler).Methods("GET")
	log.Println("Setting up webhook server")
	_ = http.ListenAndServe(":80", errorChain.Then(router))

	return monzo

}

func genericIgnore(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	_, _ = io.WriteString(w, "Get off my server you prick. You wont find anything here.")
}

func loggerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("new request: %s %s %v (from: %s)", r.Method, r.URL.Path, time.Since(start), r.Header.Get("X-Real-Ip"))
	})
}

func recoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func timeoutHandler(h http.Handler) http.Handler {
	return http.TimeoutHandler(h, 1*time.Second, "timed out")
}

func (a *MonzoCustomisation) extendAuth() {
	for {
		select {
		case <-a.ticker.C:
			_ = a.refreshAuth()
		}
	}
}

func (a *MonzoCustomisation) findUserForAccount(accountId string) (*User, error) {
	if acc := a.accounts[accountId]; acc != nil {
		return acc.user, nil
	} else {
		return nil, errors.New("cannot find Account")
	}
}

func (a *MonzoCustomisation) processTodaysTransactions(userId string) {
	a.usersLock.RLock()

	var user = a.users[userId]
	var today = timeToDate(time.Now())

	for _, acc := range user.accounts {
		res, err := a.client.GetTransactionsSinceTimestamp(acc.id, user.auth.AccessToken, today)
		if err != nil {
			return
		}

		for _, transact := range res.Transactions {
			a.handleTransaction(&transact)
		}

		dailyTotal, found := acc.dailyTotal.Load(today)
		if found {
			log.Printf("Processed todays transactions [%d] for account %s. Total is: %d", len(res.Transactions), acc.type_, dailyTotal)
		} else {
			log.Printf("Processed todays transactions [%d] for account %s. Found none.", len(res.Transactions), acc.type_)
		}
	}
	a.usersLock.RUnlock()

}

func timeToDate(timestamp time.Time) time.Time {
	return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, timestamp.Location())
}

func (a *MonzoCustomisation) runBasicInfo(userId string) {
	authToken := a.users[userId].auth.AccessToken

	listAccountRes, err := a.client.ListAccounts(authToken)
	if err != nil {
		log.Fatalln("ListAccount error", err)
	}

	log.Println("Retrieving pots:")
	pots, err := a.client.GetPots(authToken)
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
			balance, err := a.client.GetBalance(account.Id, authToken)
			if err != nil {
				log.Printf("Error getting balance: %+v", err)
			}
			log.Printf("Balance for account %s is %d", account.Type, balance.Balance)

			params := &monzoclient.Params{
				Title:    "tmilner.co.uk Authenticated!",
				Body:     "Woop Woop",
				ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
			}

			log.Printf("Creating a feed item: %+v", params)
			feedErr := a.createFeedItem(account.Id, params)
			if feedErr != nil {
				log.Printf("Feed error: %+v", feedErr)
			}

			err = a.registerWebhook(account.Id)
			if err != nil {
				log.Printf("Error creting webhook: %+v", err)
			}
		}
	}
}

func (a *MonzoCustomisation) saveUserAndAccounts(response *Auth, usersLocked bool) error {
	if !usersLocked {
		a.usersLock.Lock()
	}
	user := &User{
		id:       response.UserId,
		auth:     response,
		accounts: make([]*Account, 0),
	}
	a.users[response.UserId] = user

	accountRes, err := a.client.ListAccounts(response.AccessToken)
	if err != nil {
		log.Printf("Failed to get account info for authorised account %+v", err)
		return errors.New("failed to get account info")
	}

	for _, acc := range accountRes.Accounts {
		if !acc.Closed {
			owners := make([]Owner, len(acc.Owners))
			for index, owner := range acc.Owners {
				owners[index] = Owner(owner)
			}

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
				owners:                owners,
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
	}

	a.users[response.UserId] = user
	if !usersLocked {
		a.usersLock.Unlock()
	}

	return nil
}

func (a *MonzoCustomisation) refreshAuth() error {
	log.Println("Refreshing auth")

	a.usersLock.Lock()
	for _, user := range a.users {
		res, err := a.client.RefreshAuth(user.auth.AccessToken, a.config.ClientId, a.config.ClientSecret)
		if err != nil {
			return err
		}

		auth := Auth(*res)
		return a.saveUserAndAccounts(&auth, true)
	}
	a.usersLock.Unlock()

	return nil
}

func (a *MonzoCustomisation) createFeedItem(accountId string, params *monzoclient.Params) error {
	a.accountsLock.RLock()
	account, found := a.accounts[accountId]
	a.accountsLock.RUnlock()
	if !found {
		return errors.New("account not found")
	}

	a.usersLock.RLock()
	authToken := account.user.auth.AccessToken
	a.usersLock.RUnlock()

	feedItem := &monzoclient.FeedItem{
		AccountId: accountId,
		TypeParam: "basic",
		Url:       "http://tmilner.co.uk",
		Params:    params,
	}

	return a.client.CreateFeedItem(feedItem, authToken)
}

func (a *MonzoCustomisation) registerWebhook(accountId string) error {
	a.accountsLock.RLock()
	account, found := a.accounts[accountId]
	a.accountsLock.RUnlock()

	a.usersLock.RLock()
	accessToken := account.user.auth.AccessToken
	a.usersLock.RUnlock()

	if !found {
		return errors.New("account not found")
	}

	return a.client.RegisterWebhook(accountId, accessToken, a.config.WebhookURI)
}

func (a *MonzoCustomisation) authHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("auth Reuqest recieved: %+v", r)
	uri := "https://auth.monzo.com/?client_id=" + a.config.ClientId + "&redirect_uri=" + a.config.RedirectUri + "&response_type=code&state=" + a.stateToken

	http.Redirect(w, r, uri, 303)
}

func (a *MonzoCustomisation) authReturnHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Auth_Return received!")
	code := r.URL.Query().Get("code")
	stateReturned := r.URL.Query().Get("state")

	log.Printf("Got code: %s", code)

	if stateReturned != a.stateToken {
		log.Println("State token is not correct!")
		_, _ = io.WriteString(w, "Fuck Off.")
		return
	}

	res, err := a.client.Authenticate(code, a.config.ClientId, a.config.ClientSecret, a.config.RedirectUri)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	auth := Auth(*res)

	err = a.saveUserAndAccounts(&auth, false)
	if err != nil {
		return
	}

	go a.processTodaysTransactions(res.UserId)
	go a.runBasicInfo(res.UserId)

	_, _ = io.WriteString(w, "Suck it.")
}

func (a *MonzoCustomisation) webhookHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var result WebhookResponse
	err := decoder.Decode(&result)

	if err != nil {
		log.Println("Error decoding webhook")
		_, _ = io.WriteString(w, "Failed")
		return
	}
	log.Printf("Recieved new transaction! %+v", result)
	w.WriteHeader(http.StatusOK)
	a.handleTransaction(&result.Data)
}

func (a *MonzoCustomisation) handleTransaction(transaction *monzoclient.TransactionDetailsResponse) {
	a.accountsLock.Lock()
	if account, found := a.accounts[transaction.AccountId]; found {
		if _, found := account.processedTransactions.Load(transaction.Id); !found {

			account.processedTransactions.Store(transaction.Id, transaction)

			dailyTotal, found := account.dailyTotal.Load(timeToDate(transaction.Created))
			if !found {
				dailyTotal = transaction.Amount
			} else {
				dailyTotal = dailyTotal.(int64) + transaction.Amount
			}
			account.dailyTotal.Store(timeToDate(transaction.Created), dailyTotal)

			var params *monzoclient.Params

			if dailyTotal.(int64) < -5000 {
				params = &monzoclient.Params{
					Title:    "Spending a bit much aren't we?",
					Body:     fmt.Sprintf("Daily spend is at %d! Chill your spending!", dailyTotal.(int64)),
					ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
				}
			} else if transaction.Amount < -10000 {
				params = &monzoclient.Params{
					Title:    "What the fuck is this Mr Big Spender!",
					Body:     fmt.Sprintf("Daily spend is at %d! Chill your spending!", dailyTotal.(int64)),
					ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
				}
			}

			if params != nil {
				_ = a.createFeedItem(transaction.AccountId, params)
			}

		}
		a.accounts[transaction.AccountId] = account
	}
	a.accountsLock.Unlock()
}
