package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/danvixent/buycoin-challenge2/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Client struct {
	db *gorm.DB
}

// New is a postgres database constructor
func New(
	ctx context.Context,
	cfg *config.PostgresConfig,
) *Client {

	// use username/password credentials if URI which should contain everything is not given
	uri := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		log.Panicf("Creating postgres connection, err=%v", err)
	}

	log.Println("Connected to postgres")

	return &Client{db: db}
}
