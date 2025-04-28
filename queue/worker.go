package queue

import (
	"context"
	"ledger-service/models"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Worker processes transactions for a specific customer
type Worker struct {
	customerID             string
	queue                  *TransactionQueue
	customersCollection    *mongo.Collection
	transactionsCollection *mongo.Collection
	stopChan               chan struct{}
	completionChan         chan models.TransactionStatusResponse
	mu                     sync.RWMutex
	stopped                bool
}

// NewWorker creates a new worker for a specific customer
func NewWorker(
	customerID string,
	queue *TransactionQueue,
	customersCollection *mongo.Collection,
	transactionsCollection *mongo.Collection,
) *Worker {
	return &Worker{
		customerID:             customerID,
		queue:                  queue,
		customersCollection:    customersCollection,
		transactionsCollection: transactionsCollection,
		stopChan:               make(chan struct{}),
		completionChan:         make(chan models.TransactionStatusResponse, 100),
	}
}

// Start begins processing transactions for the customer
func (w *Worker) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.stopped {
		go w.processTransactions()
	}
}

// Stop signals the worker to stop processing
func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.stopped {
		close(w.stopChan)
		w.stopped = true
	}
}

// GetCompletionChan returns the channel for transaction completion notifications
func (w *Worker) GetCompletionChan() <-chan models.TransactionStatusResponse {
	return w.completionChan
}

func (w *Worker) processTransactions() {
	for {
		select {
		case <-w.stopChan:
			return
		default:
			if w.queue.IsEmpty() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if t, ok := w.queue.Dequeue(); ok {
				w.processTransaction(t)
			}
		}
	}
}

func (w *Worker) processTransaction(t models.Transaction) {
	// Check for nil collections
	if w.customersCollection == nil || w.transactionsCollection == nil {
		w.completionChan <- models.TransactionStatusResponse{
			TransactionID: t.TransactionID,
			Status:        "failed",
			Balance:       0,
		}
		return
	}

	// Validate transaction
	if t.Type != "credit" && t.Type != "debit" {
		w.completionChan <- models.TransactionStatusResponse{
			TransactionID: t.TransactionID,
			Status:        "failed",
			Balance:       0,
		}
		return
	}

	if t.Amount <= 0 {
		w.completionChan <- models.TransactionStatusResponse{
			TransactionID: t.TransactionID,
			Status:        "failed",
			Balance:       0,
		}
		return
	}

	// Start MongoDB session
	session, err := w.customersCollection.Database().Client().StartSession()
	if err != nil {
		w.completionChan <- models.TransactionStatusResponse{
			TransactionID: t.TransactionID,
			Status:        "failed",
			Balance:       0,
		}
		return
	}
	defer session.EndSession(context.Background())

	var updatedBalance float64
	// Process transaction in a session
	_, err = session.WithTransaction(context.Background(), func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Get current customer
		var customer models.Customer
		err := w.customersCollection.FindOne(sessCtx, bson.M{"_id": t.CustomerID}).Decode(&customer)
		if err != nil {
			return nil, err
		}

		// Check for insufficient funds before updating balance
		if t.Type == "debit" && customer.Balance < t.Amount {
			return nil, models.ErrInsufficientFunds
		}

		// Update balance
		if t.Type == "credit" {
			customer.Balance += t.Amount
		} else {
			customer.Balance -= t.Amount
		}

		// Update customer
		_, err = w.customersCollection.UpdateOne(
			sessCtx,
			bson.M{"_id": t.CustomerID},
			bson.M{"$set": bson.M{"balance": customer.Balance}},
		)
		if err != nil {
			return nil, err
		}

		// Store the updated balance
		updatedBalance = customer.Balance

		// Insert transaction
		_, err = w.transactionsCollection.InsertOne(sessCtx, t)
		return nil, err
	})

	if err != nil {
		if err == models.ErrInsufficientFunds {
			w.completionChan <- models.TransactionStatusResponse{
				TransactionID: t.TransactionID,
				Status:        "failed",
				Balance:       0,
			}
			return
		}
		w.queue.Enqueue(t)
		w.completionChan <- models.TransactionStatusResponse{
			TransactionID: t.TransactionID,
			Status:        "failed",
			Balance:       0,
		}
		return
	}

	w.completionChan <- models.TransactionStatusResponse{
		TransactionID: t.TransactionID,
		Status:        "completed",
		Balance:       updatedBalance,
	}
}
