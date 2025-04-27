package queue

import (
	"ledger-service/models"
	"sync"
)

// TransactionQueue represents a queue of transactions
type TransactionQueue struct {
	transactions []models.Transaction
	mu          sync.Mutex
}

// NewTransactionQueue creates a new transaction queue
func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{}
}

// Enqueue adds a transaction to the queue
func (q *TransactionQueue) Enqueue(t models.Transaction) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.transactions = append(q.transactions, t)
}

// Dequeue removes and returns the first transaction from the queue
func (q *TransactionQueue) Dequeue() (models.Transaction, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.transactions) == 0 {
		return models.Transaction{}, false
	}

	t := q.transactions[0]
	q.transactions = q.transactions[1:]
	return t, true
}

// IsEmpty checks if the queue is empty
func (q *TransactionQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.transactions) == 0
} 