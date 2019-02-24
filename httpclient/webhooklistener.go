package httpclient

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		io.WriteString(w, "FAILED")
	}

	log.Printf("Recieved messgae: %+v", string(body))
	io.WriteString(w, "Received")
}

func SetupWebhookInterface() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	log.Println("Setting up webhook server")
	http.ListenAndServe(":80", mux)
}
