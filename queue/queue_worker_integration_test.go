package queue

import (
	"context"
	"ledger-service/models"
	"os"
	"path/filepath"
	"testing"
	"time"

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

	return customers, transactions
}

func TestQueueAndWorkerIntegration(t *testing.T) {
	customers, transactions := setupTestDB(t)
	queue := NewTransactionQueue()

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

	// Create a worker
	worker := NewWorker("test_customer", queue, customers, transactions)
	worker.Start()
	defer worker.Stop()

	// Test credit transaction
	creditTx := models.Transaction{
		TransactionID: models.GenerateTransactionID(),
		CustomerID:    "test_customer",
		Type:         "credit",
		Amount:       100,
		Timestamp:    time.Now(),
	}

	queue.Enqueue(creditTx)

	// Wait for transaction completion
	select {
	case status := <-worker.GetCompletionChan():
		if status.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", status.Status)
		}
		if status.Balance != 1100 {
			t.Errorf("Expected balance 1100, got %f", status.Balance)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Transaction processing timed out")
	}

	// Test debit transaction
	debitTx := models.Transaction{
		TransactionID: models.GenerateTransactionID(),
		CustomerID:    "test_customer",
		Type:         "debit",
		Amount:       50,
		Timestamp:    time.Now(),
	}

	queue.Enqueue(debitTx)

	// Wait for transaction completion
	select {
	case status := <-worker.GetCompletionChan():
		if status.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", status.Status)
		}
		if status.Balance != 1050 {
			t.Errorf("Expected balance 1050, got %f", status.Balance)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Transaction processing timed out")
	}

	// Test insufficient funds
	insufficientTx := models.Transaction{
		TransactionID: models.GenerateTransactionID(),
		CustomerID:    "test_customer",
		Type:         "debit",
		Amount:       2000,
		Timestamp:    time.Now(),
	}

	queue.Enqueue(insufficientTx)

	// Wait for transaction completion
	select {
	case status := <-worker.GetCompletionChan():
		if status.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", status.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Transaction processing timed out")
	}

	// Verify final balance
	var updatedCustomer models.Customer
	err = customers.FindOne(context.Background(), bson.M{"_id": "test_customer"}).Decode(&updatedCustomer)
	if err != nil {
		t.Fatalf("Failed to get updated customer: %v", err)
	}
	if updatedCustomer.Balance != 1050 {
		t.Errorf("Expected final balance 1050, got %f", updatedCustomer.Balance)
	}

	// Verify transaction history
	cursor, err := transactions.Find(context.Background(), bson.M{"customer_id": "test_customer"})
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}
	defer cursor.Close(context.Background())

	var txHistory []models.Transaction
	if err := cursor.All(context.Background(), &txHistory); err != nil {
		t.Fatalf("Failed to decode transactions: %v", err)
	}

	if len(txHistory) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txHistory))
	}
} 