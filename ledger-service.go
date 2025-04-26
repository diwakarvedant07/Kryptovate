package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/joho/godotenv"
	"log"
)

type Account struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	Name    string            `json:"name" bson:"name"`
	Balance string            `json:"balance" bson:"balance"`
}

var client *mongo.Client
var accountsCollection *mongo.Collection

func main() {
	app := fiber.New()
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseURL := os.Getenv("MONGO_CLUSTER")

	clientOptions := options.Client().ApplyURI(databaseURL)
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get the accounts collection
	accountsCollection = client.Database("kryptovate").Collection("accounts")

	fmt.Println("Connected to MongoDB!")

	// Health Check Route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Get all accounts
	app.Get("/accounts", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := accountsCollection.Find(ctx, bson.M{})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to fetch accounts",
			})
		}
		defer cursor.Close(ctx)

		var accounts []Account
		if err = cursor.All(ctx, &accounts); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to decode accounts",
			})
		}

		return c.JSON(accounts)
	})

	// Start server
	app.Listen(":3005")
}
