package httpclient

import (
	"encoding/json"
	. "github.com/tmilner/monzo-customisation/configuration"
	. "github.com/tmilner/monzo-customisation/domain"
	"io/ioutil"
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

func GetTransactions(client *http.Client, config *Configuration, accountId string) (*TransactionsResponse, error) {
	req, err := http.NewRequest("GET", monzoapi+"/transactions?expand[]=merchant&account_id="+accountId, nil)
	req.Header.Add("Authorization", "Bearer "+config.Authorization)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TransactionsResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}

func (t *TransactionsResponse) ToDomain() ([]*Transaction, error) {
	transactions := make([]*Transaction, 0)

	for _, transaction := range t.Transactions {
		domain, err := transaction.ToDomain()
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, domain)
	}

	return transactions, nil
}

func (t *TransactionDetailsResponse) ToDomain() (*Transaction, error) {
	merchant, err := t.Merchant.ToDomain()
	if err != nil {
		return nil, err
	}

	return &Transaction{
		AccountBalance: t.AccountBalance,
		Amount: t.Amount,
		Created: t.Created,
		Currency: t.Currency,
		Description: t.Description,
		Id: t.Id,
		Merchant: merchant,
		Notes: t.Notes,
		IsLoad: t.IsLoad,
		Settled: t.Settled,
		DeclineReason: t.DeclineReason,
		Category: t.Category,
	}, nil
}

func (m *MerchantResponse) ToDomain() (*Merchant, error) {
	addr, err := m.Address.ToDomain()
	if err != nil {
		return nil, err
	}

	return &Merchant{
		Created:  m.Created,
		Id:       m.Id,
		Logo:     m.Logo,
		Emoji:    m.Emoji,
		Name:     m.Name,
		Category: m.Category,
		Address:  addr,
		Atm:      m.Atm,
	}, nil
}

func (a *AddressResponse) ToDomain() (*Address, error) {
	return &Address{
		Address:   a.Address,
		City:      a.City,
		Country:   a.Country,
		Postcode:  a.Postcode,
		Region:    a.Region,
		Formatted: a.Formatted,
	}, nil
}
