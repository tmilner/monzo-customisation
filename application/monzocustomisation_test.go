package application

import (
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/tmilner/monzo-customisation/adapters/monzorestclient"
	"github.com/twinj/uuid"
)

func TestCreateMonzoApi(t *testing.T) {
	type args struct {
		config *Config
		client *monzorestclient.MonzoRestClient
	}

	config := &Config{}
	client := &monzorestclient.MonzoRestClient{}

	tests := []struct {
		name string
		args args
		want *MonzoCustomisation
	}{
		{
			name: "Values are initialised correct",
			args: args{
				config: config,
				client: client,
			},
			want: &MonzoCustomisation{
				client:       client,
				config:       config,
				users:        map[string]*User{},
				usersLock:    sync.RWMutex{},
				accounts:     map[string]*Account{},
				accountsLock: sync.RWMutex{},
				ticker:       time.NewTicker(2 * time.Hour),
				stateToken:   uuid.NewV4().String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateMonzoCustomisation(tt.args.client, tt.args.config)
			got.stateToken = tt.want.stateToken
			got.ticker = tt.want.ticker
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateMonzoCustomisation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonzoApi_findUserForAccount(t *testing.T) {
	type fields struct {
		url          string
		client       *monzorestclient.MonzoRestClient
		clientConfig *Config
		users        map[string]*User
		usersLock    sync.RWMutex
		accounts     map[string]*Account
		accountsLock sync.RWMutex
	}
	type args struct {
		accountId string
	}

	accountId := "12345"
	user := &User{}
	account := &Account{
		id:   accountId,
		user: user,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *User
		wantErr bool
	}{
		{
			name: "Return the user details when an account is found",
			fields: fields{
				accountsLock: sync.RWMutex{},
				accounts: map[string]*Account{
					accountId: account,
				},
			},
			args:    args{accountId: accountId},
			want:    user,
			wantErr: false,
		},
		{
			name: "Returns an error when the account ID cannot be found",
			fields: fields{
				accountsLock: sync.RWMutex{},
				accounts:     map[string]*Account{},
			},
			args:    args{accountId: accountId},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &MonzoCustomisation{
				client:       tt.fields.client,
				config:       tt.fields.clientConfig,
				users:        tt.fields.users,
				usersLock:    tt.fields.usersLock,
				accounts:     tt.fields.accounts,
				accountsLock: tt.fields.accountsLock,
			}
			got, err := a.findUserForAccount(tt.args.accountId)
			if (err != nil) != tt.wantErr {
				t.Errorf("MonzoCustomisation.findUserForAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MonzoCustomisation.findUserForAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*func TestMonzoApi_runBasicInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		client       *monzorestclient.MonzoRestClient
		config       *Config
		users        map[string]*User
		usersLock    sync.RWMutex
		accounts     map[string]*Account
		accountsLock sync.RWMutex
	}
	type args struct {
		userId string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "",
			fields: fields{
				client:
			},
			args:   args{
				userId: "23232",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &MonzoCustomisation{
				client:       tt.fields.client,
				config:       tt.fields.config,
				users:        tt.fields.users,
				usersLock:    tt.fields.usersLock,
				accounts:     tt.fields.accounts,
				accountsLock: tt.fields.accountsLock,
			}
			a.runBasicInfo(tt.args.userId)
		})
	}
}
*/

func TestMonzoCustomisation_handleTransaction(t *testing.T) {
	type fields struct {
		users        map[string]*User
		usersLock    sync.RWMutex
		accounts     map[string]*Account
		accountsLock sync.RWMutex
	}
	type args struct {
		transaction    *monzorestclient.TransactionDetailsResponse
		expectedAmount int64
	}
	monzoclient := monzorestclient.CreateMonzoRestClient("localhost", &http.Client{})
	config := &Config{
		ClientId:     "TEST",
		ClientSecret: "",
		URI:          "",
		RedirectUri:  "",
		WebhookURI:   "",
	}
	var account *Account
	user := &User{"User123", nil, []*Account{account}}
	dateWithExistingTransactions := time.Date(1991, time.December, 04, 12, 04, 12, 0, time.UTC)
	account = &Account{
		id: "12345",
		processedTransactions: sync.Map{},
		dailyTotal:            sync.Map{},
		closed:                false,
		description:           "",
		created:               "",
		type_:                 "retail",
		accountNumber:         "123456",
		sortCode:              "04-00-04",
		owners:                []Owner{{user.id, "123", "12"}},
		user:                  user,
	}
	account.dailyTotal.Store(timeToDate(dateWithExistingTransactions), int64(-500))
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Daily total is the same as the transaction amount if its the first of the day",
			fields: fields{
				users: map[string]*User{
					user.id: user,
				},
				usersLock: sync.RWMutex{},
				accounts: map[string]*Account{
					account.id: account,
				},
				accountsLock: sync.RWMutex{},
			},
			args: args{
				transaction: &monzorestclient.TransactionDetailsResponse{
					AccountId:   account.id,
					Amount:      -500,
					Created:     time.Now(),
					Currency:    "GBP",
					Description: "Pret",
					Id:          "123",
					Settled:     "23423",
					Category:    "food",
				},
				expectedAmount: -500,
			},
		},
		{
			name: "Daily total is updated for the transaction amount if there have already been transactions today",
			fields: fields{
				users: map[string]*User{
					user.id: user,
				},
				usersLock: sync.RWMutex{},
				accounts: map[string]*Account{
					account.id: account,
				},
				accountsLock: sync.RWMutex{},
			},
			args: args{
				transaction: &monzorestclient.TransactionDetailsResponse{
					AccountId:   account.id,
					Amount:      -500,
					Created:     dateWithExistingTransactions,
					Currency:    "GBP",
					Description: "Pret",
					Id:          "1432",
					Settled:     "23423",
					Category:    "food",
				},
				expectedAmount: -1000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &MonzoCustomisation{
				client:       monzoclient,
				config:       config,
				users:        tt.fields.users,
				usersLock:    tt.fields.usersLock,
				accounts:     tt.fields.accounts,
				accountsLock: tt.fields.accountsLock,
				ticker:       time.NewTicker(2 * time.Hour),
				stateToken:   uuid.NewV4().String(),
			}
			a.handleTransaction(tt.args.transaction)
			total, found := tt.fields.accounts[account.id].dailyTotal.Load(timeToDate(tt.args.transaction.Created))
			if !found {
				t.Error("Did not store an amount for today!")
			}
			if total.(int64) != tt.args.expectedAmount {
				t.Errorf("daily total is inocrrect! Should be %d, is %d", tt.args.expectedAmount, total.(int64))
			}
		})
	}
}
