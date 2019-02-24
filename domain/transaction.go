package domain

import (
	"log"
	"sort"
	"time"
)

type Transaction struct {
	AccountBalance int64
	Amount         int64
	Created        time.Time
	Currency       string
	Description    string
	Id             string
	Merchant       *Merchant
	Notes          string
	IsLoad         bool
	Settled        string
	DeclineReason  string
	Category       string
}

func RankAndPrintMerchants(transactions []*Transaction) {
	merchantCount := make(map[string]int)
	merchantNames := make(map[string]*Merchant)

	for _, transaction := range transactions {
		if transaction.Merchant != nil {
			merchantCount[transaction.Merchant.Id]++
			merchantNames[transaction.Merchant.Id] = transaction.Merchant
		}
	}

	log.Println("Ranking Merchants by number of visits:")
	for rank, pair := range rank(merchantCount)[0:5] {
		log.Printf("In %d position - %+v with %d visits! (id is %s)", rank+1, merchantNames[pair.MerchantId].Name, pair.VisitCount, pair.MerchantId)
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
	MerchantId string
	VisitCount int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].VisitCount < p[j].VisitCount }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
