package monzorestclient

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCreateMonzoClient(t *testing.T) {
	type args struct {
		url    string
		client *http.Client
	}
	url := "monzo"
	client := &http.Client{}
	tests := []struct {
		name string
		args args
		want *MonzoRestClient
	}{
		{
			name: "Client is created successfully",
			args: args{url, client},
			want: &MonzoRestClient{url, client},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateMonzoRestClient(tt.args.url, tt.args.client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateMonzoRestClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonzoClient_processGetRequest(t *testing.T) {
	type args struct {
		path      string
		authToken string
	}
	authToken := "9876"
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
		server  *httptest.Server
	}{
		{
			name: "Calls the API and returns the response as a byte array if the user is authenticated.",
			args: args{
				path:      "/test",
				authToken: authToken,
			},
			want:    []byte(`hello`),
			wantErr: false,
			server: httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer "+authToken {
					http.Error(w, "Error", http.StatusUnauthorized)
				} else {
					_, _ = w.Write([]byte(`hello`))
				}

			})),
		},
		{
			name: "Calls the API and returns an error if the server returns a non 200 code",
			args: args{
				path:      "/test",
				authToken: authToken,
			},
			want:    nil,
			wantErr: true,
			server: httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Error", http.StatusInternalServerError)
			})),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.server != nil {
				defer tt.server.Close()
				tt.server.Start()
			}

			a := &MonzoRestClient{
				url:    tt.server.URL,
				client: &http.Client{},
			}
			got, err := a.processGetRequest(tt.args.path, tt.args.authToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("MonzoRestClient.processGetRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MonzoRestClient.processGetRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
