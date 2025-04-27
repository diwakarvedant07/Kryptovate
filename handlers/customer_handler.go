package handlers

import (
	"context"
	"ledger-service/models"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CustomerHandler handles customer-related HTTP requests
type CustomerHandler struct {
	customersCollection    *mongo.Collection
	transactionsCollection *mongo.Collection
	mu                    sync.RWMutex
}

// NewCustomerHandler creates a new CustomerHandler
func NewCustomerHandler(customersCollection, transactionsCollection *mongo.Collection) *CustomerHandler {
	return &CustomerHandler{
		customersCollection:    customersCollection,
		transactionsCollection: transactionsCollection,
	}
}

// CreateCustomerRequest represents the request body for creating a customer
// @Description Request body for creating a new customer
type CreateCustomerRequest struct {
	Name    string   `json:"name" validate:"required" example:"John Doe" description:"The name of the customer"`
	Balance *float64 `json:"balance,omitempty" example:"100.50" description:"Initial balance (optional, defaults to 0)"`
}

// CreateCustomer handles the creation of a new customer
// @Summary Create a new customer
// @Description Creates a new customer with an optional initial balance (defaults to 0)
// @Tags customers
// @Accept json
// @Produce json
// @Param customer body CreateCustomerRequest true "Customer details"
// @Success 201 {object} models.Customer "Customer created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(c *fiber.Ctx) error {
	var req CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Set default balance to 0 if not provided
	initialBalance := 0.0
	if req.Balance != nil {
		initialBalance = *req.Balance
	}

	customer := models.Customer{
		CustomerID: models.GenerateCustomerID(),
		Name:      req.Name,
		Balance:   initialBalance,
	}

	_, err := h.customersCollection.InsertOne(context.Background(), customer)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to create customer",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(customer)
}

// GetBalance handles retrieving a customer's balance
// @Summary Get customer balance
// @Description Retrieves the current balance of a customer
// @Tags customers
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Success 200 {object} models.BalanceResponse "Customer balance retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 404 {object} models.ErrorResponse "Customer not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /customers/{customer_id}/balance [get]
func (h *CustomerHandler) GetBalance(c *fiber.Ctx) error {
	customerID := c.Params("customer_id")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Customer ID is required",
		})
	}

	var customer models.Customer
	err := h.customersCollection.FindOne(context.Background(), bson.M{"_id": customerID}).Decode(&customer)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Customer not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(models.BalanceResponse{
		CustomerID: customer.CustomerID,
		Balance:    customer.Balance,
	})
}

// TransactionHistoryResponse represents a transaction in the history
type TransactionHistoryResponse struct {
	TransactionID string  `json:"transaction_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Type         string  `json:"type" example:"credit"`
	Amount       float64 `json:"amount" example:"100.00"`
	Timestamp    string  `json:"timestamp" example:"2025-04-27T11:03:15Z"`
}

// GetTransactionHistory handles retrieving a customer's transaction history
// @Summary Get transaction history
// @Description Retrieves the transaction history for a customer
// @Tags customers
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Success 200 {array} TransactionHistoryResponse "Transaction history retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 404 {object} models.ErrorResponse "Customer not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /customers/{customer_id}/transactions [get]
func (h *CustomerHandler) GetTransactionHistory(c *fiber.Ctx) error {
	customerID := c.Params("customer_id")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Customer ID is required",
		})
	}

	var customer models.Customer
	err := h.customersCollection.FindOne(context.Background(), bson.M{"_id": customerID}).Decode(&customer)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Customer not found",
		})
	}

	cursor, err := h.transactionsCollection.Find(context.Background(), bson.M{"customer_id": customerID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to fetch transactions",
		})
	}
	defer cursor.Close(context.Background())

	var transactions []models.Transaction
	if err := cursor.All(context.Background(), &transactions); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to decode transactions",
		})
	}

	// Convert to response format without customer_id
	response := make([]TransactionHistoryResponse, len(transactions))
	for i, t := range transactions {
		response[i] = TransactionHistoryResponse{
			TransactionID: t.TransactionID,
			Type:         t.Type,
			Amount:       t.Amount,
			Timestamp:    t.Timestamp.Format(time.RFC3339),
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// RegisterRoutes registers the customer routes
func (h *CustomerHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/customers", h.CreateCustomer)
	app.Get("/customers/:customer_id/balance", h.GetBalance)
	app.Get("/customers/:customer_id/transactions", h.GetTransactionHistory)
} 