package csv

import (
	"encoding/csv"
	"fmt"
	"path/filepath"
	"reconciliation/pkg/types"
	"strconv"
	"strings"
	"time"
)

// NewCSVReader creates a new CSVReader
func NewCSVReader(reader *csv.Reader, opts ...Option) *CSVReaderImpl {
	r := &CSVReaderImpl{
		reader: reader,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// ReadSystemTransactionsFromCSV reads a CSV file and parses it into a slice of Transaction
func (r *CSVReaderImpl) ReadSystemTransactionsFromCSV() ([]types.Transaction, error) {
	records, err := r.reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		return []types.Transaction{}, nil
	}

	// Pre-allocate slice with estimated capacity
	transactions := make([]types.Transaction, 0, len(records)-1)

	// Check time range once
	hasTimeRange := !r.start.IsZero() && !r.end.IsZero()

	// Determine starting index based on skipHeader flag
	startIdx := 0
	if r.skipHeader {
		startIdx = 1
	}

	for i, record := range records[startIdx:] {
		if len(record) != 4 {
			return nil, fmt.Errorf("invalid format [%s] in row %d of file", strings.Join(record, ","), i+startIdx+1)
		}

		amount, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount [%s] in row %d of file", record[1], i+startIdx+1)
		}

		// Check negative amount
		if amount < 0 {
			return nil, fmt.Errorf("negative amount [%s] in row %d of file", record[1], i+startIdx+1)
		}

		// Parse date in YYYY-MM-DD HH:MM:SS format
		date, err := time.Parse("2006-01-02 15:04:05", record[3])
		if err != nil {
			return nil, fmt.Errorf("invalid date [%s] in row %d of file", record[3], i+startIdx+1)
		}

		// Skip if outside time range when range is set
		if hasTimeRange {
			dateForComparison := date.Truncate(24 * time.Hour)
			if dateForComparison.Before(r.start) || dateForComparison.After(r.end) {
				continue
			}
		}

		transactions = append(transactions, types.Transaction{
			TrxID:           record[0],
			Amount:          amount,
			Type:            types.TransactionType(record[2]),
			TransactionTime: date,
		})
	}
	return transactions, nil
}

// ReadBankStatementsFromCSV reads a CSV file and parses it into a slice of BankStatement
func (r *CSVReaderImpl) ReadBankStatementsFromCSV() ([]types.BankStatement, error) {
	records, err := r.reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) == 0 {
		return []types.BankStatement{}, nil
	}

	// Pre-allocate slice with estimated capacity
	statements := make([]types.BankStatement, 0, len(records)-1)

	// Check time range once
	hasTimeRange := !r.start.IsZero() && !r.end.IsZero()

	// Determine starting index based on skipHeader flag
	startIdx := 0
	if r.skipHeader {
		startIdx = 1
	}

	// Get bank name from filename
	bankName := filepath.Base(r.filename)
	bankName = strings.TrimSuffix(bankName, filepath.Ext(bankName))
	bankName = strings.ToUpper(bankName)

	for i, record := range records[startIdx:] {
		if len(record) != 3 {
			return nil, fmt.Errorf("invalid format [%s] in row %d of file", strings.Join(record, ","), i+startIdx+1)
		}

		amount, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount [%s] in row %d of file", record[1], i+startIdx+1)
		}

		// Parse date in YYYY-MM-DD format
		date, err := time.Parse("2006-01-02", record[2])
		if err != nil {
			return nil, fmt.Errorf("invalid date [%s] in row %d of file", record[2], i+startIdx+1)
		}

		// Skip if outside time range when range is set
		if hasTimeRange {
			if date.Before(r.start) || date.After(r.end) {
				continue
			}
		}

		statements = append(statements, types.BankStatement{
			BankName: bankName,
			UniqueID: record[0],
			Amount:   amount,
			Date:     date,
		})
	}
	return statements, nil
}
