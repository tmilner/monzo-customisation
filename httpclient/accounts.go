package httpclient

import (
	"encoding/json"
)

type AccountListResponse struct {
	Accounts []AccountResponse
}
type AccountResponse struct {
	Id            string           `json:"id"`
	Closed        bool             `json:"closed"`
	Description   string           `json:"description"`
	Created       string           `json:"created"`
	Type          string           `json:"type"`
	AccountNumber string           `json:"account_number,omitempty"`
	SortCode      string           `json:"sort_code,omitempty"`
	Owners        []OwnersResponse `json:"owners"`
}
type OwnersResponse struct {
	UserId             string `json:"user_id"`
	PreferredName      string `json:"preferred_name"`
	PreferredFirstName string `json:"preferred_first_name"`
}

func (a *MonzoApi) ListAccounts(userId string) (*AccountListResponse, error) {
	body, err := a.processGetRequest("/accounts", userId)

	if err != nil {
		return nil, err
	}

	var result *AccountListResponse
	err = json.Unmarshal(body, &result)
	return result, err
}
