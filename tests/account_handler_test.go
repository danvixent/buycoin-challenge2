package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	app "github.com/danvixent/buycoin-challenge2"
	"github.com/danvixent/buycoin-challenge2/password"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	err := deleteAllUserBankAccounts()
	if !assert.NoError(t, err) {
		return
	}

	err = deleteAllUsers()
	if !assert.NoError(t, err) {
		return
	}

	tests := []struct {
		name         string
		wantCode     int
		gqlQuery     string
		wantErr      bool
		checkData    bool
		errorMessage string
	}{
		{
			name:      "should_register_user_for_correct_input",
			checkData: true,
			wantCode:  http.StatusOK,
			gqlQuery: `
					mutation{
  						registerUser(userDetails: {
    						name:"Daniel"
							password:"dfanvixent@gmail.com"
    						email:"123456"
  						}){
    						id
    						verified
    						email
    						name
    						created_at
    						updated_at
  						}
					}`,
			wantErr:      false,
			errorMessage: "",
		},
		{
			name: "should_error_for_wrong_input",
			gqlQuery: `
					mutation{
  						registerUser(userDetails: {
    						name:"Daniel"
    						email:"dan@gmail.com"
  						}){
    						id
    						verified
    						email
    						name
    						created_at
    						updated_at
  						}
					}`,
			checkData:    false,
			wantCode:     http.StatusUnprocessableEntity,
			wantErr:      true,
			errorMessage: "Field UserRegistrationInput.password of required type String! was not provided.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gql := graphql.RawParams{Query: tt.gqlQuery}

			resp, err := sendRequest(gql)
			if err != nil {
				t.Errorf("sendRequest() error = %v", err)
				return
			}

			if !assert.Equal(t, tt.wantCode, resp.StatusCode) {
				return
			}

			body := &struct {
				Errors []struct{ Message string }
				Data   struct{ RegisterUser *app.User }
			}{}

			err = getResponseData(resp.Body, body)
			if !assert.NoError(t, err) {
				return
			}

			fmt.Printf("body %+v\n", body.Data)
			if tt.checkData {
				assert.NotNil(t, body.Data)
				assert.NotEmpty(t, body.Data.RegisterUser.ID)
				assert.NotEmpty(t, body.Data.RegisterUser.Name)
				assert.NotEmpty(t, body.Data.RegisterUser.Email)
			}

			if tt.wantErr {
				assert.Equal(t, tt.errorMessage, body.Errors[0].Message)
			}
		})
	}
}

func TestAddBankAccount(t *testing.T) {
	err := deleteAllUserBankAccounts()
	if !assert.NoError(t, err) {
		return
	}

	err = deleteAllUsers()
	if !assert.NoError(t, err) {
		return
	}

	// seed one user
	user := &app.User{
		Email:    "dan@gmail.live",
		Name:     "Daniel",
		Password: generateHash("password"),
	}
	err = userRepo.CreateUser(context.Background(), user)
	if !assert.NoError(t, err) {
		return
	}

	tests := []struct {
		wantCode     int
		name         string
		gqlQuery     string
		wantErr      bool
		checkData    bool
		errorMessage string
	}{
		{
			name:      "should_register_user_for_correct_input",
			checkData: true,
			wantCode:  http.StatusOK,
			gqlQuery: `
					mutation{
  						addBankAccount(user_id:"%s"
  						input:{
    						user_bank_code:"035"
    						user_account_name:"Daniel Oluojomu"
    						user_account_number:"7811035835"
						})
					}`,
			wantErr:      false,
			errorMessage: "",
		},
		{
			name:      "should_error_for_duplicate_account",
			checkData: false,
			wantCode:  http.StatusOK,
			gqlQuery: `
					mutation{
  						addBankAccount(user_id:"%s"
  						input:{
    						user_bank_code:"035"
    						user_account_name:"Daniel Oluojomu"
    						user_account_number:"7811035835"
						})
					}`,
			wantErr:      true,
			errorMessage: "bank account already saved",
		},
		{
			name: "should_error_for_wrong_input",
			gqlQuery: `
					mutation{
  						addBankAccount(user_id:"%s"
  						input:{
    						user_bank_code:"030"
    						user_account_name:"Daniel Oluojomu"
    						user_account_number:"7811035835"
						})
					}`,
			checkData:    false,
			wantCode:     http.StatusOK,
			wantErr:      true,
			errorMessage: "failed to resolve user bank account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := fmt.Sprintf(tt.gqlQuery, user.ID)
			gql := graphql.RawParams{Query: query}

			resp, err := sendRequest(gql)
			if err != nil {
				t.Errorf("sendRequest() error = %v", err)
				return
			}

			if !assert.Equal(t, tt.wantCode, resp.StatusCode) {
				return
			}

			body := &struct {
				Errors []struct{ Message string }
				Data   struct{ AddbankAccount bool }
			}{}

			err = getResponseData(resp.Body, body)
			if !assert.NoError(t, err) {
				return
			}

			fmt.Printf("body %+v\n", body.Data)
			if tt.checkData {
				assert.NotNil(t, body.Data)
				assert.True(t, body.Data.AddbankAccount)
			}

			if tt.wantErr {
				assert.Equal(t, tt.errorMessage, body.Errors[0].Message)
			}
		})
	}
}

