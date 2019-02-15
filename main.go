package main

import (
	"github.com/tmilner/monzo-customisation/configuration"
	"github.com/tmilner/monzo-customisation/httpclient"
	"log"
	"net/http"
	"sort"
)

func main() {
	client := &http.Client{}
	config := configuration.New()

	//whoAmIRes, err := httpclient.WhoAmI(client, config)
	//if err != nil {
	//	log.Println("WhoAmI error", err)
	//}

	listAccountRes, err := httpclient.ListAccounts(client, config)
	if err != nil {
		log.Fatalln("ListAccount error", err)
	}

	log.Println("Retrieving account balances:")
	for index, account := range listAccountRes.Accounts {
		balance, err := httpclient.Balance(client, config, account.Id)
		if err != nil {
			log.Fatalln("Error getting balance", err)
		}
		log.Printf("%d Balance for account %s is %d", index, account.Type, balance.Balance)
	}

	pots, err := httpclient.Pots(client, config)
	if err != nil {
		log.Fatalln("Pots error", err)
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
		transactions, err := httpclient.Transactions(client, config, account.Id)
		if err != nil {
			log.Fatalln("Error getting transactions", err)
		}

		var maxSpend float64 = 0
		var maxIncoming float64 = 0

		merchantCount := make(map[string]int)

		for _, transaction := range transactions.Transactions {
			if float64(transaction.Amount) < maxSpend {
				maxSpend = float64(transaction.Amount)
			}
			if float64(transaction.Amount) > maxIncoming {
				maxIncoming = float64(transaction.Amount)
			}
			merchantCount[transaction.Merchant.Name]++
		}

		log.Printf("Max income = %f! Max Spend = %f", maxIncoming/100.0, (maxSpend*-1)/100.0)
		log.Printf("%+v", rank(merchantCount))
	}
}

func rank(merchantCounts map[string]int) PairList {
	pl := make(PairList, len(merchantCounts))
	i := 0
	for k, v := range merchantCounts {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Merchant   string
	VisitCount int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].VisitCount < p[j].VisitCount }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
