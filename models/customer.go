package models

// Customer represents a financial account in the system
// @Description Customer represents a financial account that can hold balance and perform transactions
type Customer struct {
	CustomerID string  `json:"customer_id" bson:"_id" example:"123e4567-e89b-12d3-a456-426614174000" description:"The unique identifier for the customer"`
	Name       string  `json:"name" bson:"name" example:"John Doe" description:"The name of the customer"`
	Balance    float64 `json:"balance" bson:"balance" example:"1000.00" description:"The current balance of the customer"`
} 