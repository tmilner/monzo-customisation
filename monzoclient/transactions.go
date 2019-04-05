package monzoclient

import (
	"encoding/json"
	"time"
)

type TransactionsResponse struct {
	Transactions []TransactionDetailsResponse `json:"transactions"`
}

type TransactionResponse struct {
	Transaction TransactionDetailsResponse `json:"transaction"`
}

type TransactionDetailsResponse struct {
	AccountId      string           `json:"account_id,omitempty"`
	AccountBalance int64            `json:"account_balance,omitempty"`
	Amount         int64            `json:"amount"`
	Created        time.Time        `json:"created"`
	Currency       string           `json:"currency"`
	Description    string           `json:"description"`
	Id             string           `json:"id"`
	Merchant       MerchantResponse `json:"merchant,omitempty"`
	Notes          string           `json:"notes,omitempty"`
	IsLoad         bool             `json:"is_load"`
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
	Atm      bool            `json:"atm,omitempty"`
}

type AddressResponse struct {
	Address        string  `json:"address"`
	City           string  `json:"city"`
	Country        string  `json:"country"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Postcode       string  `json:"postcode"`
	Region         string  `json:"region"`
	Formatted      string  `json:"formatted,omitempty"`
	ShortFormatted string  `json:"short_formatted,omitempty"`
}

func (a *MonzoClient) GetTransactions(accountId string, authToken string) (*TransactionsResponse, error) {
	body, err := a.processGetRequest("/transactions?expand[]=merchant&account_id="+accountId, authToken)
	if err != nil {
		return nil, err
	}

	var result TransactionsResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}

func (a *MonzoClient) GetTransactionsSinceTimestamp(accountId string, authToken string, timestamp time.Time) (*TransactionsResponse, error) {
	body, err := a.processGetRequest("/transactions?expand[]=merchant&account_id="+accountId+"&since="+timestamp.Format(time.RFC3339), authToken)
	if err != nil {
		return nil, err
	}

	var result TransactionsResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}
