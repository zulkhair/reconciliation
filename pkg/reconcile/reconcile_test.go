package reconcile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reconciliation/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateTransactions(count int) []types.Transaction {
	transactions := make([]types.Transaction, count)

	for i := 0; i < count; i++ {
		transactions[i] = types.Transaction{
			TrxID:           fmt.Sprintf("T%06d", i+1),
			Amount:          100.00,
			Type:            "CREDIT",
			TransactionTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
	}

	return transactions
}

func generateBankStatements(count int) []types.BankStatement {
	bankStatements := make([]types.BankStatement, count)

	for i := 0; i < count; i++ {
		bankStatements[i] = types.BankStatement{
			UniqueID: fmt.Sprintf("B%06d", i+1),
			Amount:   100.00,
			Date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}
	}

	return bankStatements
}

func TestReconcile(t *testing.T) {
	// Helper functions to create time.Time from string
	parseDateTime := func(date string) time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05", date)
		return t
	}
	parseDate := func(date string) time.Time {
		t, _ := time.Parse("2006-01-02", date)
		return t
	}

	// generate 100 transactions
	systemTxs := generateTransactions(100)
	bankTxs := generateBankStatements(100)

	tests := []struct {
		name           string
		systemTxs      []types.Transaction
		bankTxs        []types.BankStatement
		expectedResult ReconcileResult
	}{
		{
			name:      "Perfect match - single transaction",
			systemTxs: systemTxs,
			bankTxs:   bankTxs,
			expectedResult: ReconcileResult{
				TransactionProcessed: 100,
				TransactionMatched:   100,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 0,
					SystemUnmatched:      nil,
					BankUnmatched:        nil,
				},
			},
		},
		{
			name: "No matches - different dates",
			systemTxs: []types.Transaction{
				{
					TrxID:           "TRX1",
					Amount:          100.00,
					Type:            "CREDIT",
					TransactionTime: parseDateTime("2024-03-20 10:30:00"),
				},
			},
			bankTxs: []types.BankStatement{
				{
					UniqueID: "BANK1",
					Amount:   100.00,
					Date:     parseDate("2024-03-21"),
				},
			},
			expectedResult: ReconcileResult{
				TransactionProcessed: 1,
				TransactionMatched:   0,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 2,
					SystemUnmatched: []types.Transaction{
						{
							TrxID:           "TRX1",
							Amount:          100.00,
							Type:            "CREDIT",
							TransactionTime: parseDateTime("2024-03-20 10:30:00"),
						},
					},
					BankUnmatched: []types.BankStatement{
						{
							UniqueID: "BANK1",
							Amount:   100.00,
							Date:     parseDate("2024-03-21"),
						},
					},
				},
			},
		},
		{
			name: "Match with small discrepancy",
			systemTxs: []types.Transaction{
				{
					TrxID:           "TRX1",
					Amount:          100.00,
					Type:            "CREDIT",
					TransactionTime: parseDateTime("2024-03-20 10:30:00"),
				},
			},
			bankTxs: []types.BankStatement{
				{
					UniqueID: "BANK1",
					Amount:   100 + amountTolerance,
					Date:     parseDate("2024-03-20"),
				},
			},
			expectedResult: ReconcileResult{
				TransactionProcessed: 1,
				TransactionMatched:   1,
				TotalDiscrepancies:   0.005,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 0,
					SystemUnmatched:      nil,
					BankUnmatched:        nil,
				},
			},
		},
		{
			name:      "Empty input lists",
			systemTxs: []types.Transaction{},
			bankTxs:   []types.BankStatement{},
			expectedResult: ReconcileResult{
				TransactionProcessed: 0,
				TransactionMatched:   0,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 0,
					SystemUnmatched:      nil,
					BankUnmatched:        nil,
				},
			},
		},
		{
			name: "Multiple transactions with mixed matches",
			systemTxs: []types.Transaction{
				{
					TrxID:           "TRX1",
					Amount:          100.00,
					Type:            "CREDIT",
					TransactionTime: parseDateTime("2024-03-20 10:30:00"),
				},
				{
					TrxID:           "TRX2",
					Amount:          200.00,
					Type:            "CREDIT",
					TransactionTime: parseDateTime("2024-03-20 11:30:00"),
				},
				{
					TrxID:           "TRX3",
					Amount:          300.00,
					Type:            "CREDIT",
					TransactionTime: parseDateTime("2024-03-20 12:30:00"),
				},
			},
			bankTxs: []types.BankStatement{
				{
					UniqueID: "BANK1",
					Amount:   100.00,
					Date:     parseDate("2024-03-20"),
					BankName: "BankA",
				},
				{
					UniqueID: "BANK2",
					Amount:   200.0021, // Slightly different amount
					Date:     parseDate("2024-03-20"),
					BankName: "BankA",
				},
			},
			expectedResult: ReconcileResult{
				TransactionProcessed: 3,
				TransactionMatched:   2,
				TotalDiscrepancies:   0.0021,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 1,
					SystemUnmatched: []types.Transaction{
						{
							TrxID:           "TRX3",
							Amount:          300.00,
							Type:            "CREDIT",
							TransactionTime: parseDateTime("2024-03-20 12:30:00"),
						},
					},
					BankUnmatched: nil,
				},
			},
		},
		{
			name: "Only system transactions",
			systemTxs: []types.Transaction{
				{
					TrxID:           "TRX1",
					Amount:          100.00,
					Type:            "CREDIT",
					TransactionTime: parseDateTime("2024-03-20 10:30:00"),
				},
			},
			bankTxs: []types.BankStatement{},
			expectedResult: ReconcileResult{
				TransactionProcessed: 1,
				TransactionMatched:   0,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 1,
					SystemUnmatched: []types.Transaction{
						{
							TrxID:           "TRX1",
							Amount:          100.00,
							Type:            "CREDIT",
							TransactionTime: parseDateTime("2024-03-20 10:30:00"),
						},
					},
					BankUnmatched: nil,
				},
			},
		},
		{
			name:      "Only bank transactions",
			systemTxs: []types.Transaction{},
			bankTxs: []types.BankStatement{
				{
					UniqueID: "BANK1",
					Amount:   100.00,
					Date:     parseDate("2024-03-20"),
					BankName: "BankA",
				},
			},
			expectedResult: ReconcileResult{
				TransactionProcessed: 0,
				TransactionMatched:   0,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 1,
					SystemUnmatched:      nil,
					BankUnmatched: []types.BankStatement{
						{
							UniqueID: "BANK1",
							Amount:   100.00,
							Date:     parseDate("2024-03-20"),
							BankName: "BankA",
						},
					},
				},
			},
		},
		{
			name: "DEBIT transaction match",
			systemTxs: []types.Transaction{
				{
					TrxID:           "TRX1",
					Amount:          100.00,
					Type:            "DEBIT",
					TransactionTime: parseDateTime("2024-03-20 10:30:00"),
				},
			},
			bankTxs: []types.BankStatement{
				{
					UniqueID: "BANK1",
					Amount:   -100.00, // Negative amount for DEBIT
					Date:     parseDate("2024-03-20"),
					BankName: "BankA",
				},
			},
			expectedResult: ReconcileResult{
				TransactionProcessed: 1,
				TransactionMatched:   1,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 0,
					SystemUnmatched:      nil,
					BankUnmatched:        nil,
				},
			},
		},
		{
			name: "CREDIT transaction match",
			systemTxs: []types.Transaction{
				{
					TrxID:           "TRX1",
					Amount:          100.00,
					Type:            "CREDIT",
					TransactionTime: parseDateTime("2024-03-20 10:30:00"),
				},
			},
			bankTxs: []types.BankStatement{
				{
					UniqueID: "BANK1",
					Amount:   100.00, // Positive amount for CREDIT
					Date:     parseDate("2024-03-20"),
					BankName: "BankA",
				},
			},
			expectedResult: ReconcileResult{
				TransactionProcessed: 1,
				TransactionMatched:   1,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 0,
					SystemUnmatched:      nil,
					BankUnmatched:        nil,
				},
			},
		},
		{
			name: "Mixed DEBIT/CREDIT transactions",
			systemTxs: []types.Transaction{
				{
					TrxID:           "TRX1",
					Amount:          100.00,
					Type:            "DEBIT",
					TransactionTime: parseDateTime("2024-03-20 10:30:00"),
				},
				{
					TrxID:           "TRX2",
					Amount:          200.00,
					Type:            "CREDIT",
					TransactionTime: parseDateTime("2024-03-20 11:30:00"),
				},
			},
			bankTxs: []types.BankStatement{
				{
					UniqueID: "BANK1",
					Amount:   -100.00,
					Date:     parseDate("2024-03-20"),
					BankName: "BankA",
				},
				{
					UniqueID: "BANK2",
					Amount:   200.00,
					Date:     parseDate("2024-03-20"),
					BankName: "BankA",
				},
			},
			expectedResult: ReconcileResult{
				TransactionProcessed: 2,
				TransactionMatched:   2,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 0,
					SystemUnmatched:      nil,
					BankUnmatched:        nil,
				},
			},
		},
		{
			name: "Mismatched transaction types",
			systemTxs: []types.Transaction{
				{
					TrxID:           "TRX1",
					Amount:          100.00,
					Type:            "DEBIT",
					TransactionTime: parseDateTime("2024-03-20 10:30:00"),
				},
			},
			bankTxs: []types.BankStatement{
				{
					UniqueID: "BANK1",
					Amount:   100.00, // Should be negative for DEBIT
					Date:     parseDate("2024-03-20"),
					BankName: "BankA",
				},
			},
			expectedResult: ReconcileResult{
				TransactionProcessed: 1,
				TransactionMatched:   0,
				TotalDiscrepancies:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 2,
					SystemUnmatched: []types.Transaction{
						{
							TrxID:           "TRX1",
							Amount:          100.00,
							Type:            "DEBIT",
							TransactionTime: parseDateTime("2024-03-20 10:30:00"),
						},
					},
					BankUnmatched: []types.BankStatement{
						{
							UniqueID: "BANK1",
							Amount:   100.00,
							Date:     parseDate("2024-03-20"),
							BankName: "BankA",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reconcile(tt.systemTxs, tt.bankTxs)
			assert.Equal(t, tt.expectedResult.TransactionProcessed, result.TransactionProcessed)
			assert.Equal(t, tt.expectedResult.TransactionMatched, result.TransactionMatched)
			assert.InDelta(t, tt.expectedResult.TotalDiscrepancies, result.TotalDiscrepancies, amountTolerance)
			assert.Equal(t, tt.expectedResult.TransactionUnmatched.TransactionUnmatched,
				result.TransactionUnmatched.TransactionUnmatched)
			assert.Equal(t, tt.expectedResult.TransactionUnmatched.SystemUnmatched,
				result.TransactionUnmatched.SystemUnmatched)
			assert.Equal(t, tt.expectedResult.TransactionUnmatched.BankUnmatched,
				result.TransactionUnmatched.BankUnmatched)
		})
	}
}

