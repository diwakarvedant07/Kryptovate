package models

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error message"`
}

// BalanceResponse represents a balance response
type BalanceResponse struct {
	CustomerID string  `json:"customer_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Balance    float64 `json:"balance" example:"100.50"`
}

// TransactionResponse represents a transaction response
type TransactionResponse struct {
	Message     string `json:"message" example:"Transaction queued successfully"`
	Transaction struct {
		TransactionID string  `json:"transaction_id" example:"123e4567-e89b-12d3-a456-426614174000"`
		CustomerID    string  `json:"customer_id" example:"123e4567-e89b-12d3-a456-426614174000"`
		Type         string  `json:"type" example:"credit"`
		Amount       float64 `json:"amount" example:"100.00"`
		Timestamp    string  `json:"timestamp" example:"2025-04-27T11:03:15Z"`
	} `json:"transaction"`
}

// TransactionStatusResponse represents the status of a completed transaction
type TransactionStatusResponse struct {
	TransactionID string  `json:"transaction_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Status       string  `json:"status" example:"completed"`
	Balance      float64 `json:"balance" example:"100.50"`
} 