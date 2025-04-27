package models

import (
	"testing"
	"time"
)

func TestTransactionValidation(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
	}{
		{
			name: "valid credit transaction",
			tx: Transaction{
				TransactionID: "test1",
				CustomerID:    "cust1",
				Type:         "credit",
				Amount:       100,
				Timestamp:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid debit transaction",
			tx: Transaction{
				TransactionID: "test2",
				CustomerID:    "cust1",
				Type:         "debit",
				Amount:       50,
				Timestamp:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid transaction type",
			tx: Transaction{
				TransactionID: "test3",
				CustomerID:    "cust1",
				Type:         "invalid",
				Amount:       100,
				Timestamp:    time.Now(),
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			tx: Transaction{
				TransactionID: "test4",
				CustomerID:    "cust1",
				Type:         "credit",
				Amount:       -100,
				Timestamp:    time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing transaction ID",
			tx: Transaction{
				CustomerID: "cust1",
				Type:      "credit",
				Amount:    100,
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing customer ID",
			tx: Transaction{
				TransactionID: "test5",
				Type:         "credit",
				Amount:       100,
				Timestamp:    time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Transaction.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateNewBalance(t *testing.T) {
	tests := []struct {
		name           string
		tx             Transaction
		currentBalance float64
		want           float64
	}{
		{
			name:           "credit transaction",
			tx:             Transaction{Type: "credit", Amount: 100},
			currentBalance: 500,
			want:           600,
		},
		{
			name:           "debit transaction",
			tx:             Transaction{Type: "debit", Amount: 100},
			currentBalance: 500,
			want:           400,
		},
		{
			name:           "zero balance credit",
			tx:             Transaction{Type: "credit", Amount: 100},
			currentBalance: 0,
			want:           100,
		},
		{
			name:           "zero balance debit",
			tx:             Transaction{Type: "debit", Amount: 100},
			currentBalance: 0,
			want:           -100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tx.CalculateNewBalance(tt.currentBalance)
			if got != tt.want {
				t.Errorf("Transaction.CalculateNewBalance() = %v, want %v", got, tt.want)
			}
		})
	}
} 