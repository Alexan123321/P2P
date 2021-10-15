/**
BY: Deyana Atanasova, Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 18-09-2021
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed transaction system implemented as structured P2P flooding network.
**/

package ledger

import (
	"fmt"
	"strconv"
	"sync"
)

/* Transaction struct */
type SignedTransaction struct {
	Type      string
	ID        string
	From      string
	To        string
	Amount    int
	Signature string
}

/* Ledger struct */
type Ledger struct {
	Type     string
	Accounts map[string]int
	lock     sync.Mutex
}

/* Ledger constructor */
func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

/* Transaction method */
func (ledger *Ledger) Transaction(transaction SignedTransaction) {
	ledger.lock.Lock()
	defer ledger.lock.Unlock()
	ledger.Accounts[transaction.From] -= transaction.Amount
	ledger.Accounts[transaction.To] += transaction.Amount
}

/* Print ledger method */
func (ledger *Ledger) PrintLedger() {
	ledger.lock.Lock()
	for account, amount := range ledger.Accounts {
		fmt.Println("Account name: " + account + " amount: " + strconv.Itoa(amount))
	}
	ledger.lock.Unlock()
}
