package ledger

import (
	"fmt"
	"strconv"
	"sync"
)

type Transaction struct {
	Type   string
	ID     string
	From   string
	To     string
	Amount int
}

type Ledger struct {
	Type     string
	Accounts map[string]int
	lock     sync.Mutex
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (ledger *Ledger) Transaction(transaction Transaction) {
	ledger.lock.Lock()
	defer ledger.lock.Unlock()
	ledger.Accounts[transaction.From] -= transaction.Amount
	ledger.Accounts[transaction.To] += transaction.Amount
}

// Must be called to check if an account is in the system
func (ledger *Ledger) ContainsAccount(account string) bool {
	ledger.lock.Lock()
	_, found := ledger.Accounts[account]
	ledger.lock.Unlock()
	return found
}

func (ledger *Ledger) InsertAccount(account string) {
	ledger.lock.Lock()
	ledger.Accounts[account] = 0
	ledger.lock.Unlock()
}

func (ledger *Ledger) PrintLedger() {
	ledger.lock.Lock()
	for account, amount := range ledger.Accounts {
		fmt.Println("Account name: " + account + " amount: " + strconv.Itoa(amount))
	}
	ledger.lock.Unlock()
}
