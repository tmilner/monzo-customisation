package main

import (
	"log"
	"net/http"

	. "github.com/tmilner/monzo-customisation/configuration"
	. "github.com/tmilner/monzo-customisation/httpclient"
)

func main() {
	client := &http.Client{}
	config, err := NewConfig()
	if err != nil {
		log.Fatalf("Config failed to load %+v", err)
	}

	SetupWebhookInterface()

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
			Title:    "Testing",
			Body:     "Testy Test",
			ImageUrl: "https://docs.monzo.com/images/logo-46fdcf49.svg",
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
