package csv

import (
	"encoding/csv"
	"reconciliation/pkg/types"
	"time"
)

type CSVReader interface {
	ReadSystemTransactionsFromCSV() ([]types.Transaction, error)
	ReadBankStatementsFromCSV() ([]types.BankStatement, error)
}

type CSVReaderImpl struct {
	reader *csv.Reader

	// Filename of the CSV file
	filename string

	// Time range for filtering
	start time.Time
	end   time.Time

	// Skip Header
	skipHeader bool
}

// Option is a functional option for the CSVReader
type Option func(*CSVReaderImpl)

// WithTimeRange sets the time range for filtering
func WithTimeRange(start, end time.Time) Option {
	return func(r *CSVReaderImpl) {
		r.start = start
		r.end = end
	}
}

// WithSkipHeader skips the header row
func WithSkipHeader(skipHeader bool) Option {
	return func(r *CSVReaderImpl) {
		r.skipHeader = skipHeader
	}
}

// WithFilename sets the filename for the CSV reader
func WithFilename(filename string) Option {
	return func(r *CSVReaderImpl) {
		r.filename = filename
	}
}
