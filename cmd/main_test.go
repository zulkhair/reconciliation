package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessBankFiles(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "test-bank-files")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create Sample CSV files
	testFiles := []string{"bri.csv", "bni.csv", "mandiri.csv"}
	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tmpDir, file))
		assert.NoError(t, err)
		f.Close()
	}

	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "Directory with multiple CSV files",
			input:   tmpDir,
			want:    3,
			wantErr: false,
		},
		{
			name:    "Comma-separated file paths",
			input:   filepath.Join(tmpDir, "bri.csv") + "," + filepath.Join(tmpDir, "bni.csv") + "," + filepath.Join(tmpDir, "mandiri.csv"),
			want:    3,
			wantErr: false,
		},
		{
			name:    "Non-existent directory",
			input:   "/non/existent/dir",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processBankFiles(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, len(got))
		})
	}
}

func TestReadSystemTransactions(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test-system-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write test data
	testData := `TrxID,Amount,Type,TransactionTime
TX001,100.0,DEBIT,2024-01-01 10:00:00
TX002,200.0,CREDIT,2024-01-02 10:00:00`
	_, err = tmpFile.WriteString(testData)
	assert.NoError(t, err)
	tmpFile.Close()

	tests := []struct {
		name      string
		startDate string
		endDate   string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Valid date range",
			startDate: "2024-01-01",
			endDate:   "2024-01-03",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "Partial date range",
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "Invalid date range",
			startDate: "2024-01-03",
			endDate:   "2024-01-01",
			wantCount: 0,
			wantErr:   false, // expected false because this already handled in the other function
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, err := time.Parse("2006-01-02", tt.startDate)
			assert.NoError(t, err)

			end, err := time.Parse("2006-01-02", tt.endDate)
			assert.NoError(t, err)

			transactions, err := readSystemTransactions(tmpFile.Name(), start, end)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(transactions))
		})
	}
}

func TestReadBankStatements(t *testing.T) {
	// Create temporary test files
	tmpDir, err := os.MkdirTemp("", "test-bank-statements")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create and write to test files
	testFiles := []string{"bank1.csv", "bank2.csv"}
	testData := `UniqueID,Amount,Date
BS001,-100.0,2024-01-01
BS002,200.0,2024-01-02`

	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tmpDir, file))
		assert.NoError(t, err)
		_, err = f.WriteString(testData)
		assert.NoError(t, err)
		f.Close()
	}

	tests := []struct {
		name      string
		files     []string
		startDate string
		endDate   string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Multiple valid files",
			files:     []string{filepath.Join(tmpDir, "bank1.csv"), filepath.Join(tmpDir, "bank2.csv")},
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 4, // 2 transactions per file
			wantErr:   false,
		},
		{
			name:      "Non-existent file",
			files:     []string{filepath.Join(tmpDir, "nonexistent.csv")},
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, err := time.Parse("2006-01-02", tt.startDate)
			assert.NoError(t, err)

			end, err := time.Parse("2006-01-02", tt.endDate)
			assert.NoError(t, err)

			statements, err := readBankStatements(tt.files, start, end)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(statements))
		})
	}
}
