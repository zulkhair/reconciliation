package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestProcessBankFiles tests the processBankFiles function
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

	// Define test cases
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
		{
			name:    "Empty input",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Invalid file path characters",
			input:   "/invalid/\x00/path",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Directory without CSV files",
			input:   os.TempDir(),
			want:    0,
			wantErr: false,
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the processBankFiles function
			got, err := processBankFiles(tt.input)

			// Check if the result matches the expected result
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, len(got))
		})
	}
}

// TestReadSystemTransactions tests the readSystemTransactions function
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

	// Create an empty file for testing
	emptyFile, err := os.CreateTemp("", "empty-*.csv")
	assert.NoError(t, err)
	defer os.Remove(emptyFile.Name())

	// Create invalid CSV file
	invalidFile, err := os.CreateTemp("", "invalid-*.csv")
	assert.NoError(t, err)
	defer os.Remove(invalidFile.Name())
	_, err = invalidFile.WriteString("invalid,csv,format\nwithout,proper,headers")
	assert.NoError(t, err)
	invalidFile.Close()

	// Define test cases
	tests := []struct {
		name      string
		file      string
		startDate string
		endDate   string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Valid date range",
			file:      tmpFile.Name(),
			startDate: "2024-01-01",
			endDate:   "2024-01-03",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "Partial date range",
			file:      tmpFile.Name(),
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "Invalid date range",
			file:      tmpFile.Name(),
			startDate: "2024-01-03",
			endDate:   "2024-01-01",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "Non-existent file",
			file:      "nonexistent.csv",
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "Empty file",
			file:      emptyFile.Name(),
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "Invalid CSV format",
			file:      invalidFile.Name(),
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 0,
			wantErr:   true,
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the start and end dates
			start, err := time.Parse("2006-01-02", tt.startDate)
			assert.NoError(t, err)

			// Parse the end date
			end, err := time.Parse("2006-01-02", tt.endDate)
			assert.NoError(t, err)

			// Call the readSystemTransactions function
			transactions, err := readSystemTransactions(tt.file, start, end)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			// Check if the result matches the expected result
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(transactions))
		})
	}
}

// TestReadBankStatements tests the readBankStatements function
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

	// Create and write to test files
	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tmpDir, file))
		assert.NoError(t, err)
		_, err = f.WriteString(testData)
		assert.NoError(t, err)
		f.Close()
	}

	// Create invalid CSV file
	invalidFile := filepath.Join(tmpDir, "invalid.csv")
	f, err := os.Create(invalidFile)
	assert.NoError(t, err)
	_, err = f.WriteString("invalid,csv\nformat,data")
	assert.NoError(t, err)
	f.Close()

	// Define test cases
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
		{
			name:      "Empty file list",
			files:     []string{},
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "Invalid CSV format",
			files:     []string{invalidFile},
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "Mix of valid and invalid files",
			files:     []string{filepath.Join(tmpDir, "bank1.csv"), invalidFile},
			startDate: "2024-01-01",
			endDate:   "2024-01-02",
			wantCount: 0,
			wantErr:   true,
		},
	}

	// Run each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the start and end dates
			start, err := time.Parse("2006-01-02", tt.startDate)
			assert.NoError(t, err)

			// Parse the end date
			end, err := time.Parse("2006-01-02", tt.endDate)
			assert.NoError(t, err)

			// Call the readBankStatements function
			statements, err := readBankStatements(tt.files, start, end)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			// Check if the result matches the expected result
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(statements))
		})
	}
}
