package reconciliation

import "time"

type TransactionType string

const (
	Debit  TransactionType = "DEBIT"
	Credit TransactionType = "CREDIT"
)

// Money represents currency in cents to avoid float errors
type Money int64

func ToMoney(amount float64) Money {
	return Money(amount * 100)
}

func (m Money) ToFloat() float64 {
	return float64(m) / 100.0
}

type SystemTransaction struct {
	TransactionID   string
	Amount          Money
	Type            TransactionType
	TransactionTime time.Time
}

type BankTransaction struct {
	BankName string
	UniqueID string
	Amount   Money
	Date     time.Time
}

type ReconciliationResult struct {
	TotalProcessed     int
	TotalMatched       int
	TotalUnmatched     int
	TotalDiscrepancies Money
	UnmatchedSystem    []SystemTransaction
	UnmatchedBank      map[string][]BankTransaction
}
