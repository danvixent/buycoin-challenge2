package buycoin_challenge2

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/danvixent/buycoin-challenge2/password"
	"github.com/pkg/errors"
)

type User struct {
	ID        string         `json:"id" gorm:"default:gen_random_uuid()"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	Password  *password.Hash `json:"password"`
	Verified  bool           `json:"verified"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt *time.Time     `json:"deleted_at"`
}

type EmailAddress string

// String implements Stringer and makes sure email addresses are canonicalized
func (e EmailAddress) String() string {
	return strings.ToLower(string(e))
}

// MarshalGQL implements the graphql.Marshaler interface
func (e EmailAddress) MarshalGQL(w io.Writer) {
	w.Write([]byte(e.String()))
}

// UnmarshalGQL implements the graphql.UnMarshaler interface
func (e *EmailAddress) UnmarshalGQL(v interface{}) error {
	val, ok := v.(string)
	if !ok {
		return errors.New("email address must be a string")
	}

	*e = EmailAddress(strings.ToLower(val))
	return nil
}

type UserBankAccount struct {
	ID          string       `json:"id" gorm:" default:gen_random_uuid()"`
	UserID      string       `json:"user_id"`
	User        *User        `json:"user"`
	BankAccount *BankAccount `json:"bank_account"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at"`
}

// Value get value of Jsonb
func (a *BankAccount) Value() (driver.Value, error) {
	j, err := json.Marshal(a)
	return j, err
}

// Scan scan value into Hash
func (a *BankAccount) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, a)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	FindUserByID(ctx context.Context, id string) (*User, error)

	SaveUserBankAccount(ctx context.Context, account *UserBankAccount) error
	FindUserBankAccount(ctx context.Context, bankCode string, accountNumber string) (*UserBankAccount, error)
}
