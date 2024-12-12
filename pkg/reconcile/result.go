package reconcile

import (
	"encoding/json"
	"fmt"
	"os"
	"reconciliation/pkg/types"
	"strings"
)

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

type ReconcileUnmatched struct {
	// TransactionUnmatched is the number of transactions that were not matched to a bank statement
	TransactionUnmatched int

	// SystemUnmatched is the number of transactions that were not matched to a bank statement
	SystemUnmatched []types.Transaction

	// BankUnmatched is the number of transactions that were not matched to a system transaction
	BankUnmatched []types.BankStatement
}

func (r *ReconcileResult) String() string {
	var result strings.Builder
	result.WriteString("Reconciliation Summary:\n------------------------\n")
	fmt.Fprintf(&result, "Total transactions processed: %d\n", r.TransactionProcessed)
	fmt.Fprintf(&result, "Total matched transactions: %d\n", r.TransactionMatched)
	fmt.Fprintf(&result, "Total unmatched transactions: %d\n", r.TransactionUnmatched.TransactionUnmatched)

	if len(r.TransactionUnmatched.SystemUnmatched) > 0 {
		result.WriteString("\nSystem transactions missing from bank statements:\n")
		for _, tx := range r.TransactionUnmatched.SystemUnmatched {
			fmt.Fprintf(&result, "- TrxID: %s, Amount: %.2f, Date: %s\n",
				tx.TrxID,
				tx.Amount,
				tx.TransactionTime.Format("2006-01-02 15:04:05"))
		}
	}

	if len(r.TransactionUnmatched.BankUnmatched) > 0 {
		result.WriteString("\nBank statements missing from system transactions:\n")
		// Pre-allocate map with capacity
		bankGroups := make(map[string][]types.BankStatement, len(r.TransactionUnmatched.BankUnmatched))
		for _, stmt := range r.TransactionUnmatched.BankUnmatched {
			bankGroups[stmt.BankName] = append(bankGroups[stmt.BankName], stmt)
		}

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

	fmt.Fprintf(&result, "\nTotal amount discrepancies: %.2f\n", r.TotalDiscrepancies)
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

	result := jsonResult{}
	result.Summary.TotalTransactionsProcessed = r.TransactionProcessed
	result.Summary.TotalTransactionsMatched = r.TransactionMatched
	result.Summary.TotalTransactionsUnmatched = r.TransactionUnmatched.TransactionUnmatched
	result.Summary.TotalDiscrepancies = r.TotalDiscrepancies
	result.UnmatchedDetails.SystemTransactions = r.TransactionUnmatched.SystemUnmatched
	result.UnmatchedDetails.BankStatements = bankGroups

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
