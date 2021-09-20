package paystack

type ResolveBankAccountRequest struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
}

type ResolveBankAccountResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    *Data  `json:"data"`
}

type Data struct {
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	BankID        int    `json:"bank_id"`
}
