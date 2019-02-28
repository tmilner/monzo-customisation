package main

import (
	"github.com/justinas/alice"
	"github.com/justinas/nosurf"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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
	errorChain := alice.New(loggerHandler, recoverHandler, timeoutHandler, nosurf.NewPure)

	mux := http.NewServeMux()
	mux.HandleFunc("/", genericIgnore)
	mux.HandleFunc("/webhook", api.WebhookHandler)
	mux.HandleFunc("/auth_return", api.AuthReturnHandler)
	mux.HandleFunc("/auth_start", api.AuthHandler)
	log.Println("Setting up webhook server")
	_ = http.ListenAndServe(":80", errorChain.Then(mux))
}

func genericIgnore(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("Dodgy request on /! Body failed to parse")

	} else {
		log.Printf("Dodgy request on /! Body: %s", string(body))
	}

	_, _ = io.WriteString(w, "Get off my server you prick. You wont find anything here.")
}

func loggerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("new request: %s %s %v (from: %s)", r.Method, r.URL.Path, time.Since(start), r.RemoteAddr)
	})
}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func timeoutHandler(h http.Handler) http.Handler {
	return http.TimeoutHandler(h, 1*time.Second, "timed out")
}
