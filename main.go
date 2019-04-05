package main

import (
	"github.com/tmilner/monzo-customisation/domain"
	"github.com/tmilner/monzo-customisation/monzoclient"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetPrefix("[MONZO]")

	log.Println("Starting!")

	if len(os.Args) != 4 {
		log.Fatalln("Not enough arguments supplied")
	}

	config := &domain.Config{
		ClientId:     os.Args[1],
		ClientSecret: os.Args[2],
		URI:          os.Args[3],
		RedirectUri:  os.Args[3] + "/auth_return",
		WebhookURI:   os.Args[3] + "/webhook",
	}

	client := monzoclient.CreateMonzoClient("https://api.monzo.com/", &http.Client{})

	_ = domain.CreateMonzoCustomisation(client, config)
}
