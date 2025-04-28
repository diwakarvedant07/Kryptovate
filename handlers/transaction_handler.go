package handlers

import (
	"ledger-service/models"
	"ledger-service/queue"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TransactionHandler handles transaction-related requests
type TransactionHandler struct {
	queue                  *queue.TransactionQueue
	customersCollection    *mongo.Collection
	transactionsCollection *mongo.Collection
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(
	queue *queue.TransactionQueue,
	customersCollection *mongo.Collection,
	transactionsCollection *mongo.Collection,
) *TransactionHandler {
	return &TransactionHandler{
		queue:                  queue,
		customersCollection:    customersCollection,
		transactionsCollection: transactionsCollection,
	}
}

// CreateTransactionRequest represents the request body for creating a transaction
type CreateTransactionRequest struct {
	CustomerID string  `json:"customer_id" example:"ef48ae68-182f-4f2f-bb62-8a0016a9ca94"`
	Type       string  `json:"type" example:"credit"`
	Amount     float64 `json:"amount" example:"100"`
}

// CreateTransaction handles the creation of a new transaction
// @Summary Create a new transaction
// @Description Creates a new credit or debit transaction for a customer
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body CreateTransactionRequest true "Transaction details"
// @Success 200 {object} models.TransactionStatusResponse "Transaction processed successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 404 {object} models.ErrorResponse "Customer not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /transactions [post]
func (h *TransactionHandler) CreateTransaction(c *fiber.Ctx) error {
	var req CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: err.Error()})
	}

	// Validate transaction type
	if req.Type != "credit" && req.Type != "debit" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Invalid transaction type. Must be 'credit' or 'debit'"})
	}

	// Validate amount
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "Amount must be greater than 0"})
	}

	// Check if customer exists
	var customer models.Customer
	err := h.customersCollection.FindOne(c.Context(), bson.M{"_id": req.CustomerID}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "Customer not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "Failed to check customer existence"})
	}

	// Create transaction with generated ID and timestamp
	transaction := models.Transaction{
		TransactionID: models.GenerateTransactionID(),
		CustomerID:    req.CustomerID,
		Type:          req.Type,
		Amount:        req.Amount,
		Timestamp:     models.GenerateTimestamp(),
	}

	// Create a worker for this customer if one doesn't exist
	worker := queue.NewWorker(
		transaction.CustomerID,
		h.queue,
		h.customersCollection,
		h.transactionsCollection,
	)
	worker.Start()
	defer worker.Stop()

	// Enqueue the transaction
	h.queue.Enqueue(transaction)

	// Wait for transaction completion with timeout
	select {
	case status := <-worker.GetCompletionChan():
		return c.Status(fiber.StatusOK).JSON(status)
	case <-time.After(30 * time.Second):
		return c.Status(fiber.StatusRequestTimeout).JSON(models.ErrorResponse{Error: "Transaction processing timed out"})
	}
}

// RegisterRoutes registers the transaction routes
func (h *TransactionHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/transactions", h.CreateTransaction)
}
