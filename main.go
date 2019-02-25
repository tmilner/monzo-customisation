package main

import (
	"log"
	"net/http"
	"os"

	. "github.com/tmilner/monzo-customisation/httpclient"
)

func main() {
	log.SetPrefix("[MONZO]")

	log.Println("Starting!")

	if len(os.Args) != 4 {
		log.Fatalln("Not enough arguments supplied")
	}

	clientConfig := &ClientConfig{
		ClientId:     os.Args[1],
		ClientSecret: os.Args[2],
		RedirectUri:  os.Args[3] + "/auth_return",
	}

	monzoApi := CreateMonzoApi(clientConfig)

	setupWebhookInterface(monzoApi)
}

func setupWebhookInterface(api *MonzoApi) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", api.WebhookHandler)
	mux.HandleFunc("/auth", api.AuthHandler)
	mux.HandleFunc("/auth_return", api.AuthReturnHandler)
	log.Println("Setting up webhook server")
	http.ListenAndServe(":80", mux)
}
