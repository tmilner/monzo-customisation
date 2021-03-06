package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/tmilner/monzo-customisation/adapters/monzorestclient"
	"github.com/twinj/uuid"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

//TODO [TM] Move response objects out of client impl and split up this into multiple files
type MonzoClient interface {
	GetTransactions(accountId string, authToken string) (*monzorestclient.TransactionsResponse, error)
	UpdateTransaction(transactionId string, authToken string, metadata map[string]string) (*monzorestclient.TransactionsResponse, error)
	GetTransactionsSinceTimestamp(accountId string, authToken string, timestamp string) (*monzorestclient.TransactionsResponse, error)
	GetPots(authToken string) (*monzorestclient.PotsResponse, error)
	GetBalance(accountId string, authToken string) (*monzorestclient.BalanceResponse, error)
	ListAccounts(authToken string) (*monzorestclient.AccountListResponse, error)
	CreateFeedItem(item *monzorestclient.FeedItem, authToken string) error
	RegisterWebhook(accountId string, accessToken string, uri string) error
	Authenticate(code string, clientId string, clientSecret string, redirectUri string) (*monzorestclient.AuthResponse, error)
	RefreshAuth(auth string, clientId string, clientSecret string) (*monzorestclient.AuthResponse, error)
	WhoAmI(authToken string) (*monzorestclient.WhoAmIResponse, error)
}

type MonzoCustomisation struct {
	client       MonzoClient
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
	dailyInfo             sync.Map
	closed                bool
	description           string
	created               string
	type_                 string
	accountNumber         string
	sortCode              string
	owners                []Owner
	user                  *User
}

type DailyInfo struct {
	total                        int64
	sent100QuidLimitNotification bool
}
type Owner struct {
	UserId             string
	PreferredName      string
	PreferredFirstName string
}

type WebhookResponse struct {
	TransactionType string                                     `json:"type"`
	Data            monzorestclient.TransactionDetailsResponse `json:"data"`
}

func CreateMonzoCustomisation(client *monzorestclient.MonzoRestClient, config *Config) *MonzoCustomisation {
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
	router.HandleFunc("/webhook", monzo.webhookHandler).Methods("POST")
	router.HandleFunc("/auth_return", monzo.authReturnHandler).Methods("GET")
	router.HandleFunc("/auth_start", monzo.authHandler).Methods("GET")
	log.Println("Setting up webhook server")
	_ = http.ListenAndServe(":80", errorChain.Then(router))

	return monzo

}

func loggerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//start := time.Now()
		h.ServeHTTP(w, r)
		//log.Printf("#### #### new request: %s %s %v (from: %s)", r.Method, r.URL.Path, time.Since(start), r.Header.Get("X-Real-Ip"))
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
	defer a.usersLock.RUnlock()
	defer a.accountsLock.Unlock()

	a.usersLock.RLock()
	a.accountsLock.Lock()

	var user = a.users[userId]

	var today = timeToDate(time.Now())

	for _, acc := range user.accounts {
		res, err := a.client.GetTransactionsSinceTimestamp(acc.id, user.auth.AccessToken, today)
		if err != nil {
			log.Println("Error getting transactions for today! :( ")
			return
		} else {
			log.Printf("Got transactison! %d", len(res.Transactions))
		}

		for _, transact := range res.Transactions {
			a.handleTransaction(&transact, true, true)
		}

		if dailyTotal, found := acc.dailyInfo.Load(today); found {
			log.Printf("Processed todays transactions [%d] for account %s. Total is: %d", len(res.Transactions), acc.type_, dailyTotal.(DailyInfo).total)
		} else {
			log.Printf("Processed todays transactions [%d] for account %s. Found none.", len(res.Transactions), acc.type_)
		}
	}
	a.usersLock.RUnlock()

}

func timeToDate(timestamp time.Time) string {
	return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, timestamp.Location()).Format(time.RFC3339)
}

func (a *MonzoCustomisation) runBasicInfo(userId string) {
	defer a.usersLock.RUnlock()
	defer a.accountsLock.Unlock()
	a.usersLock.RLock()
	a.accountsLock.Lock()

	user := a.users[userId]
	authToken := a.users[userId].auth.AccessToken

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

	for _, account := range user.accounts {
		if !account.closed {
			balance, err := a.client.GetBalance(account.id, authToken)
			if err != nil {
				log.Printf("Error getting balance: %+v", err)
			}
			log.Printf("Balance for account %s is %d", account.type_, balance.Balance)

			params := &monzorestclient.Params{
				Title:    "tmilner.co.uk Authenticated!",
				Body:     "Woop Woop",
				ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
			}

			log.Printf("Creating a feed item: %+v", params)
			feedErr := a.createFeedItem(account.id, params, true, true)
			if feedErr != nil {
				log.Printf("Feed error: %+v", feedErr)
			}

			err = a.registerWebhook(account.id)
			if err != nil {
				log.Printf("Error creting webhook: %+v", err)
			}
		}
	}
}

