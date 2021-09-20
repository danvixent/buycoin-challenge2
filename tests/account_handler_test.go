package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	app "github.com/danvixent/buycoin-challenge2"
	"github.com/danvixent/buycoin-challenge2/handlers/account"
)

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name         string
		input        *account.UserRegistrationVM
		wantCode     int
		wantErr      bool
		checkData    bool
		errorMessage string
	}{
		{
			name: "should_register_user_correct_input",
			input: &account.UserRegistrationVM{
				Name:     "Daniel",
				Email:    "danvixent@gmail.com",
				Password: "123456",
			},
			checkData:    true,
			wantCode:     http.StatusOK,
			wantErr:      false,
			errorMessage: "",
		},
		{
			name: "should_error_for_wrong_input",
			input: &account.UserRegistrationVM{
				Name:     "Daniel",
				Email:    "danvixent@gmail.com",
				Password: "",
			},
			checkData:    false,
			wantCode:     http.StatusUnprocessableEntity,
			wantErr:      true,
			errorMessage: "Field UserRegistrationInput.password of required type String! was not provided.",
		},
	}

	const baseQuery = `
	mutation{
  		registerUser(userDetails: {
    		name:%s 
			password:%s"
    		email:%s
  		}){
    		id
    		verified
    		email
    		name
    		created_at
    		updated_at
  		}
	}`
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			query := fmt.Sprintf(baseQuery, tt.input.Name, tt.input.Password, tt.input.Email)
			gql := graphql.Mutation{Query: query}

			resp, err := sendRequest(gql)
			if err != nil {
				t.Errorf("sendRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp.StatusCode != tt.wantCode {
				t.Errorf("sendRequest() status code = %v, want %v", resp.StatusCode, tt.wantCode)
				return
			}

			body := &struct {
				errors []string
				data   *app.User
			}{}

			if err = getResponseData(resp.Body, body); err != nil {
				t.Errorf("getResponseData failed: %v", err)
				return
			}
		})
	}
}

func getResponseData(respBody io.ReadCloser, data interface{}) error {
	responseData, err := ioutil.ReadAll(respBody)
	if err != nil {
		return err
	}

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
