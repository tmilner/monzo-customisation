package httpclient

import (
	"encoding/json"
	"github.com/tmilner/monzo-customisation/configuration"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type TransactionsResponse struct {
	Transactions []TransactionDetailsResponse `json:"transactions"`
}

type TransactionResponse struct {
	Transaction TransactionDetailsResponse `json:"transaction"`
}

type TransactionDetailsResponse struct {
	AccountBalance int64            `json:"account_balance"`
	Amount         int64            `json:"amount"`
	Created        time.Time        `json:"created"`
	Currency       string           `json:"currency"`
	Description    string           `json:"description"`
	Id             string           `json:"id"`
	Merchant       MerchantResponse `json:"merchant,omitempty"`
	Notes          string           `json:"notes"`
	IsLoad         bool             `json:"is_lode"`
	Settled        string           `json:"settled"`
	DeclineReason  string           `json:"decline_reason,omitempty"`
	Category       string           `json:"category"`
}

type MerchantResponse struct {
	Created  time.Time       `json:"created"`
	Id       string          `json:"id"`
	Logo     string          `json:"logo"`
	Emoji    string          `json:"emoji"`
	Name     string          `json:"name"`
	Category string          `json:"category"`
	Address  AddressResponse `json:"address"`
	Atm      bool            `json:"atm"`
}

type AddressResponse struct {
	Address        string  `json:"address"`
	City           string  `json:"city"`
	Country        string  `json:"country"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Postcode       string  `json:"postcode"`
	Region         string  `json:"region"`
	Formatted      string  `json:"formatted"`
	ShortFormatted string  `json:"short_formatted"`
}

func Transactions(client *http.Client, config *configuration.Configuration, accountId string) (*TransactionsResponse, error) {
	req, err := http.NewRequest("GET", monzoapi+"/transactions?expand[]=merchant&account_id="+accountId, nil)
	req.Header.Add("Authorization", "Bearer "+config.Authorization)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	var result TransactionsResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}
