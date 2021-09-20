package paystack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

type APIClient struct {
	apiKey string
}

func NewAPIClient(apiKey string) *APIClient {
	return &APIClient{apiKey: apiKey}
}

const (
	resolveBankAccountURL = "https://api.paystack.co/bank/resolve?account_number=%s&bank_code=%s"
	authorizationHeader   = "Authorization"
)

func (a *APIClient) ResolveBankAccount(ctx context.Context, acccount *ResolveBankAccountRequest) (*Data, error) {
	url := fmt.Sprintf(resolveBankAccountURL, acccount.AccountNumber, acccount.BankCode)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add(authorizationHeader, "Bearer "+a.apiKey)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("resolve bank account failed: status code %d", resp.StatusCode)
	}

	responseData := &ResolveBankAccountResponse{}
	err = getResponseData(resp.Body, responseData)
	if err != nil {
		return nil, err
	}
	return responseData.Data, nil
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
