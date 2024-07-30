package model

type Transaction struct {
	TokenAddress string  `json:"token_address"`
	Action       string  `json:"action"`
	Amount       float64 `json:"amount"`
}
