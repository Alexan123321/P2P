/**
BY: Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 10-09-2021
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed chat implemented as structured P2P flooding network.
**/

package peer

import (
	"after_feedback/src/packages/ledger"
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

/* Message struct */
type msgType struct {
	Msg string
}

/******************/
/* MAKE INTERFACE */
/******************/

/* Peer struct */
type Peer struct {
	outIP            string
	outPort          string
	inIP             string
	inPort           string
	broadcast        chan string
	ln               net.Listener
	transactionsMade map[string]bool
	encLedger        map[net.Conn]gob.Encoder
	ledger           ledger.Ledger
	lock             sync.Mutex
	peers            []string
}

/* Initialize peer */
func (p *Peer) StartPeer() {
	fmt.Println("Please enter IP to connect to:")
	fmt.Scanln(&p.outIP)
	fmt.Println("Please enter port to connect to:")
	fmt.Scanln(&p.outPort)

	ln, _ := net.Listen("tcp", ":")
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	p.inPort = port
	p.inIP = "127.0.0.69"
	p.ln = ln
	p.broadcast = make(chan string)
	p.transactionsMade = make(map[string]bool)
	p.encLedger = make(map[net.Conn]gob.Encoder)
	p.ledger = ledger.MakeLedger()

	p.connect(p.ln, p.outIP, p.outPort)
}

/* Accept connection method */
func (p *Peer) connect(ln net.Listener, ip, port string) {
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		fmt.Println("Error at peer destination... Connecting to own network...")
		defer p.connect(p.ln, p.inIP, p.inPort)
		return
	}
	defer conn.Close()
	p.encLedger[conn] = *gob.NewEncoder(conn)
	p.printDetails()
	go p.read(conn)
	go p.write(conn)
	go p.broadcastMsg()
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a connection...")
		//Her skal du indsætte en decoder, så den anden peer kan sende sin ledger
		//eller bare forwarde ledger'eren til peer'en direkte...
		p.encLedger[conn] = *gob.NewEncoder(conn)
		go p.read(conn)
	}
}

/* Print details method of client */
func (p *Peer) printDetails() {
	ip, port, _ := net.SplitHostPort(p.ln.Addr().String())
	fmt.Println("Listening on address " + ip + ":" + port)
}

/* Read method of server */
func (p *Peer) read(conn net.Conn) {
	defer conn.Close()
	transaction := &ledger.Transaction{}
	dec := gob.NewDecoder(conn)

	for {
		err := dec.Decode(transaction)
		if err == io.EOF {
			p.acceptDisconnect(conn)
			return
		}

		if err != nil {
			log.Println(err.Error())
			return
		}
		p.handleTransaction(*transaction)
	}
}

/* Handle transaction */
func (p *Peer) handleTransaction(transaction ledger.Transaction) {
	if p.locateTransaction(transaction) == false {
		p.addTransaction(transaction)
		p.ledger.Transaction(transaction)
		defer p.ledger.PrintLedger()
	}
}

func (p *Peer) locateTransaction(transaction ledger.Transaction) bool {
	p.lock.Lock()
	_, found := p.transactionsMade[transaction.ID]
	p.lock.Unlock()
	return found
}

func (p *Peer) addTransaction(transaction ledger.Transaction) {
	p.lock.Lock()
	p.transactionsMade[transaction.ID] = true
	p.lock.Unlock()
}

/* Write method for client */
func (p *Peer) write(conn net.Conn) {
	fmt.Println("Please make transactions in the format: AMOUNT FROM TO followed by an empty character!")
	for {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		m, err := reader.ReadString('\n')
		if err != nil || m == "quit\n" {
			return
		}
		p.broadcast <- m
	}
}

/* Accept disconnect */
func (p *Peer) acceptDisconnect(conn net.Conn) {
	for conn, _ := range p.encLedger {
		if conn == conn {
			delete(p.encLedger, conn)
			return
		}
	}
	fmt.Println("Connection not found...")
	return
}

/* Broadcast handler */
func (p *Peer) broadcastMsg() {
	var i int
	for {
		inpString := strings.Split(<-p.broadcast, " ")
		amount, _ := strconv.Atoi(inpString[0])
		transaction := &ledger.Transaction{
			ID:     inpString[1] + strconv.Itoa(i),
			From:   inpString[1],
			To:     inpString[2],
			Amount: amount}
		i++
		for _, enc := range p.encLedger {
			enc.Encode(transaction)
		}
	}
}
