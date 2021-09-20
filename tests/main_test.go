package tests

import (
	"context"
	"fmt"
	"github.com/danvixent/buycoin-challenge2/config"
	"github.com/danvixent/buycoin-challenge2/datastore/postgres"
	"github.com/danvixent/buycoin-challenge2/graphql"
	"github.com/danvixent/buycoin-challenge2/handlers/account"
	"github.com/danvixent/buycoin-challenge2/providers/paystack"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

var baseURL = "http://localhost:%s/graphql"

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
	userRepo := postgres.NewUserRepository(postgresClient)
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

	time.Sleep(2 * time.Second)
	// run the tests
	code := m.Run()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so no need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	select {
	case <-ctx.Done():
		log.Print("timeout of 1 seconds.")
	}
	log.Print("server exiting")

	os.Exit(code)
}
