package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/joho/godotenv"
	"log"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"ledger-service/handlers"
	"ledger-service/queue"
	_ "ledger-service/docs" // This is required for swagger

	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title           Ledger Service API
// @version         1.0
// @description     This is a ledger service API for managing customers and transactions
// @termsOfService  http://swagger.io/terms/
// @contact.name    DEV TEAM
// @contact.email   dev@kryptovate.com
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:3005
// @BasePath  /
// @schemes   http

func main() {
	app := fiber.New()
	 app.Use(cors.New(cors.Config{
			AllowOrigins: "*", // For testing, be more specific in production
			AllowMethods: "GET,POST,PUT,DELETE",
			AllowHeaders: "Origin, Content-Type, Accept",
	}))
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseURL := os.Getenv("MONGO_CLUSTER")

	clientOptions := options.Client().ApplyURI(databaseURL)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Test the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get collections
	customersCollection := client.Database("kryptovate").Collection("customers")
	transactionsCollection := client.Database("kryptovate").Collection("transactions")

	// Initialize transaction queue
	transactionQueue := queue.NewTransactionQueue()

	// Initialize route handlers
	customersHandler := handlers.NewCustomerHandler(customersCollection, transactionsCollection)
	transactionsHandler := handlers.NewTransactionHandler(transactionQueue, customersCollection, transactionsCollection)

	// Swagger configuration
	// app.Get("/swagger/*", swagger.New(swagger.Config{
	// 	URL:         "/swagger/doc.json",
	// 	DeepLinking: true,
	// }))

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Register routes
	customersHandler.RegisterRoutes(app)
	transactionsHandler.RegisterRoutes(app)

	// Health Check Route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	fmt.Println("Connected to MongoDB!")
	app.Listen(":3005")
}
