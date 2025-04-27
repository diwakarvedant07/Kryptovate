package models

import (
	"errors"
	"time"
	"github.com/google/uuid"
)

// ErrInsufficientFunds is returned when a debit transaction would result in a negative balance
var ErrInsufficientFunds = errors.New("insufficient funds")

// Transaction represents a financial transaction in the system
// @Description Transaction represents a credit or debit operation on a customer's account
type Transaction struct {
	TransactionID string    `json:"transaction_id" bson:"_id" example:"123e4567-e89b-12d3-a456-426614174000" description:"The unique identifier for the transaction"`
	CustomerID    string    `json:"customer_id" bson:"customer_id" example:"123e4567-e89b-12d3-a456-426614174000" description:"The ID of the customer"`
	Type          string    `json:"type" bson:"type" example:"credit" description:"The type of transaction (credit or debit)"`
	Amount        float64   `json:"amount" bson:"amount" example:"100.00" description:"The amount of the transaction"`
	Timestamp     time.Time `json:"timestamp" bson:"timestamp" example:"2025-04-06T10:45:00Z" description:"The timestamp of the transaction"`
}

// GenerateTransactionID generates a unique transaction ID
func GenerateTransactionID() string {
	return uuid.New().String()
}

// GenerateTimestamp generates the current timestamp
func GenerateTimestamp() time.Time {
	return time.Now()
}

// GenerateCustomerID generates a unique customer ID
func GenerateCustomerID() string {
	return uuid.New().String()
}

// Validate checks if the transaction is valid
func (t *Transaction) Validate() error {
	if t.TransactionID == "" {
		return errors.New("transaction ID is required")
	}
	if t.CustomerID == "" {
		return errors.New("customer ID is required")
	}
	if t.Type != "credit" && t.Type != "debit" {
		return errors.New("invalid transaction type")
	}
	if t.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	return nil
}

// CalculateNewBalance calculates the new balance after applying the transaction
func (t *Transaction) CalculateNewBalance(currentBalance float64) float64 {
	if t.Type == "credit" {
		return currentBalance + t.Amount
	}
	return currentBalance - t.Amount
} 