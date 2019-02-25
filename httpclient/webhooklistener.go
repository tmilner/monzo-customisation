package httpclient

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func (a *MonzoApi) WebhookHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		io.WriteString(w, "FAILED")
	}

	log.Printf("Recieved messgae: %+v", string(body))
	io.WriteString(w, "Suck it.")
}
