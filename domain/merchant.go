package domain

import "time"

type Merchant struct {
	Created  time.Time
	Id       string
	Logo     string
	Emoji    string
	Name     string
	Category string
	Address  *Address
	Atm      bool
}
