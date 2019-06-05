package monzorestclient

import (
	"encoding/json"
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

func (a *MonzoRestClient) GetBalance(accountId string, authToken string) (*BalanceResponse, error) {
	body, err := a.processGetRequest("/balance?account_id="+accountId, authToken)

	if err != nil {
		return nil, err
	}

	var result *BalanceResponse
	err = json.Unmarshal(body, &result)
	return result, err
}
