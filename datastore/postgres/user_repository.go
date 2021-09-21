package postgres

import (
	"context"
	"errors"
	app "github.com/danvixent/buycoin-challenge2"
	"gorm.io/gorm"
	"time"
)

type UserRepository struct {
	client *Client
}

func NewUserRepository(client *Client) app.UserRepository {
	return &UserRepository{client: client}
}

func (u *UserRepository) UpdateUser(ctx context.Context, user *app.User) error {
	user.UpdatedAt = time.Now()
	return u.client.db.Model(user).Updates(user).Error
}

func (u *UserRepository) FindUserByID(ctx context.Context, id string) (*app.User, error) {
	user := &app.User{ID: id}
	err := u.client.db.Model(user).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepository) CreateUser(ctx context.Context, user *app.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return u.client.db.Create(user).Error
}

func (u *UserRepository) SaveUserBankAccount(ctx context.Context, account *app.UserBankAccount) error {
	var count int64
	err := u.client.db.
		Model(account).
		Where("user_id = ?", account.UserID).
		Where("bank_account ->> 'user_bank_code' = ?", account.BankAccount.UserBankCode).
		Where("bank_account ->> 'user_account_number' = ?", account.BankAccount.UserAccountNumber).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("bank account already saved")
	}

	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()
	return u.client.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(account).Error
}

func (u *UserRepository) FindUserBankAccount(ctx context.Context, bankCode string, accountNumber string) (*app.UserBankAccount, error) {
	account := &app.UserBankAccount{}
	err := u.client.db.
		Model(account).
		Where("bank_account ->> 'user_bank_code' = ?", bankCode).
		Where("bank_account ->> 'user_account_number' = ?", accountNumber).
		First(account).Error
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (u *UserRepository) DeleteAllUsers() error {
	return u.client.db.Model(&app.User{}).Where("id IS NOT NULL").Delete("").Error
}

func (u *UserRepository) DeleteAllUserBankAccounts() error {
	return u.client.db.Model(&app.UserBankAccount{}).Where("id IS NOT NULL").Delete("").Error
}
