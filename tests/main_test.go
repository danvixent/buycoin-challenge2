package tests

import (
	"context"
	"fmt"
	app "github.com/danvixent/buycoin-challenge2"
	"github.com/danvixent/buycoin-challenge2/config"
	"github.com/danvixent/buycoin-challenge2/datastore/postgres"
	"github.com/danvixent/buycoin-challenge2/graphql"
	"github.com/danvixent/buycoin-challenge2/handlers/account"
	"github.com/danvixent/buycoin-challenge2/providers/paystack"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	baseURL  = "http://localhost:%s/graphql"
	userRepo app.UserRepository
)

func TestMain(m *testing.M) {
	file, err := os.Open("../config/config.yml")
	if err != nil {
		log.Fatalf("unable to open config file: %v", err)
	}

	cfg := &config.BaseConfig{}
	err = yaml.NewDecoder(file).Decode(cfg)
	if err != nil {
		log.Fatalf("failed to decode config file: %v", err)
	}

	postgresClient := postgres.New(context.Background(), cfg.Postgres)
	userRepo = postgres.NewUserRepository(postgresClient)
	paystackClient := paystack.NewAPIClient(cfg.PaystackAPIKey)

	accountHandler := account.NewHandler(userRepo, paystackClient)
	graphqlHandler := graphql.NewHandler(accountHandler)

	mux := http.NewServeMux()
	graphqlHandler.SetupRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	baseURL = fmt.Sprintf(baseURL, port)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("serving graphql endpoint at http://localhost:%s/graphql", port)

	// start server in new goroutine so we can listen for CLI signals
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicf("unable to listen: %s", err)
		}
	}()

	// allow the goroutine above start the server
	time.Sleep(time.Second)

	// run the tests
	code := m.Run()

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("unable to shutdown server gracefully: %v", err)
	}

	os.Exit(code)
}

func deleteAllUsers() error {
	return userRepo.DeleteAllUsers()
}

func deleteAllUserBankAccounts() error {
	return userRepo.DeleteAllUserBankAccounts()
}
