package ledger

import (
	"fmt"
	"strconv"
	"sync"
)

type Transaction struct {
	ID     string
	From   string
	To     string
	Amount int
}

type Ledger struct {
	Accounts   map[string]int
	ledgerLock sync.Mutex
}

func MakeLedger() Ledger {
	ledger := Ledger{}
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (l *Ledger) Transaction(t Transaction) {
	l.ledgerLock.Lock()
	l.Accounts[t.From] -= t.Amount
	l.ledgerLock.Unlock()
	l.test(t)
}

func (l *Ledger) test(t Transaction) {
	l.ledgerLock.Lock()
	l.Accounts[t.To] += t.Amount
	l.ledgerLock.Unlock()
}

// Must be called to check if an account is in the system
func (l *Ledger) ContainsAccount(account string) bool {
	l.ledgerLock.Lock()
	_, found := l.Accounts[account]
	l.ledgerLock.Unlock()
	return found
}

func (l *Ledger) InsertAccount(account string) {
	l.ledgerLock.Lock()
	l.Accounts[account] = 0
	l.ledgerLock.Unlock()
}

func (l *Ledger) PrintLedger() {
	l.ledgerLock.Lock()
	for account, amount := range l.Accounts {
		fmt.Println("Account name: " + account + " amount: " + strconv.Itoa(amount))
	}
	l.ledgerLock.Unlock()
}
