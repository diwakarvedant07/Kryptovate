package queue

import (
	"context"
	"ledger-service/models"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mockCollection implements mongo.Collection interface for testing
type mockCollection struct {
	findOneResult   interface{}
	findOneErr      error
	updateOneResult *mongo.UpdateResult
	updateOneErr    error
	insertOneResult *mongo.InsertOneResult
	insertOneErr    error
	countResult     int64
	countErr        error
}

func (m *mockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return mongo.NewSingleResultFromDocument(m.findOneResult, m.findOneErr, nil)
}

func (m *mockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.updateOneResult, m.updateOneErr
}

func (m *mockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.insertOneResult, m.insertOneErr
}

func (m *mockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return m.countResult, m.countErr
}

// Required interface methods that we don't need to implement for our tests
func (m *mockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return nil, nil
}

func (m *mockCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	return nil
}

func (m *mockCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	return nil
}

func (m *mockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return nil, nil
}

func (m *mockCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return nil, nil
}

func (m *mockCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return nil, nil
}

func (m *mockCollection) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	return nil, nil
}

func (m *mockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	return nil, nil
}

func (m *mockCollection) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return nil, nil
}

func (m *mockCollection) Indexes() mongo.IndexView {
	return mongo.IndexView{}
}

func (m *mockCollection) Drop(ctx context.Context) error {
	return nil
}

func (m *mockCollection) Name() string {
	return "mock_collection"
}

func (m *mockCollection) Database() *mongo.Database {
	return nil
}

func TestWorkerQueueOperations(t *testing.T) {
	queue := NewTransactionQueue()
	worker := NewWorker("test_customer", queue, nil, nil)
	worker.Start()
	defer worker.Stop()

	// Test credit transaction
	t1 := models.Transaction{
		TransactionID: "test1",
		CustomerID:    "test_customer",
		Type:          "credit",
		Amount:        100,
		Timestamp:     time.Now(),
	}

	queue.Enqueue(t1)

	select {
	case status := <-worker.GetCompletionChan():
		if status.Status != "failed" {
			t.Error("Transaction should fail without database connection")
		}
	case <-time.After(1 * time.Second):
		t.Error("Transaction processing timed out")
	}
}

func TestWorkerLifecycle(t *testing.T) {
	queue := NewTransactionQueue()
	worker := NewWorker("test_customer", queue, nil, nil)
	worker.Start()
	defer worker.Stop()

	// Test worker stop
	worker.Stop()
	time.Sleep(100 * time.Millisecond) // Give worker time to stop

	// Verify worker is stopped
	t1 := models.Transaction{
		TransactionID: "test1",
		CustomerID:    "test_customer",
		Type:          "credit",
		Amount:        100,
		Timestamp:     time.Now(),
	}

	queue.Enqueue(t1)
	select {
	case <-worker.GetCompletionChan():
		t.Error("Worker should not process transactions after stop")
	case <-time.After(1 * time.Second):
		// Expected: no response from stopped worker
	}
}

func TestWorkerTransactionValidation(t *testing.T) {
	queue := NewTransactionQueue()
	worker := NewWorker("test_customer", queue, nil, nil)
	worker.Start()
	defer worker.Stop()

	// Test invalid transaction type
	t1 := models.Transaction{
		TransactionID: "test1",
		CustomerID:    "test_customer",
		Type:          "invalid",
		Amount:        100,
		Timestamp:     time.Now(),
	}

	queue.Enqueue(t1)

	select {
	case status := <-worker.GetCompletionChan():
		if status.Status != "failed" {
			t.Error("Transaction with invalid type should fail")
		}
	case <-time.After(1 * time.Second):
		t.Error("Transaction processing timed out")
	}

	// Test negative amount
	t2 := models.Transaction{
		TransactionID: "test2",
		CustomerID:    "test_customer",
		Type:          "credit",
		Amount:        -100,
		Timestamp:     time.Now(),
	}

	queue.Enqueue(t2)

	select {
	case status := <-worker.GetCompletionChan():
		if status.Status != "failed" {
			t.Error("Transaction with negative amount should fail")
		}
	case <-time.After(1 * time.Second):
		t.Error("Transaction processing timed out")
	}
}
