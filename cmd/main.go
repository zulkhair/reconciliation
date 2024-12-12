package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	pkgcsv "reconciliation/pkg/csv"
	"reconciliation/pkg/reconcile"
	"reconciliation/pkg/types"
)

// rootCmd is the root command for the reconciliation tool
var rootCmd = &cobra.Command{
	Short: "A tool to reconcile system transactions with bank statements",
	RunE: func(cmd *cobra.Command, args []string) error {
		systemFile, _ := cmd.Flags().GetString("system")
		bankFile, _ := cmd.Flags().GetString("bank")
		startDate, _ := cmd.Flags().GetString("start")
		endDate, _ := cmd.Flags().GetString("end")
		print, _ := cmd.Flags().GetBool("print")

		// Validate required flags
		if systemFile == "" {
			return fmt.Errorf("system transaction file path is required")
		}
		if bankFile == "" {
			return fmt.Errorf("at least one bank statement file path is required")
		}
		if startDate == "" || endDate == "" {
			return fmt.Errorf("start and end dates are required")
		}

		// Parse dates
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return fmt.Errorf("invalid start date format. Use YYYY-MM-DD")
		}
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return fmt.Errorf("invalid end date format. Use YYYY-MM-DD")
		}

		// Validate date range
		if end.Before(start) {
			return fmt.Errorf("end date cannot be before start date")
		}

		// Start timer for read CSV
		startTimer := time.Now()

		// Read system transactions
		systemTransactions, err := readSystemTransactions(systemFile, start, end)
		if err != nil {
			return fmt.Errorf("failed to read system transactions: %w", err)
		}

		// Read bank statements
		bankFiles, err := processBankFiles(bankFile)
		if err != nil {
			return fmt.Errorf("failed to process bank files: %w", err)
		}
		bankStatements, err := readBankStatements(bankFiles, start, end)
		if err != nil {
			return fmt.Errorf("failed to read bank statements: %w", err)
		}

		// Stop timer for read CSV
		endTimer := time.Now()
		fmt.Printf("Read CSV time: %s\n", endTimer.Sub(startTimer))

		// Start timer for reconcile
		startTimer = time.Now()

		// Reconcile transactions
		result := reconcile.Reconcile(systemTransactions, bankStatements)
		if err != nil {
			return fmt.Errorf("failed to reconcile transactions: %w", err)
		}

		// Stop timer for reconcile
		endTimer = time.Now()
		fmt.Printf("Reconcile time: %s\n", endTimer.Sub(startTimer))

		// Start timer for generate result
		startTimer = time.Now()

		if print {
			// Print reconciled transactions
			fmt.Println(result.String())
		}

		// Generate JSON file
		outputFile, _ := cmd.Flags().GetString("output")
		if outputFile != "" {
			if err := result.GenerateJSON(outputFile); err != nil {
				return fmt.Errorf("failed to generate JSON file: %w", err)
			}
		}

		// Stop timer for generate result
		endTimer = time.Now()
		fmt.Printf("Generate result time: %s\n", endTimer.Sub(startTimer))

		return nil
	},
	SilenceErrors: true,
}

func main() {
	// Start timer
	start := time.Now()

	// Define command line flags
	rootCmd.Flags().StringP("system", "s", "", "Path to system transaction CSV file (required)")
	rootCmd.Flags().StringP("bank", "b", "", "Directory path contains bank statement CSV files or Comma-separated paths to bank statement CSV files (required)")
	rootCmd.Flags().StringP("start", "t", "", "Start date for reconciliation in YYYY-MM-DD format (required)")
	rootCmd.Flags().StringP("end", "e", "", "End date for reconciliation in YYYY-MM-DD format (required)")
	rootCmd.Flags().StringP("output", "o", "", "Path to output JSON file")
	rootCmd.Flags().BoolP("print", "p", false, "Print the result to the console")

	// Mark required flags
	err := rootCmd.MarkFlagRequired("system")
	if err != nil {
		fmt.Printf("Error: %s\n\n", err)
		os.Exit(1)
	}
	err = rootCmd.MarkFlagRequired("bank")
	if err != nil {
		fmt.Printf("Error: %s\n\n", err)
		os.Exit(1)
	}
	err = rootCmd.MarkFlagRequired("start")
	if err != nil {
		fmt.Printf("Error: %s\n\n", err)
		os.Exit(1)
	}
	err = rootCmd.MarkFlagRequired("end")
	if err != nil {
		fmt.Printf("Error: %s\n\n", err)
		os.Exit(1)
	}

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %s\n\n", err)
	}

	// Stop timer
	end := time.Now()
	fmt.Printf("Total execution time: %s\n", end.Sub(start))
}

// processBankFiles reads the bank statements from the given files
func processBankFiles(bankFileString string) ([]string, error) {
	// Check if path is a directory
	fileInfo, err := os.Stat(bankFileString)
	if err == nil {
		// If the bank file is a directory, read all CSV files in the directory
		if fileInfo.IsDir() {
			files, err := filepath.Glob(filepath.Join(bankFileString, "*.csv"))
			if err != nil {
				return nil, fmt.Errorf("failed to read bank files: %w", err)
			}
			return files, nil
		}
	}

	// Create separate paths from comma-separated string
	bankFiles := strings.Split(bankFileString, ",")
	for _, file := range bankFiles {
		_, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read bank files: %w", err)
		}
	}

	return bankFiles, nil
}

// readSystemTransactions reads the system transactions from the given file
func readSystemTransactions(systemFile string, start, end time.Time) ([]types.Transaction, error) {
	// Open the system file
	systemFileHandle, err := os.Open(systemFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open system file: %w", err)
	}
	defer systemFileHandle.Close()

	// Create a CSV reader with the system file
	systemReader := pkgcsv.NewCSVReader(
		csv.NewReader(systemFileHandle),
		pkgcsv.WithSkipHeader(true),
		pkgcsv.WithTimeRange(start, end),
	)

	// Read the system transactions
	systemTransactions, err := systemReader.ReadSystemTransactionsFromCSV()
	if err != nil {
		return nil, fmt.Errorf("failed to read system transactions: %w", err)
	}

	return systemTransactions, nil
}

// readBankStatements reads the bank statements from the given files
func readBankStatements(bankFiles []string, start, end time.Time) ([]types.BankStatement, error) {
	bankStatements := []types.BankStatement{}

	// Process files concurrently using worker pool
	type result struct {
		statements []types.BankStatement
		err        error
	}

	// Create a channel to receive results
	resultCh := make(chan result, len(bankFiles))

	// Create a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Process each bank file concurrently
	for _, bankFile := range bankFiles {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()

			bankFileHandle, err := os.Open(filename)
			if err != nil {
				resultCh <- result{nil, fmt.Errorf("failed to open bank file: %w", err)}
				return
			}
			defer bankFileHandle.Close()

			// Create a CSV reader with the bank file
			bankReader := pkgcsv.NewCSVReader(
				csv.NewReader(bankFileHandle),
				pkgcsv.WithSkipHeader(true),
				pkgcsv.WithTimeRange(start, end),
				pkgcsv.WithFilename(filename),
			)

			// Read the bank statements
			statements, err := bankReader.ReadBankStatementsFromCSV()
			if err != nil {
				resultCh <- result{nil, fmt.Errorf("failed to read bank statements: %w", err)}
				return
			}

			// Send the statements to the result channel
			resultCh <- result{statements, nil}
		}(bankFile)
	}

	// Close result channel once all goroutines complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	for res := range resultCh {
		if res.err != nil {
			return nil, res.err
		}
		bankStatements = append(bankStatements, res.statements...)
	}

	return bankStatements, nil
}
