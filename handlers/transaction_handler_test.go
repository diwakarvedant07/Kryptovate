package handlers

import (
	"context"
	"ledger-service/models"
	"ledger-service/queue"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) (*mongo.Collection, *mongo.Collection) {
	// Get the current directory and go back one level
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Go back one level to find the .env file
	envPath := filepath.Join(filepath.Dir(currentDir), ".env")

	if err := godotenv.Load(envPath); err != nil {
		t.Fatalf("Error loading .env file from %s: %v", envPath, err)
	}

	databaseURL := os.Getenv("MONGO_CLUSTER")
	if databaseURL == "" {
		t.Fatal("MONGO_CLUSTER environment variable is not set in .env file")
	}

	clientOptions := options.Client().ApplyURI(databaseURL)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Test the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Use a separate test database
	db := client.Database("kryptovate_test")
	customers := db.Collection("customers")
	transactions := db.Collection("transactions")

	// Clean up collections before tests
	_, err = customers.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("Failed to clean customers collection: %v", err)
	}
	_, err = transactions.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("Failed to clean transactions collection: %v", err)
	}

	return customers, transactions
}

func TestCreateTransaction(t *testing.T) {
	customers, transactions := setupTestDB(t)
	transactionQueue := queue.NewTransactionQueue()
	handler := NewTransactionHandler(transactionQueue, customers, transactions)

	// Create a test customer
	customer := models.Customer{
		CustomerID: "test_customer",
		Name:      "Test Customer",
		Balance:   1000,
	}
	_, err := customers.InsertOne(context.Background(), customer)
	if err != nil {
		t.Fatalf("Failed to create test customer: %v", err)
	}

	app := fiber.New()
	app.Post("/transactions", handler.CreateTransaction)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "valid credit transaction",
			requestBody:    `{"customer_id": "test_customer", "type": "credit", "amount": 100}`,
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "valid debit transaction",
			requestBody:    `{"customer_id": "test_customer", "type": "debit", "amount": 50}`,
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "invalid transaction type",
			requestBody:    `{"customer_id": "test_customer", "type": "invalid", "amount": 100}`,
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "negative amount",
			requestBody:    `{"customer_id": "test_customer", "type": "credit", "amount": -100}`,
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name:           "non-existent customer",
			requestBody:    `{"customer_id": "non_existent", "type": "credit", "amount": 100}`,
			expectedStatus: fiber.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/transactions", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
} 