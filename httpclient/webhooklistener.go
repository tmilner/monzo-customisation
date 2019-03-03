package httpclient

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type WebhookResponse struct {
	TransactionType string                     `json:"type"`
	Data            TransactionDetailsResponse `json:"data"`
}

func (a *MonzoApi) WebhookHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		_, _ = io.WriteString(w, "FAILED")
		return
	}
	var result TransactionsResponse
	err = json.Unmarshal(body, &result)

	log.Printf("Recieved new transaction! %+v", result)
	_, _ = io.WriteString(w, "Suck it.")
}
