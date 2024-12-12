package reconcile

import (
	"math"
	"reconciliation/pkg/types"
)

// amountTolerance is the amount of discrepancy allowed
const amountTolerance = 0.01

// Reconcile reconciles the system transactions against the bank statements
func Reconcile(system []types.Transaction, bank []types.BankStatement) ReconcileResult {
	// Initialize the result
	result := ReconcileResult{
		TransactionUnmatched: ReconcileUnmatched{},
	}

	// Pre-allocate maps with expected capacity
	matchedSystem := make(map[string]bool, len(system))
	matchedBank := make(map[string]bool, len(bank))

	// Set the total number of transactions processed
	result.TransactionProcessed = len(system)

	// Compare each system transaction against bank statements
	for _, sysTx := range system {
		matched := false

		// Compare each system transaction against bank statements
		for _, bankTx := range bank {
			// Skip already matched bank transactions
			if matchedBank[bankTx.UniqueID] {
				continue
			}

			// Check if the system transaction matches the bank transaction
			if isMatch(sysTx, bankTx) {
				// Set the matched flag to true
				matched = true

				// Add the system transaction to the matched map
				matchedSystem[sysTx.TrxID] = true

				// Add the bank transaction to the matched map
				matchedBank[bankTx.UniqueID] = true

				// Increment the matched transaction count
				result.TransactionMatched++

				// Add any amount discrepancy to total
				result.TotalDiscrepancies += round(abs(sysTx.Amount - abs(bankTx.Amount)))

				// Break out of the loop
				break
			}
		}

		// If no match is found, add the system transaction to the unmatched map
		if !matched {
			result.TransactionUnmatched.TransactionUnmatched++
			result.TransactionUnmatched.SystemUnmatched = append(result.TransactionUnmatched.SystemUnmatched, sysTx)
		}
	}

	// Collect unmatched bank statements
	for _, bankTx := range bank {
		// Skip already matched bank transactions
		if matchedBank[bankTx.UniqueID] {
			continue
		}

		// Add the bank transaction to the unmatched map
		result.TransactionUnmatched.TransactionUnmatched++
		result.TransactionUnmatched.BankUnmatched = append(result.TransactionUnmatched.BankUnmatched, bankTx)
	}

	// Return the result
	return result
}

// isMatch checks if a system transaction matches a bank transaction
func isMatch(sysTx types.Transaction, bankTx types.BankStatement) bool {
	// Match by amount and transaction type
	bankAmount := bankTx.Amount

	// For system DEBIT transactions, bank amount should be negative
	// For system CREDIT transactions, bank amount should be positive
	if sysTx.Type == "DEBIT" && bankAmount > 0 {
		return false
	}
	if sysTx.Type == "CREDIT" && bankAmount < 0 {
		return false
	}

	if round(abs(sysTx.Amount-abs(bankAmount))) > amountTolerance {
		return false
	}

	// Match by date
	return sysTx.TransactionTime.Format("2006-01-02") == bankTx.Date.Format("2006-01-02")
}

// Assumes the value is only to 2 decimal places
func round(value float64) float64 {
	return math.Round(value*100) / 100
}

// abs returns the absolute value of a float64
func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
