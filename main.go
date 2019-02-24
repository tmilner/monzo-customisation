package main

import (
	"github.com/tmilner/monzo-customisation/configuration"
	"log"
	"net/http"
	"os"

	. "github.com/tmilner/monzo-customisation/httpclient"
)

func main() {
	client := &http.Client{}
	config := &configuration.Configuration{
		Authorization: os.Args[1],
	}

	f, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	go runBasicInfo(client, config)

	SetupWebhookInterface()
}

func runBasicInfo(client *http.Client, config *configuration.Configuration) {
	//whoAmIRes, err := httpclient.WhoAmI(client, config)
	//if err != nil {
	//	log.Println("WhoAmI error", err)
	//}

	listAccountRes, err := ListAccounts(client, config)
	if err != nil {
		log.Fatalln("ListAccount error", err)
	}

	log.Println("Retrieving account balances:")
	for index, account := range listAccountRes.Accounts {
		balance, err := GetBalance(client, config, account.Id)
		if err != nil {
			log.Fatalln("Error getting balance", err)
		}
		log.Printf("%d GetBalance for account %s is %d", index, account.Type, balance.Balance)
	}

	pots, err := GetPots(client, config)
	if err != nil {
		log.Fatalln("GetPots error", err)
	}

	log.Println("Retrieving pots:")
	for index, pot := range pots.Pots {
		var deleted = "active"
		if pot.Deleted {
			deleted = "deleted"
		}
		log.Printf("%d: Found a %s pot called %s, its got a balence of %d and is currently using the style %s", index, deleted, pot.Name, pot.Balance, pot.Style)
	}

	log.Println("Lets get the transactions for all accounts:")

	for _, account := range listAccountRes.Accounts {
		//transactions, err := GetTransactions(client, config, account.Id)
		//if err != nil {
		//	log.Fatalln("Error getting transactions", err)
		//}
		//
		//domain, err := transactions.ToDomain()
		//if err != nil {
		//	log.Fatalln("Error converting to domain type")
		//}
		//RankAndPrintMerchants(domain)
		params := Params{
			Title:    "Starting Service",
			Body:     "Service is starting",
			ImageUrl: "https://d33wubrfki0l68.cloudfront.net/673084cc885831461ab2cdd1151ad577cda6a49a/92a4d/static/images/favicon.png",
		}

		feedItem := &FeedItem{
			TypeParam: "basic",
			AccountId: account.Id,
			Url:       "http://tmilner.co.uk",
			Params:    params,
		}
		log.Printf("Creating a feed item %+v", feedItem)
		feedErr := CreateFeedItem(client, config, feedItem)
		if feedErr != nil {
			log.Fatalf("Feed error!! %+v", feedErr)
		}
	}
}