package httpclient

import (
	"encoding/json"
	"github.com/tmilner/monzo-customisation/configuration"
	"io/ioutil"
	"net/http"
)

type BalanceResponse struct {
	Balance                   int64                `json:"balance"`
	TotalBalance              int64                `json:"total_balance"`
	BalanceIncFlexibleSavings int64                `json:"balance_including_flexible_savings"`
	Currency                  string               `json:"currency"`
	SpendToday                int64                `json:"spend_today"`
	LocalCurrency             string               `json:"local_currency"`
	LocalExchangeRate         int64                `json:"local_exchange_rate"`
	LocalSpend                []LocalSpendResponse `json:"local_spend"`
}

type LocalSpendResponse struct {
	SpendToday int64  `json:"spend_today"`
	Currency   string `json:"currency"`
}

func GetBalance(client *http.Client, config *configuration.Configuration, accountId string) (*BalanceResponse, error) {
	req, err := http.NewRequest("GET", monzoapi+"/balance?account_id="+accountId, nil)
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

	var result BalanceResponse
	err = json.Unmarshal(body, &result)

	return &result, err
}
