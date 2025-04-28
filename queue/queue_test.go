package queue

import (
	"ledger-service/models"
	"testing"
	"time"
)

func TestTransactionQueue(t *testing.T) {
	queue := NewTransactionQueue()

	// Test empty queue
	if !queue.IsEmpty() {
		t.Error("New queue should be empty")
	}

	// Test enqueue and dequeue
	t1 := models.Transaction{
		TransactionID: "test1",
		CustomerID:    "test_customer",
		Type:          "credit",
		Amount:        100,
		Timestamp:     time.Now(),
	}

	queue.Enqueue(t1)
	if queue.IsEmpty() {
		t.Error("Queue should not be empty after enqueue")
	}

	t2, ok := queue.Dequeue()
	if !ok {
		t.Error("Dequeue should succeed")
	}
	if t2.TransactionID != t1.TransactionID {
		t.Error("Dequeued transaction should match enqueued transaction")
	}
	if !queue.IsEmpty() {
		t.Error("Queue should be empty after dequeue")
	}
}

func TestTransactionQueueConcurrent(t *testing.T) {
	queue := NewTransactionQueue()
	done := make(chan bool)
	timeout := time.After(5 * time.Second) // Add timeout

	// Enqueue transactions concurrently
	for i := 0; i < 10; i++ {
		go func(i int) {
			transaction := models.Transaction{
				TransactionID: string(rune(i)),
				CustomerID:    "test_customer",
				Type:          "credit",
				Amount:        10,
				Timestamp:     time.Now(),
			}
			queue.Enqueue(transaction)
			done <- true
		}(i)
	}

	// Wait for all enqueues to complete or timeout
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			continue
		case <-timeout:
			t.Fatal("Test timed out waiting for concurrent operations")
		}
	}

	// Verify all transactions were enqueued
	count := 0
	for !queue.IsEmpty() {
		_, ok := queue.Dequeue()
		if !ok {
			t.Error("Dequeue should succeed")
		}
		count++
	}
	if count != 10 {
		t.Errorf("Expected 10 transactions, got %d", count)
	}
}
