/**
BY: Deyana Atanasova, Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 16-10-2021
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed transaction system implemented as structured P2P flooding network.
**/

package ledger

import (
	"after_feedback/src/packages/RSA"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"sync"
)

type Transaction struct {
	ID     string  // ID of the transaction
	From   RSA.Key // Sender of the transaction(RSA encoded public key)
	To     RSA.Key // Receiver of the transaction(RSA encoded public key)
	Amount int     // Amount to transfer
}

/* Signed transaction struct */
type SignedTransaction struct {
	Type        string      // signedTransaction
	Transaction Transaction // Transaction
	Signature   *big.Int    // Signature of the transaction
}

func (transaction *Transaction) ToBytes() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(transaction)
	if err != nil {
		log.Println(err.Error())
	}
	return buf.Bytes()
}

/* Ledger struct */
type Ledger struct {
	Type     string
	Accounts map[RSA.Key]int
	lock     sync.Mutex
}

/* Ledger constructor */
func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[RSA.Key]int)
	return ledger
}

/* Transaction method */
func (ledger *Ledger) Transaction(signedTransaction SignedTransaction) {
	ledger.lock.Lock()
	defer ledger.lock.Unlock()
	ledger.Accounts[signedTransaction.Transaction.From] -= signedTransaction.Transaction.Amount
	ledger.Accounts[signedTransaction.Transaction.To] += signedTransaction.Transaction.Amount
}

/* Print ledger method */
func (ledger *Ledger) PrintLedger() {
	ledger.lock.Lock()
	for account, amount := range ledger.Accounts {
		fmt.Println("Account name: " + account.N.String() + " amount: " + strconv.Itoa(amount))
	}
	ledger.lock.Unlock()
}

func (signedTransaction *SignedTransaction) VerifySignedTransaction(peersMap map[string]RSA.Key) bool {
	/* Hash transaction with SHA-256 and get integer representation of hash, */
	hashedMessage := RSA.ByteArrayToInt(RSA.HashMessage(signedTransaction.Transaction.ToBytes()))
	/* Verify RSA signature */
	for _, publicKey := range peersMap {
		valid := RSA.VerifySignature(hashedMessage, signedTransaction.Signature, publicKey)
		if valid {
			return true
		}
	}
	return false
}
