//go:generate go run github.com/99designs/gqlgen --verbose

package graphql

import (
	"context"
	app "github.com/danvixent/buycoin-challenge2"
	"github.com/danvixent/buycoin-challenge2/handlers/account"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"
)

type Resolver struct {
	accountHandler *account.Handler
}

func (r *Resolver) User() UserResolver {
	return &userResolver{r}
}

type userResolver struct {
	*Resolver
}

func (u *userResolver) CreatedAt(ctx context.Context, obj *app.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.CreatedAt.Format(time.RFC3339), nil
}

func (u *userResolver) UpdatedAt(ctx context.Context, obj *app.User) (string, error) {
	if obj == nil {
		return "", nil
	}
	return obj.UpdatedAt.Format(time.RFC3339), nil
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct {
	*Resolver
}

func (q *queryResolver) ResolveAccount(ctx context.Context, bankCode string, accountNumber string) (string, error) {
	logger := log.WithField("bank_code", bankCode).WithField("account_number", accountNumber)
	logger.Info("resolve_account")

	if bankCode == "" {
		return "", errors.New("bank code is required")
	}

	if accountNumber == "" {
		return "", errors.New("bank code is required")
	}

	accountName, err := q.accountHandler.ResolveAccount(ctx, bankCode, accountNumber, logger)
	if err != nil {
		logger.Errorf("resolve account failed: %v", err)
		return "", err
	}
	return accountName, nil
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct {
	*Resolver
}

func (m *mutationResolver) RegisterUser(ctx context.Context, userDetails account.UserRegistrationVM) (*app.User, error) {
	if userDetails.Name == "" {
		return nil, errors.New("name is required")
	}

	if userDetails.Password == "" {
		return nil, errors.New("password is required")
	}

	if userDetails.Email == "" {
		return nil, errors.New("email is required")
	}

	logger := log.WithFields(map[string]interface{}{})
	user, err := m.accountHandler.RegisterUser(ctx, &userDetails, logger)
	if err != nil {
		logger.Errorf("register user failed: %v", err)
		return nil, err
	}

	return user, nil
}

func (m mutationResolver) AddBankAccount(ctx context.Context, userID string, input app.BankAccount) (bool, error) {
	if input.UserAccountName == "" {
		return false, errors.New("user_account_name is required")
	}

	if input.UserBankCode == "" {
		return false, errors.New("user_bank_code is required")
	}

	if input.UserAccountNumber == "" {
		return false, errors.New("user_account_number is required")
	}

	logger := log.WithFields(map[string]interface{}{})
	ok, err := m.accountHandler.AddBankAccount(ctx, userID, input, logger)
	if err != nil {
		logger.Errorf("add bank account failed: %v", err)
		return false, err
	}
	return ok, nil
}
