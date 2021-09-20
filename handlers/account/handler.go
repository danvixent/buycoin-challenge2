package account

import (
	"context"
	"github.com/agnivade/levenshtein"
	app "github.com/danvixent/buycoin-challenge2"
	"github.com/danvixent/buycoin-challenge2/password"
	"github.com/danvixent/buycoin-challenge2/providers/paystack"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	userRepo          app.UserRepository
	paystackAPIClient *paystack.APIClient
}

func NewHandler(userRepo app.UserRepository, paystackAPIClient *paystack.APIClient) *Handler {
	return &Handler{userRepo: userRepo, paystackAPIClient: paystackAPIClient}
}

func (h *Handler) RegisterUser(ctx context.Context, input *UserRegistrationVM, logger *log.Entry) (*app.User, error) {
	user := &app.User{
		Name:     input.Name,
		Email:    input.Email,
		Verified: false,
	}

	hash, err := password.NewPasswordHash(input.Password)
	if err != nil {
		logger.WithError(err).WithField("password_string", input.Password).Error("failed to generate password hash")
		return nil, errors.Wrap(err, "failed to generate password hash")
	}
	user.Password = hash

	err = h.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create user")
	}
	return user, nil
}

func (h *Handler) AddBankAccount(ctx context.Context, userID string, account app.BankAccount, logger *log.Entry) (bool, error) {
	user, err := h.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return false, errors.Wrap(err, "failed to find user by id")
	}

	r := &paystack.ResolveBankAccountRequest{
		AccountNumber: account.UserAccountNumber,
		BankCode:      account.UserBankCode,
	}

	data, err := h.paystackAPIClient.ResolveBankAccount(ctx, r)
	if err != nil {
		return false, errors.Wrap(err, "failed to resolve user bank account")
	}

	userBankAccount := &app.UserBankAccount{
		UserID:      user.ID,
		BankAccount: &account,
		User:        user,
	}

	if data.AccountName == account.UserAccountName {
		return h.verifyUser(ctx, userBankAccount)
	}

	distance := levenshtein.ComputeDistance(data.AccountName, account.UserAccountName)
	if distance <= 2 {
		return h.verifyUser(ctx, userBankAccount)
	}

	return false, errors.New("failed to add bank account")
}

func (h *Handler) verifyUser(ctx context.Context, userBankAccount *app.UserBankAccount) (bool, error) {
	userBankAccount.User.Verified = true
	err := h.userRepo.SaveUserBankAccount(ctx, userBankAccount)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (h *Handler) ResolveAccount(ctx context.Context, bankCode string, accountNumber string, logger *log.Entry) (string, error) {
	account, err := h.userRepo.FindUserBankAccount(ctx, bankCode, accountNumber)
	if err != nil {
		return "", errors.Wrap(err, "find user bank account failed")
	}

	return account.BankAccount.UserAccountName, nil
}