func TestIsMatch(t *testing.T) {
	parseDateTime := func(date string) time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05", date)
		return t
	}

	parseDate := func(date string) time.Time {
		t, _ := time.Parse("2006-01-02", date)
		return t
	}

	tests := []struct {
		name     string
		sysTx    types.Transaction
		bankTx   types.BankStatement
		expected bool
	}{
		{
			name: "Exact match",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "CREDIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: 100.00,
				Date:   parseDate("2024-03-20"),
			},
			expected: true,
		},
		{
			name: "Different date",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "CREDIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: 100.00,
				Date:   parseDate("2024-03-21"),
			},
			expected: false,
		},
		{
			name: "Amount outside tolerance",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "CREDIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: 101.00,
				Date:   parseDate("2024-03-20"),
			},
			expected: false,
		},
		{
			name: "Zero amount transactions",
			sysTx: types.Transaction{
				Amount:          0.00,
				Type:            "CREDIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: 0.00,
				Date:   parseDate("2024-03-20"),
			},
			expected: true,
		},
		{
			name: "Negative amount transactions",
			sysTx: types.Transaction{
				Amount:          -100.00,
				Type:            "DEBIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: -100.00,
				Date:   parseDate("2024-03-20"),
			},
			expected: false,
		},
		{
			name: "Amount within tolerance (upper bound)",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "CREDIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: 100.00 + amountTolerance,
				Date:   parseDate("2024-03-20"),
			},
			expected: true,
		},
		{
			name: "Amount within tolerance (lower bound)",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "CREDIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: 100.00 - amountTolerance,
				Date:   parseDate("2024-03-20"),
			},
			expected: true,
		},
		{
			name: "DEBIT transaction match",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "DEBIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: -100.00,
				Date:   parseDate("2024-03-20"),
			},
			expected: true,
		},
		{
			name: "CREDIT transaction match",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "CREDIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: 100.00,
				Date:   parseDate("2024-03-20"),
			},
			expected: true,
		},
		{
			name: "DEBIT transaction with wrong sign",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "DEBIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: 100.00, // Should be negative for DEBIT
				Date:   parseDate("2024-03-20"),
			},
			expected: false,
		},
		{
			name: "CREDIT transaction with wrong sign",
			sysTx: types.Transaction{
				Amount:          100.00,
				Type:            "CREDIT",
				TransactionTime: parseDateTime("2024-03-20 10:30:00"),
			},
			bankTx: types.BankStatement{
				Amount: -100.00, // Should be positive for CREDIT
				Date:   parseDate("2024-03-20"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMatch(tt.sysTx, tt.bankTx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReconcileResult_String(t *testing.T) {
	parseDateTime := func(date string) time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05", date)
		return t
	}

	parseDate := func(date string) time.Time {
		t, _ := time.Parse("2006-01-02", date)
		return t
	}

	tests := []struct {
		name            string
		reconcileResult ReconcileResult
		expectedOutput  string
	}{
		{
			name: "Empty result",
			reconcileResult: ReconcileResult{
				TransactionProcessed: 0,
				TransactionMatched:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 0,
				},
				TotalDiscrepancies: 0,
			},
			expectedOutput: "Reconciliation Summary:\n" +
				"------------------------\n" +
				"Total transactions processed: 0\n" +
				"Total matched transactions: 0\n" +
				"Total unmatched transactions: 0\n" +
				"\nTotal amount discrepancies: 0.00\n",
		},
		{
			name: "Result with unmatched transactions",
			reconcileResult: ReconcileResult{
				TransactionProcessed: 3,
				TransactionMatched:   1,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 2,
					SystemUnmatched: []types.Transaction{
						{
							TrxID:           "TRX1",
							Amount:          100.00,
							Type:            "CREDIT",
							TransactionTime: parseDateTime("2024-03-20 10:30:00"),
						},
					},
					BankUnmatched: []types.BankStatement{
						{
							UniqueID: "BANK1",
							Amount:   200.00,
							Date:     parseDate("2024-03-20"),
							BankName: "BankA",
						},
					},
				},
				TotalDiscrepancies: 0.50,
			},
			expectedOutput: "Reconciliation Summary:\n" +
				"------------------------\n" +
				"Total transactions processed: 3\n" +
				"Total matched transactions: 1\n" +
				"Total unmatched transactions: 2\n" +
				"\nSystem transactions missing from bank statements:\n" +
				"- TrxID: TRX1, Amount: 100.00, Date: 2024-03-20 10:30:00\n" +
				"\nBank statements missing from system transactions:\n" +
				"\nBank: BankA\n" +
				"- ID: BANK1, Amount: 200.00, Date: 2024-03-20\n" +
				"\nTotal amount discrepancies: 0.50\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.reconcileResult.String()
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

func TestReconcileResult_GenerateJSON(t *testing.T) {
	parseDateTime := func(date string) time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05", date)
		return t
	}

	parseDate := func(date string) time.Time {
		t, _ := time.Parse("2006-01-02", date)
		return t
	}

	tests := []struct {
		name            string
		reconcileResult ReconcileResult
		expectedError   bool
		validateJSON    func(t *testing.T, filename string)
	}{
		{
			name: "Generate JSON with complete data",
			reconcileResult: ReconcileResult{
				TransactionProcessed: 3,
				TransactionMatched:   1,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 2,
					SystemUnmatched: []types.Transaction{
						{
							TrxID:           "TRX1",
							Amount:          100.00,
							Type:            "CREDIT",
							TransactionTime: parseDateTime("2024-03-20 10:30:00"),
						},
					},
					BankUnmatched: []types.BankStatement{
						{
							UniqueID: "BANK1",
							Amount:   200.00,
							Date:     parseDate("2024-03-20"),
							BankName: "BankA",
						},
					},
				},
				TotalDiscrepancies: 0.50,
			},
			expectedError: false,
			validateJSON: func(t *testing.T, filename string) {
				data, err := os.ReadFile(filename)
				assert.NoError(t, err)

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				assert.NoError(t, err)

				summary, ok := result["summary"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, float64(3), summary["total_transactions_processed"])
				assert.Equal(t, float64(1), summary["total_transactions_matched"])
				assert.Equal(t, float64(2), summary["total_transactions_unmatched"])
				assert.Equal(t, 0.50, summary["total_discrepancies"])

				unmatchedDetails, ok := result["unmatched_details"].(map[string]interface{})
				assert.True(t, ok)

				systemTxs, ok := unmatchedDetails["system_transactions"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, systemTxs, 1)

				bankStmts, ok := unmatchedDetails["bank_statements"].(map[string]interface{})
				assert.True(t, ok)
				assert.Contains(t, bankStmts, "BankA")
			},
		},
		{
			name: "Generate JSON with empty data",
			reconcileResult: ReconcileResult{
				TransactionProcessed: 0,
				TransactionMatched:   0,
				TransactionUnmatched: ReconcileUnmatched{
					TransactionUnmatched: 0,
				},
				TotalDiscrepancies: 0,
			},
			expectedError: false,
			validateJSON: func(t *testing.T, filename string) {
				data, err := os.ReadFile(filename)
				assert.NoError(t, err)

				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				assert.NoError(t, err)

				summary, ok := result["summary"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, float64(0), summary["total_transactions_processed"])
				assert.Equal(t, float64(0), summary["total_transactions_matched"])
				assert.Equal(t, float64(0), summary["total_transactions_unmatched"])
				assert.Equal(t, float64(0), summary["total_discrepancies"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := filepath.Join(t.TempDir(), "result.json")
			err := tt.reconcileResult.GenerateJSON(tempFile)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validateJSON(t, tempFile)
			}

			// Cleanup
			os.Remove(tempFile)
		})
	}
}
