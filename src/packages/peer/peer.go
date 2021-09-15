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
	ledger           *ledger.Ledger
	lock             sync.Mutex
	peers            []string
}

/* Initialize peer */
func (peer *Peer) StartPeer() {
	fmt.Println("Please enter IP to connect to:")
	fmt.Scanln(&peer.outIP)
	fmt.Println("Please enter port to connect to:")
	fmt.Scanln(&peer.outPort)

	ln, _ := net.Listen("tcp", "127.0.0.1:")
	ip, port, _ := net.SplitHostPort(ln.Addr().String())
	peer.ln = ln
	peer.inIP = ip
	peer.inPort = port
	peer.broadcast = make(chan string)
	peer.transactionsMade = make(map[string]bool)
	peer.encLedger = make(map[net.Conn]gob.Encoder)
	peer.ledger = ledger.MakeLedger()

	peer.connect(peer.ln, peer.outIP, peer.outPort)
}

/* Accept connection method */
func (peer *Peer) connect(ln net.Listener, ip, port string) {
	fmt.Println("Attempting connection to peer " + ip + ":" + port)
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		fmt.Println("Error at peer destination. Connecting to own network...")
		defer peer.connect(peer.ln, peer.inIP, peer.inPort)
		return
	}
	defer conn.Close()
	peer.encLedger[conn] = *gob.NewEncoder(conn)
	peer.printDetails()
	go peer.read(conn)
	go peer.write(conn)
	go peer.broadcastMsg()
	defer ln.Close()
	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a connection from " + conn.RemoteAddr().String())
		//Her skal du indsætte en decoder, så den anden peer kan sende sin ledger
		//eller bare forwarde ledger'eren til peer'en direkte...
		peer.addToPeers(conn.RemoteAddr().String())
		peer.encLedger[conn] = *gob.NewEncoder(conn)
		go peer.read(conn)
	}
}

/* Print details method of client */
func (peer *Peer) printDetails() {
	ip, port, _ := net.SplitHostPort(peer.ln.Addr().String())
	fmt.Println("Listening on address " + ip + ":" + port)
}

/* Read method of server */
func (peer *Peer) read(conn net.Conn) {
	defer conn.Close()
	transaction := &ledger.Transaction{}
	dec := gob.NewDecoder(conn)

	for {
		err := dec.Decode(transaction)
		if err == io.EOF {
			peer.acceptDisconnect(conn)
			return
		}

		if err != nil {
			log.Println(err.Error())
			return
		}
		peer.handleTransaction(*transaction)
	}
}

/* Handle transaction */
func (peer *Peer) handleTransaction(transaction ledger.Transaction) {
	if peer.locateTransaction(transaction) == false {
		peer.addTransaction(transaction)
		peer.ledger.Transaction(transaction)
		defer peer.ledger.PrintLedger()
	}
}

func (peer *Peer) locateTransaction(transaction ledger.Transaction) bool {
	peer.lock.Lock()
	_, found := peer.transactionsMade[transaction.ID]
	peer.lock.Unlock()
	return found
}

func (peer *Peer) addTransaction(transaction ledger.Transaction) {
	peer.lock.Lock()
	peer.transactionsMade[transaction.ID] = true
	peer.lock.Unlock()
}

/* Write method for client */
func (peer *Peer) write(conn net.Conn) {
	fmt.Println("Please make transactions in the format: AMOUNT FROM TO followed by an empty character!")
	for {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		m, err := reader.ReadString('\n')
		if err != nil || m == "quit\n" {
			return
		}
		peer.broadcast <- m
	}
}

/* Accept disconnect */
func (peer *Peer) acceptDisconnect(conn net.Conn) {
	for conn, _ := range peer.encLedger {
		if conn == conn {
			delete(peer.encLedger, conn)
			return
		}
	}
	fmt.Println("Connection not found...")
	return
}

/* Broadcast handler */
func (peer *Peer) broadcastMsg() {
	var i int
	for {
		inpString := strings.Split(<-peer.broadcast, " ")
		amount, _ := strconv.Atoi(inpString[0])
		transaction := &ledger.Transaction{
			ID:     inpString[1] + strconv.Itoa(i),
			From:   inpString[1],
			To:     inpString[2],
			Amount: amount}
		i++
		for _, enc := range peer.encLedger {
			enc.Encode(transaction)
		}
	}
}

func (peer *Peer) addToPeers(address string) {
	peer.peers = append(peer.peers, address)
	fmt.Printf("List of peers: %v\n", peer.peers)
}
