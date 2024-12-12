package csv

import (
	"bytes"
	"encoding/csv"
	"reconciliation/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CSVReaderTestSuite is a test suite for the CSVReader
type CSVReaderTestSuite struct {
	suite.Suite
}

// TestCSVReaderSuite runs the test suite
func TestCSVReaderSuite(t *testing.T) {
	suite.Run(t, new(CSVReaderTestSuite))
}

// TestReadSystemTransactionsFromCSV tests the ReadSystemTransactionsFromCSV function
func (s *CSVReaderTestSuite) TestReadSystemTransactionsFromCSV() {
	// Define test cases
	testCases := []struct {
		name          string
		csvContent    string
		timeRange     *struct{ start, end time.Time }
		skipHeader    bool
		expected      []types.Transaction
		expectedError string
	}{
		{
			name: "valid system transactions",
			csvContent: `TrxID,Amount,Type,TransactionTime
TX001,100.0,DEBIT,2024-01-01 10:00:00
TX002,200.0,CREDIT,2024-01-02 10:00:00`,
			skipHeader: true,
			expected: []types.Transaction{
				{
					TrxID:           "TX001",
					Amount:          100.0,
					Type:            types.TransactionTypeDebit,
					TransactionTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				},
				{
					TrxID:           "TX002",
					Amount:          200.0,
					Type:            types.TransactionTypeCredit,
					TransactionTime: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "invalid system transactions with negative amounts",
			csvContent: `TrxID,Amount,Type,TransactionTime
TX001,-100.0,DEBIT,2024-01-01 10:00:00
TX002,-200.0,CREDIT,2024-01-02 10:00:00`,
			skipHeader:    true,
			expectedError: "negative amount [-100.0] in row 2 of file",
		},
		{
			name: "invalid amount format",
			csvContent: `TrxID,Amount,Type,TransactionTime
TX001,invalid,DEBIT,2024-01-01 10:00:00`,
			skipHeader:    true,
			expectedError: "invalid amount [invalid] in row 2 of file",
		},
		{
			name: "invalid date format",
			csvContent: `TrxID,Amount,Type,TransactionTime
TX001,100.0,DEBIT,invalid-date`,
			skipHeader:    true,
			expectedError: "invalid date [invalid-date] in row 2 of file",
		},
		{
			name: "with time range filter",
			csvContent: `TrxID,Amount,Type,TransactionTime
TX001,100.0,DEBIT,2024-01-01 10:00:00
TX002,200.0,CREDIT,2024-01-02 10:00:00
TX003,300.0,DEBIT,2024-01-03 10:00:00`,
			skipHeader: true,
			timeRange: &struct{ start, end time.Time }{
				start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 1, 2, 23, 59, 59, 0, time.UTC),
			},
			expected: []types.Transaction{
				{
					TrxID:           "TX001",
					Amount:          100.0,
					Type:            types.TransactionTypeDebit,
					TransactionTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
				},
				{
					TrxID:           "TX002",
					Amount:          200.0,
					Type:            types.TransactionTypeCredit,
					TransactionTime: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "empty CSV file",
			csvContent: `TrxID,Amount,Type,TransactionTime
`,
			skipHeader: true,
			expected:   []types.Transaction{},
		},
		{
			name: "missing required columns",
			csvContent: `TrxID,Amount,Type
TX001,100.0,DEBIT`,
			skipHeader:    true,
			expectedError: "invalid format [TX001,100.0,DEBIT] in row 2 of file",
		},
		{
			name: "too many columns",
			csvContent: `TrxID,Amount,Type,TransactionTime
TX001,100.0,DEBIT,2024-01-01 10:00:00,extra`,
			skipHeader:    true,
			expectedError: "failed to read CSV file: record on line 2: wrong number of fields",
		},
		{
			name:       "completely empty file",
			csvContent: "",
			expected:   []types.Transaction{},
		},
		{
			name: "invalid time range (end before start)",
			csvContent: `TrxID,Amount,Type,TransactionTime
TX001,100.0,DEBIT,2024-01-01 10:00:00`,
			skipHeader: true,
			timeRange: &struct{ start, end time.Time }{
				start: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: []types.Transaction{},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create a CSV reader with the test case's CSV content
			reader := csv.NewReader(bytes.NewBufferString(tc.csvContent))

			// Apply options
			var opts []Option
			if tc.timeRange != nil {
				opts = append(opts, WithTimeRange(tc.timeRange.start, tc.timeRange.end))
			}
			if tc.skipHeader {
				opts = append(opts, WithSkipHeader(true))
			}
			csvReader := NewCSVReader(reader, opts...)

			// Read the system transactions
			transactions, err := csvReader.ReadSystemTransactionsFromCSV()

			// Check if there was an error
			if tc.expectedError != "" {
				assert.EqualError(s.T(), err, tc.expectedError)
			} else {
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), tc.expected, transactions)
			}
		})
	}
}

// TestReadBankStatementsFromCSV tests the ReadBankStatementsFromCSV function
func (s *CSVReaderTestSuite) TestReadBankStatementsFromCSV() {
	// Define test cases
	testCases := []struct {
		name          string
		csvContent    string
		filename      string
		timeRange     *struct{ start, end time.Time }
		skipHeader    bool
		expected      []types.BankStatement
		expectedError string
	}{
		{
			name: "valid bank statements",
			csvContent: `UniqueID,Amount,Date
BS001,-100.0,2024-01-01
BS002,200.0,2024-01-02`,
			filename:   "bri.csv",
			skipHeader: true,
			expected: []types.BankStatement{
				{
					BankName: "BRI",
					UniqueID: "BS001",
					Amount:   -100.0,
					Date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					BankName: "BRI",
					UniqueID: "BS002",
					Amount:   200.0,
					Date:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "invalid amount format",
			csvContent: `UniqueID,Amount,Date
BS001,invalid,2024-01-01`,
			filename:      "bri.csv",
			skipHeader:    true,
			expectedError: "invalid amount [invalid] in row 2 of file",
		},
		{
			name: "invalid date format",
			csvContent: `UniqueID,Amount,Date
BS001,100.0,invalid-date`,
			filename:      "bri.csv",
			skipHeader:    true,
			expectedError: "invalid date [invalid-date] in row 2 of file",
		},
		{
			name: "with time range filter",
			csvContent: `UniqueID,Amount,Date
BS001,-100.0,2024-01-01
BS002,200.0,2024-01-02
BS003,-300.0,2024-01-03`,
			filename:   "bri.csv",
			skipHeader: true,
			timeRange: &struct{ start, end time.Time }{
				start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 1, 2, 23, 59, 59, 0, time.UTC),
			},
			expected: []types.BankStatement{
				{
					BankName: "BRI",
					UniqueID: "BS001",
					Amount:   -100.0,
					Date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					BankName: "BRI",
					UniqueID: "BS002",
					Amount:   200.0,
					Date:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "empty CSV file",
			csvContent: `UniqueID,Amount,Date
`,
			filename:   "bri.csv",
			skipHeader: true,
			expected:   []types.BankStatement{},
		},
		{
			name: "missing required columns",
			csvContent: `UniqueID,Amount
BS001,100.0`,
			filename:      "bri.csv",
			skipHeader:    true,
			expectedError: "invalid format [BS001,100.0] in row 2 of file",
		},
		{
			name: "too many columns",
			csvContent: `UniqueID,Amount,Date
BS001,100.0,2024-01-01,extra`,
			filename:      "bri.csv",
			skipHeader:    true,
			expectedError: "failed to read CSV file: record on line 2: wrong number of fields",
		},
		{
			name:       "completely empty file",
			csvContent: "",
			filename:   "bri.csv",
			expected:   []types.BankStatement{},
		},
		{
			name: "invalid time range (end before start)",
			csvContent: `UniqueID,Amount,Date
BS001,100.0,2024-01-01`,
			filename:   "bri.csv",
			skipHeader: true,
			timeRange: &struct{ start, end time.Time }{
				start: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: []types.BankStatement{}, // Should return empty when no statements in range
		},
	}

	// Run each test case
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create a CSV reader with the test case's CSV content
			reader := csv.NewReader(bytes.NewBufferString(tc.csvContent))

			// Apply options
			var opts []Option
			if tc.timeRange != nil {
				opts = append(opts, WithTimeRange(tc.timeRange.start, tc.timeRange.end))
			}
			if tc.skipHeader {
				opts = append(opts, WithSkipHeader(true))
			}
			if tc.filename != "" {
				opts = append(opts, WithFilename(tc.filename))
			}
			csvReader := NewCSVReader(reader, opts...)

			// Read the bank statements
			statements, err := csvReader.ReadBankStatementsFromCSV()

			// Check if there was an error
			if tc.expectedError != "" {
				assert.EqualError(s.T(), err, tc.expectedError)
			} else {
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), tc.expected, statements)
			}
		})
	}
}
