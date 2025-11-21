package reconciliation

import (
	"bytes"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToMoney(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected Money
	}{
		{
			name:     "positive amount",
			input:    100.50,
			expected: 10050,
		},
		{
			name:     "negative amount",
			input:    -50.25,
			expected: -5025,
		},
		{
			name:     "zero amount",
			input:    0.0,
			expected: 0,
		},
		{
			name:     "small decimal",
			input:    0.01,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToMoney(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoneyToFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    Money
		expected float64
	}{
		{
			name:     "positive money",
			input:    10050,
			expected: 100.50,
		},
		{
			name:     "negative money",
			input:    -5025,
			expected: -50.25,
		},
		{
			name:     "zero money",
			input:    0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToFloat()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadSystemTransactions(t *testing.T) {
	tests := []struct {
		name        string
		csvData     string
		startDate   time.Time
		endDate     time.Time
		expectedLen int
		expectError bool
	}{
		{
			name: "valid transactions within range",
			csvData: `trx_id,amount,type,timestamp
TRX001,100.50,CREDIT,2025-01-15 10:30:00
TRX002,50.25,DEBIT,2025-01-16 14:20:00
TRX003,200.00,CREDIT,2025-01-17 09:15:00`,
			startDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:     time.Date(2025, 1, 17, 23, 59, 59, 0, time.UTC),
			expectedLen: 3,
			expectError: false,
		},
		{
			name: "transactions outside range",
			csvData: `trx_id,amount,type,timestamp
TRX001,100.50,CREDIT,2025-01-10 10:30:00
TRX002,50.25,DEBIT,2025-01-20 14:20:00`,
			startDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:     time.Date(2025, 1, 17, 23, 59, 59, 0, time.UTC),
			expectedLen: 0,
			expectError: false,
		},
		{
			name: "mixed valid and invalid transactions",
			csvData: `trx_id,amount,type,timestamp
TRX001,100.50,CREDIT,2025-01-14 10:30:00
TRX002,50.25,DEBIT,2025-01-16 14:20:00
TRX003,200.00,CREDIT,2025-01-18 09:15:00`,
			startDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:     time.Date(2025, 1, 17, 23, 59, 59, 0, time.UTC),
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "empty CSV with header only",
			csvData:     `trx_id,amount,type,timestamp`,
			startDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:     time.Date(2025, 1, 17, 23, 59, 59, 0, time.UTC),
			expectedLen: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvData)
			transactions, err := LoadSystemTransactions(reader, tt.startDate, tt.endDate)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(transactions))
			}
		})
	}
}

func TestLoadBankStatement(t *testing.T) {
	tests := []struct {
		name        string
		csvData     string
		bankName    string
		startDate   time.Time
		endDate     time.Time
		expectedLen int
	}{
		{
			name: "valid bank transactions within range",
			csvData: `unique_id,amount,date
BANK001,100.50,2025-01-15
BANK002,-50.25,2025-01-16
BANK003,200.00,2025-01-17`,
			bankName:    "Test Bank",
			startDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:     time.Date(2025, 1, 17, 23, 59, 59, 0, time.UTC),
			expectedLen: 3,
		},
		{
			name: "transactions outside date range",
			csvData: `unique_id,amount,date
BANK001,100.50,2025-01-10
BANK002,-50.25,2025-01-20`,
			bankName:    "Test Bank",
			startDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:     time.Date(2025, 1, 17, 23, 59, 59, 0, time.UTC),
			expectedLen: 0,
		},
		{
			name:        "empty statement",
			csvData:     `unique_id,amount,date`,
			bankName:    "Test Bank",
			startDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			endDate:     time.Date(2025, 1, 17, 23, 59, 59, 0, time.UTC),
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.csvData)
			transactions, err := LoadBankStatement(reader, tt.bankName, tt.startDate, tt.endDate)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLen, len(transactions))

			for _, trx := range transactions {
				assert.Equal(t, tt.bankName, trx.BankName)
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		amount   Money
		expected string
	}{
		{
			name:     "positive amount",
			date:     time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			amount:   10050,
			expected: "2025-01-15 10:30-10050",
		},
		{
			name:     "negative amount",
			date:     time.Date(2025, 1, 15, 14, 45, 0, 0, time.UTC),
			amount:   -5025,
			expected: "2025-01-15 14:45--5025",
		},
		{
			name:     "zero amount",
			date:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			amount:   0,
			expected: "2025-01-15 00:00-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateKey(tt.date, tt.amount)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSignedAmount(t *testing.T) {
	tests := []struct {
		name        string
		transaction SystemTransaction
		expected    Money
	}{
		{
			name: "debit transaction - positive amount",
			transaction: SystemTransaction{
				Amount: 10050,
				Type:   Debit,
			},
			expected: -10050,
		},
		{
			name: "credit transaction - positive amount",
			transaction: SystemTransaction{
				Amount: 10050,
				Type:   Credit,
			},
			expected: 10050,
		},
		{
			name: "debit transaction - negative amount",
			transaction: SystemTransaction{
				Amount: -5025,
				Type:   Debit,
			},
			expected: -5025,
		},
		{
			name: "credit transaction - negative amount",
			transaction: SystemTransaction{
				Amount: -5025,
				Type:   Credit,
			},
			expected: 5025,
		},
		{
			name: "debit transaction - lowercase",
			transaction: SystemTransaction{
				Amount: 10000,
				Type:   "debit",
			},
			expected: -10000,
		},
		{
			name: "credit transaction - with whitespace",
			transaction: SystemTransaction{
				Amount: 10000,
				Type:   " CREDIT ",
			},
			expected: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSignedAmount(tt.transaction)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReconcileProcess(t *testing.T) {
	tests := []struct {
		name                   string
		systemTransactions     []SystemTransaction
		bankTransactions       []BankTransaction
		expectedMatched        int
		expectedUnmatched      int
		expectedUnmatchedSys   int
		expectedUnmatchedBank  int
		expectedTotalProcessed int
	}{
		{
			name: "perfect match - all transactions matched",
			systemTransactions: []SystemTransaction{
				{
					TransactionID:   "SYS001",
					Amount:          10050,
					Type:            Credit,
					TransactionTime: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
				},
				{
					TransactionID:   "SYS002",
					Amount:          5025,
					Type:            Debit,
					TransactionTime: time.Date(2025, 1, 16, 14, 20, 0, 0, time.UTC),
				},
			},
			bankTransactions: []BankTransaction{
				{
					BankName: "Bank A",
					UniqueID: "BANK001",
					Amount:   10050,
					Date:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				},
				{
					BankName: "Bank A",
					UniqueID: "BANK002",
					Amount:   -5025,
					Date:     time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedMatched:        2,
			expectedUnmatched:      0,
			expectedUnmatchedSys:   0,
			expectedUnmatchedBank:  0,
			expectedTotalProcessed: 4,
		},
		{
			name: "partial match - some unmatched",
			systemTransactions: []SystemTransaction{
				{
					TransactionID:   "SYS001",
					Amount:          10050,
					Type:            Credit,
					TransactionTime: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
				},
				{
					TransactionID:   "SYS002",
					Amount:          5025,
					Type:            Debit,
					TransactionTime: time.Date(2025, 1, 16, 14, 20, 0, 0, time.UTC),
				},
			},
			bankTransactions: []BankTransaction{
				{
					BankName: "Bank A",
					UniqueID: "BANK001",
					Amount:   10050,
					Date:     time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedMatched:        1,
			expectedUnmatched:      1,
			expectedUnmatchedSys:   1,
			expectedUnmatchedBank:  0,
			expectedTotalProcessed: 3,
		},
		{
			name: "no matches",
			systemTransactions: []SystemTransaction{
				{
					TransactionID:   "SYS001",
					Amount:          10050,
					Type:            Credit,
					TransactionTime: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			bankTransactions: []BankTransaction{
				{
					BankName: "Bank A",
					UniqueID: "BANK001",
					Amount:   20000,
					Date:     time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedMatched:        0,
			expectedUnmatched:      2,
			expectedUnmatchedSys:   1,
			expectedUnmatchedBank:  1,
			expectedTotalProcessed: 2,
		},
		{
			name:                   "empty transactions",
			systemTransactions:     []SystemTransaction{},
			bankTransactions:       []BankTransaction{},
			expectedMatched:        0,
			expectedUnmatched:      0,
			expectedUnmatchedSys:   0,
			expectedUnmatchedBank:  0,
			expectedTotalProcessed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconcileProcess(tt.systemTransactions, tt.bankTransactions)

			assert.Equal(t, tt.expectedMatched, result.TotalMatched, "TotalMatched mismatch")
			assert.Equal(t, tt.expectedUnmatched, result.TotalUnmatched, "TotalUnmatched mismatch")
			assert.Equal(t, tt.expectedUnmatchedSys, len(result.UnmatchedSystem), "UnmatchedSystem count mismatch")
			assert.Equal(t, tt.expectedTotalProcessed, result.TotalProcessed, "TotalProcessed mismatch")

			totalUnmatchedBank := 0
			for _, v := range result.UnmatchedBank {
				totalUnmatchedBank += len(v)
			}
			assert.Equal(t, tt.expectedUnmatchedBank, totalUnmatchedBank, "UnmatchedBank count mismatch")
		})
	}
}

func TestReconcile(t *testing.T) {
	tests := []struct {
		name          string
		startDate     string
		endDate       string
		sysCSV        string
		bankCSV       string
		expectError   bool
		errorContains string
	}{
		{
			name:      "valid date range and data",
			startDate: "2025-01-15",
			endDate:   "2025-01-17",
			sysCSV: `trx_id,amount,type,timestamp
SYS001,100.50,CREDIT,2025-01-15 10:30:00`,
			bankCSV: `unique_id,amount,date
BANK001,100.50,2025-01-15`,
			expectError: false,
		},
		{
			name:          "invalid start date format",
			startDate:     "15-01-2025",
			endDate:       "2025-01-17",
			sysCSV:        "trx_id,amount,type,timestamp",
			bankCSV:       "unique_id,amount,date",
			expectError:   true,
			errorContains: "invalid start_date",
		},
		{
			name:          "invalid end date format",
			startDate:     "2025-01-15",
			endDate:       "17/01/2025",
			sysCSV:        "trx_id,amount,type,timestamp",
			bankCSV:       "unique_id,amount,date",
			expectError:   true,
			errorContains: "invalid end_date",
		},
		{
			name:      "empty system transactions",
			startDate: "2025-01-15",
			endDate:   "2025-01-17",
			sysCSV:    "trx_id,amount,type,timestamp",
			bankCSV: `unique_id,amount,date
BANK001,100.50,2025-01-15`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewReconciliationService()

			sysReader := strings.NewReader(tt.sysCSV)

			// Create multipart form with bank CSV
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("bank_csv", "bank.csv")
			require.NoError(t, err)
			_, err = part.Write([]byte(tt.bankCSV))
			require.NoError(t, err)
			writer.Close()

			// Parse the multipart form
			reader := multipart.NewReader(body, writer.Boundary())
			form, err := reader.ReadForm(10 << 20)
			require.NoError(t, err)

			result, err := service.Reconcile(
				tt.startDate,
				tt.endDate,
				sysReader,
				form,
				nil,
				nil,
			)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestNewReconciliationService(t *testing.T) {
	service := NewReconciliationService()
	assert.NotNil(t, service)
	assert.Implements(t, (*ReconciliationService)(nil), service)
}