func TestResolveAccount(t *testing.T) {
	err := deleteAllUserBankAccounts()
	if !assert.NoError(t, err) {
		return
	}

	err = deleteAllUsers()
	if !assert.NoError(t, err) {
		return
	}

	// seed one user
	user := &app.User{
		Email:    "danb@gmail.live",
		Name:     "Daniel",
		Password: generateHash("password"),
	}
	err = userRepo.CreateUser(context.Background(), user)
	if !assert.NoError(t, err) {
		return
	}

	account := &app.UserBankAccount{
		UserID: user.ID,
		User:   user,
		BankAccount: &app.BankAccount{
			UserAccountNumber: "7811035835",
			UserBankCode:      "035",
			UserAccountName:   "Daniel Oluojomu",
		},
	}

	err = userRepo.SaveUserBankAccount(context.Background(), account)
	if !assert.NoError(t, err) {
		return
	}

	tests := []struct {
		wantCode        int
		name            string
		gqlQuery        string
		wantErr         bool
		wantAccountName string
		checkData       bool
		errorMessage    string
	}{
		{
			name: "should_resolve_correctly",
			gqlQuery: `
					query{
  						resolveAccount(
    						bank_code:"035"
    						account_number:"7811035835"
  						)
					}`,
			wantErr:         false,
			checkData:       true,
			wantCode:        http.StatusOK,
			wantAccountName: "Daniel Oluojomu",
			errorMessage:    "",
		},
		{
			name:      "should_error_for_incorrect_input",
			checkData: false,
			wantCode:  http.StatusOK,
			gqlQuery: `
					query{
  						resolveAccount(
    						bank_code:"02"
    						account_number:"7811035835"
  						)
					}`,
			wantErr:         true,
			wantAccountName: "",
			errorMessage:    "find user bank account failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gql := graphql.RawParams{Query: tt.gqlQuery}

			resp, err := sendRequest(gql)
			if err != nil {
				t.Errorf("sendRequest() error = %v", err)
				return
			}

			body := &struct {
				Errors []struct{ Message string }
				Data   struct{ ResolveAccount string }
			}{}

			err = getResponseData(resp.Body, body)
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.wantCode, resp.StatusCode)
			if tt.checkData {
				assert.NotNil(t, body.Data)
				assert.Equal(t, tt.wantAccountName, body.Data.ResolveAccount)
			}

			if tt.wantErr {
				assert.Equal(t, tt.errorMessage, body.Errors[0].Message)
			}
		})
	}
}

func generateHash(s string) *password.Hash {
	hash, err := password.NewPasswordHash(s)
	if err != nil {
		log.Panicf("failed to generate hash: %v", err)
	}
	return hash
}

func getResponseData(respBody io.ReadCloser, data interface{}) error {
	responseData, err := ioutil.ReadAll(respBody)
	if err != nil {
		return err
	}
	fmt.Println("resp", string(responseData))
	if err = json.Unmarshal(responseData, data); err != nil {
		return err
	}

	return nil
}

func sendRequest(body interface{}) (*http.Response, error) {
	return POST(baseURL, serialize(body))
}

// serialize obj into json bytes
func serialize(obj interface{}) *bytes.Buffer {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(obj); err != nil {
		// if json encoding fails, stop the test immediately
		log.Fatalf("unable to serialize obj: %v", err)
	}
	return buf
}

func POST(url string, body *bytes.Buffer) (*http.Response, error) {
	return http.Post(url, "application/json", body)
}
