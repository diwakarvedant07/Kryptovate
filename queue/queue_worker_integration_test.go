package queue

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
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
