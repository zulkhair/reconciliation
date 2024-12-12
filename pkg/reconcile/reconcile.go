package reconcile

import (
	"math"
	"reconciliation/pkg/types"
)

// amountTolerance is the amount of discrepancy allowed
const amountTolerance = 0.01

func Reconcile(system []types.Transaction, bank []types.BankStatement) ReconcileResult {
	result := ReconcileResult{
		TransactionUnmatched: ReconcileUnmatched{},
	}

	// Pre-allocate maps with expected capacity
	matchedSystem := make(map[string]bool, len(system))
	matchedBank := make(map[string]bool, len(bank))

	result.TransactionProcessed = len(system)

	// Compare each system transaction against bank statements
	for _, sysTx := range system {
		matched := false
		for _, bankTx := range bank {
			// Skip already matched bank transactions
			if matchedBank[bankTx.UniqueID] {
				continue
			}

			if isMatch(sysTx, bankTx) {
				matched = true
				matchedSystem[sysTx.TrxID] = true
				matchedBank[bankTx.UniqueID] = true
				result.TransactionMatched++

				// Add any amount discrepancy to total
				result.TotalDiscrepancies += round(abs(sysTx.Amount - abs(bankTx.Amount)))
				break
			}
		}

		if !matched {
			result.TransactionUnmatched.TransactionUnmatched++
			result.TransactionUnmatched.SystemUnmatched = append(result.TransactionUnmatched.SystemUnmatched, sysTx)
		}
	}

	// Collect unmatched bank statements
	for _, bankTx := range bank {
		if !matchedBank[bankTx.UniqueID] {
			result.TransactionUnmatched.TransactionUnmatched++
			result.TransactionUnmatched.BankUnmatched = append(result.TransactionUnmatched.BankUnmatched, bankTx)
		}
	}

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