func (a *MonzoCustomisation) saveUserAndAccounts(response *Auth, usersLocked bool) error {
	if !usersLocked {
		defer a.usersLock.Unlock()
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
				dailyInfo:             sync.Map{},
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

func (a *MonzoCustomisation) createFeedItem(accountId string, params *monzorestclient.Params, hasAccountLock bool, hasUserReadLock bool) error {
	if !hasAccountLock {
		a.accountsLock.RLock()
		defer a.accountsLock.RUnlock()
	}
	if !hasUserReadLock {
		a.usersLock.RLock()
		defer a.usersLock.RUnlock()
	}

	account, found := a.accounts[accountId]
	if !found {
		return errors.New("account not found")
	}

	authToken := account.user.auth.AccessToken

	feedItem := &monzorestclient.FeedItem{
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
	defer r.Body.Close()

	uri := "https://auth.monzo.com/?client_id=" + a.config.ClientId + "&redirect_uri=" + a.config.RedirectUri + "&response_type=code&state=" + a.stateToken

	http.Redirect(w, r, uri, 303)
}

func (a *MonzoCustomisation) authReturnHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	log.Println("Auth_Return received!")
	code := r.URL.Query().Get("code")
	stateReturned := r.URL.Query().Get("state")

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
	defer req.Body.Close()
	decoder := json.NewDecoder(req.Body)
	var result WebhookResponse
	err := decoder.Decode(&result)

	if err != nil {
		log.Printf("Error decoding webhook: %v", decoder)
		http.NotFound(w, req)
		return
	}
	w.WriteHeader(http.StatusOK)
	a.handleTransaction(&result.Data, false, false)
}

func (a *MonzoCustomisation) handleTransaction(transaction *monzorestclient.TransactionDetailsResponse, hasAccountLock bool, hasUserLock bool) {
	if !hasAccountLock {
		defer a.accountsLock.Unlock()
		a.accountsLock.Lock()
	}
	if account, found := a.accounts[transaction.AccountId]; found {
		if _, found := account.processedTransactions.Load(transaction.Id); !found {

			log.Printf("New Tranasaction! %v", transaction)

			account.processedTransactions.Store(transaction.Id, transaction)
			//TODO: Check if this is a pot transfer before counting towards the daily total.
			transCreated := timeToDate(transaction.Created)

			dailyInfoI, found := account.dailyInfo.Load(transCreated)
			var dailyInfo DailyInfo

			if !found {
				dailyInfo = DailyInfo{total: transaction.Amount, sent100QuidLimitNotification: false}
			} else {
				dailyInfo = dailyInfoI.(DailyInfo)
				dailyInfo = DailyInfo{total: dailyInfo.total + transaction.Amount, sent100QuidLimitNotification: dailyInfo.sent100QuidLimitNotification}
			}

			var params *monzorestclient.Params

			log.Printf("Current Daily Total: %d (%s)", dailyInfo.total, transCreated)

			if dailyInfo.total < -5000 {
				log.Println("Spent more than 50 at once! Chill")
				spending := (dailyInfo.total / 100) * -1
				params = &monzorestclient.Params{
					Title:    "Spending a bit much aren't we?",
					Body:     fmt.Sprintf("You swpnt more than £50! Daily spend is at £%d! Chill your spending!", spending),
					ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
				}
			} else if transaction.Amount < -10000 && !dailyInfo.sent100QuidLimitNotification {
				log.Println("Spent more than 100 in a day! Big spender")
				dailyInfo = DailyInfo{total: dailyInfo.total, sent100QuidLimitNotification: true}
				spending := (dailyInfo.total / 100) * -1
				params = &monzorestclient.Params{
					Title:    "What the fuck is this Mr Big Spender!",
					Body:     fmt.Sprintf("Daily spend is at %d! Chill your spending!", spending),
					ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
				}
			}

			account.dailyInfo.Store(transCreated, dailyInfo)

			if transaction.Merchant.Name == "Tfl Cycle Hire" {
				log.Println("Boris Bikes!")
				_, err := a.client.UpdateTransaction(transaction.Id, account.user.auth.AccessToken, map[string]string{"notes": "#cyceling"})
				if err != nil {
					log.Print("Updated Boris Bike transaction.")
				}
			} else if transaction.Merchant.Name == "Amoret Coffee" {
				log.Println("Coffeeeee")
				_, err := a.client.UpdateTransaction(transaction.Id, account.user.auth.AccessToken, map[string]string{"notes": "#coffee"})
				if err != nil {
					log.Print("Updated Amoret transaction")
				}
			}

			if params != nil {
				log.Println("Creating feed item.")
				err := a.createFeedItem(transaction.AccountId, params, hasAccountLock, hasUserLock)
				if err != nil {
					log.Printf("Error creating feed item for transaction: %s", transaction.Id)
				}
			}
			a.accounts[transaction.AccountId] = account
		} else {
			log.Printf("Recieved duplicate webhook call: %v", transaction)
		}
	} else {
		log.Printf("Tried to process transaction for acount %s but account not found", transaction.AccountId)
	}
}
