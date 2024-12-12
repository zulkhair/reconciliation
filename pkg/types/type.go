package types

import "time"

// TransactionType is the type of the transaction
type TransactionType string

const (
	// Enum for transaction type
	TransactionTypeDebit  TransactionType = "DEBIT"
	TransactionTypeCredit TransactionType = "CREDIT"
)

// Transaction is a transaction from the system
type Transaction struct {
	// Unique identifier for the transaction
	TrxID string

	// Transaction amount
	// Assume the format is 1234.56
	Amount float64

	// Transaction type
	// DEBIT or CREDIT
	Type TransactionType

	// Date and time of the transaction
	// Assume the format is YYYY-MM-DD HH:MM:SS
	TransactionTime time.Time
}

// BankStatement is a bank statement
type BankStatement struct {
	// Bank name
	// Assume the name is parsed from file name
	BankName string

	// Unique identifier for the bank statement
	UniqueID string

	// Transaction amount
	// Assume the format is 1234.56
	Amount float64

	// Date of the transaction
	// Assume the format is YYYY-MM-DD
	Date time.Time
}
