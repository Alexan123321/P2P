/**
BY: Deyana Atanasova, Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 16-10-2021
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed transaction system implemented as structured P2P flooding network.
**/

package ledger

import (
	"after_feedback/src/packages/RSA"
	"crypto/sha256"
	"fmt"
	"hash"
	"math/big"
	"strconv"
	"sync"
)

/* Signed transaction struct */
type SignedTransaction struct {
	Type      string   // signedTransaction
	ID        string   // ID of the transaction
	From      string   // Sender of the transaction (public key)
	To        string   // Receiver of the transaction (public key)
	Amount    int      // Amount to transfer
	Signature *big.Int // Signature of the transaction
}

/* Computes the hash for some of the fields in a signed transaction */
func (signedTransaction SignedTransaction) ComputeHash() []byte {
	transactionHash := sha256.New()
	AddToHash(&transactionHash, signedTransaction.ID)
	AddToHash(&transactionHash, signedTransaction.From)
	AddToHash(&transactionHash, signedTransaction.To)
	AddToHash(&transactionHash, strconv.Itoa(signedTransaction.Amount))
	transactionHashSum := transactionHash.Sum(nil)
	return transactionHashSum[:]
}

func AddToHash(h *hash.Hash, str string) {
	if _, err := (*h).Write([]byte(str)); err != nil {
		panic(err)
	}
}

/* Generate RSA signature */
func (signedTransaction SignedTransaction) GenerateSignature(privateKeyString string) *big.Int {
	/* Hash transaction with SHA-256 and get integer representation of hash, */
	hashedMessage := RSA.ByteArrayToInt(signedTransaction.ComputeHash())

	/* Turn the string-encoded private key into key */
	privateKey := RSA.ToKey(privateKeyString)

	/* Encrypt the hashed message with the private key */
	ciphertext := RSA.Encrypt(hashedMessage, privateKey)

	/* Pad ciphertext with zeros */
	ciphertextInBytes := ciphertext.Bytes()
	keyInBytes := privateKey.N.Bytes()
	if len(ciphertextInBytes) < len(keyInBytes) {
		padding := make([]byte, len(keyInBytes)-len(ciphertextInBytes))
		ciphertextInBytes = append(padding, ciphertextInBytes...)
	}

	return new(big.Int).SetBytes(ciphertextInBytes)
}

func (signedTransaction *SignedTransaction) VerifySignedTransaction(peersMap map[string]string) bool {
	/* Hash transaction with SHA-256 and get integer representation of hash, */
	hashedMessage := RSA.ByteArrayToInt(signedTransaction.ComputeHash())

	/* Verify RSA signature */
	for _, publicKey := range peersMap {
		fmt.Println("Verifying for key " + publicKey)
		valid := RSA.VerifySignature(hashedMessage, signedTransaction.Signature, publicKey)
		if valid {
			return true
		}
	}
	return false
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
func (ledger *Ledger) Transaction(signedTransaction SignedTransaction) {
	ledger.lock.Lock()
	defer ledger.lock.Unlock()
	ledger.Accounts[signedTransaction.From] -= signedTransaction.Amount
	ledger.Accounts[signedTransaction.To] += signedTransaction.Amount
}

/* Print ledger method */
func (ledger *Ledger) PrintLedger() {
	ledger.lock.Lock()
	for account, amount := range ledger.Accounts {
		fmt.Println("Account name: " + account + " amount: " + strconv.Itoa(amount))
	}
	ledger.lock.Unlock()
}
