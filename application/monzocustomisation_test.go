package application

import (
	"github.com/tmilner/monzo-customisation/adapters/monzorestclient"
	"github.com/twinj/uuid"
	"reflect"
	"sync"
	"testing"
	"time"
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
