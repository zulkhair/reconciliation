package reconcile

import (
	"encoding/json"
	"fmt"
	"os"
	"reconciliation/pkg/types"
	"strings"
)

// ReconcileResult is the result of the reconciliation process
type ReconcileResult struct {
	// TransactionProcessed is the number of transactions that were processed
	TransactionProcessed int

	// TransactionMatched is the number of transactions that were matched
	TransactionMatched int

	// TransactionUnmatched is the details of transactions that were not matched
	TransactionUnmatched ReconcileUnmatched

	// TotalDiscrepancies is sum of absolute differences in amount between matched transactions
	TotalDiscrepancies float64
}

// ReconcileUnmatched is the details of transactions that were not matched
type ReconcileUnmatched struct {
	// TransactionUnmatched is the number of transactions that were not matched to a bank statement
	TransactionUnmatched int

	// SystemUnmatched is the number of transactions that were not matched to a bank statement
	SystemUnmatched []types.Transaction

	// BankUnmatched is the number of transactions that were not matched to a system transaction
	BankUnmatched []types.BankStatement
}

// String returns a string representation of the reconciliation result
func (r *ReconcileResult) String() string {
	// Initialize a new strings.Builder
	var result strings.Builder

	// Write the summary header
	result.WriteString("Reconciliation Summary:\n------------------------\n")

	// Write the total transactions processed
	fmt.Fprintf(&result, "Total transactions processed: %d\n", r.TransactionProcessed)

	// Write the total matched transactions
	fmt.Fprintf(&result, "Total matched transactions: %d\n", r.TransactionMatched)

	// Write the total unmatched transactions
	fmt.Fprintf(&result, "Total unmatched transactions: %d\n", r.TransactionUnmatched.TransactionUnmatched)

	// Write the system transactions missing from bank statements
	if len(r.TransactionUnmatched.SystemUnmatched) > 0 {
		result.WriteString("\nSystem transactions missing from bank statements:\n")
		for _, tx := range r.TransactionUnmatched.SystemUnmatched {
			fmt.Fprintf(&result, "- TrxID: %s, Amount: %.2f, Type: %s, Date: %s\n",
				tx.TrxID,
				tx.Amount,
				tx.Type,
				tx.TransactionTime.Format("2006-01-02 15:04:05"))
		}
	}

	// Write the bank statements missing from system transactions
	if len(r.TransactionUnmatched.BankUnmatched) > 0 {
		result.WriteString("\nBank statements missing from system transactions:\n")

		// Pre-allocate map with capacity
		bankGroups := make(map[string][]types.BankStatement, len(r.TransactionUnmatched.BankUnmatched))
		for _, stmt := range r.TransactionUnmatched.BankUnmatched {
			bankGroups[stmt.BankName] = append(bankGroups[stmt.BankName], stmt)
		}

		// Write the bank statements missing from system transactions
		for bankName, statements := range bankGroups {
			fmt.Fprintf(&result, "\nBank: %s\n", bankName)
			for _, stmt := range statements {
				fmt.Fprintf(&result, "- ID: %s, Amount: %.2f, Date: %s\n",
					stmt.UniqueID,
					stmt.Amount,
					stmt.Date.Format("2006-01-02"))
			}
		}
	}

	// Write the total amount discrepancies
	fmt.Fprintf(&result, "\nTotal amount discrepancies: %.2f\n", r.TotalDiscrepancies)

	// Return the result as a string
	return result.String()
}

// GenerateJSON generates a JSON file containing reconciliation results
func (r *ReconcileResult) GenerateJSON(filename string) error {
	// Define the result structure at package level to avoid recreating it
	type jsonResult struct {
		Summary struct {
			TotalTransactionsProcessed int     `json:"total_transactions_processed"`
			TotalTransactionsMatched   int     `json:"total_transactions_matched"`
			TotalTransactionsUnmatched int     `json:"total_transactions_unmatched"`
			TotalDiscrepancies         float64 `json:"total_discrepancies"`
		} `json:"summary"`
		UnmatchedDetails struct {
			SystemTransactions []types.Transaction              `json:"system_transactions,omitempty"`
			BankStatements     map[string][]types.BankStatement `json:"bank_statements,omitempty"`
		} `json:"unmatched_details"`
	}

	// Pre-allocate map with capacity
	bankGroups := make(map[string][]types.BankStatement, len(r.TransactionUnmatched.BankUnmatched))
	for _, stmt := range r.TransactionUnmatched.BankUnmatched {
		bankGroups[stmt.BankName] = append(bankGroups[stmt.BankName], stmt)
	}

	// Initialize the result
	result := jsonResult{}

	// Set the summary values
	result.Summary.TotalTransactionsProcessed = r.TransactionProcessed
	result.Summary.TotalTransactionsMatched = r.TransactionMatched
	result.Summary.TotalTransactionsUnmatched = r.TransactionUnmatched.TransactionUnmatched
	result.Summary.TotalDiscrepancies = r.TotalDiscrepancies

	// Set the unmatched details
	result.UnmatchedDetails.SystemTransactions = r.TransactionUnmatched.SystemUnmatched
	result.UnmatchedDetails.BankStatements = bankGroups

	// Create the JSON file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Set the JSON encoder to use indentation
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// Encode the result
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
