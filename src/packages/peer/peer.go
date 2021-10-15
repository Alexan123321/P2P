/**
BY: Deyana Atanasova, Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 18-09-2021
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed transaction system implemented as structured P2P flooding network.
**/

package peer

import (
	"after_feedback/src/packages/ledger"
	"after_feedback/src/packages/rsa"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const MAX_CON = 10
const e = 3

/* Message struct containing list of peers */
type Msg_peerslist struct {
	Type string
	List []string
	//ListwKeys map[string][string]
}

/* Message struct containing address of new peer */
type Msg_newpeer struct {
	Type    string
	Address string
	//PublicKey string
}

/* Peer struct */
type Peer struct {
	outIP            string
	outPort          string
	inIP             string
	inPort           string
	broadcast        chan []byte
	ln               net.Listener
	transactionsMade map[string]bool
	connections      map[string]net.Conn
	ledger           *ledger.Ledger
	lock             sync.Mutex
	peers            Msg_peerslist
	publicKey        rsa.Key
	privateKey       rsa.Key
}

/* Initialize peer method */
func (peer *Peer) StartPeer() {
	/* User input */
	fmt.Println("Please enter IP to connect to:")
	fmt.Scanln(&peer.outIP)
	fmt.Println("Please enter port to connect to:")
	fmt.Scanln(&peer.outPort)

	/* Initialize variables */
	ln, _ := net.Listen("tcp", "127.0.0.1:")
	ip, port, _ := net.SplitHostPort(ln.Addr().String())
	peer.ln = ln
	peer.inIP = ip
	peer.inPort = port
	peer.broadcast = make(chan []byte)
	peer.transactionsMade = make(map[string]bool)
	peer.connections = make(map[string]net.Conn, 0)
	peer.ledger = ledger.MakeLedger()
	peer.peers.Type = "peerlist"
	peer.peers.List = make([]string, 0)
	peer.publicKey, peer.privateKey = rsa.KeyGen(rsa.GenerateRandomK, e)

	/* Print address for connectivity */
	peer.printDetails()

	/* Initialize connection and routines */
	peer.connect(peer.outIP + ":" + peer.outPort)
	go peer.write()
	go peer.broadcastMsg()
	go peer.acceptConnect()
}

/* Accept connection method */
func (peer *Peer) connect(address string) {
	/* Check if the peers are already connected */
	for addresses, _ := range peer.connections {
		if addresses == address {
			fmt.Println("Already connected to peer: " + address)
			return
		}
	}
	/* Otherwise, dial the connection */
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error at peer destination. Connecting to own network...")
		defer peer.connect(peer.inIP + ":" + peer.inPort)
		return
	}
	/* Store the connection for broadcasting */
	peer.connections[address] = conn

	/* Initialize reading routine associated with the conenction */
	go peer.read(conn)
}

/* Accept connect method */
func (peer *Peer) acceptConnect() {
	for {
		/* Accept connection that dials */
		conn, _ := peer.ln.Accept()
		peer.connections[conn.RemoteAddr().String()] = conn
		fmt.Println("Got a connection from " + conn.RemoteAddr().String())
		defer peer.ln.Close()

		/* Forward local list of peers */
		jsonString, _ := json.Marshal(peer.peers)
		conn.Write(jsonString)

		/* Start reading input from the connection */
		go peer.read(conn)
	}
}

/* Accept disconnect */
func (peer *Peer) acceptDisconnect(conn net.Conn) {
	/* Locate address and remove it */
	for address, conn := range peer.connections {
		if conn == conn {
			delete(peer.connections, address)
			return
		}
	}
	fmt.Println("Connection not found...")
	return
}

/* Read method of server */
func (peer *Peer) read(conn net.Conn) {
	defer conn.Close()
	/* Decode every message into a string-interface map */
	var temp map[string]interface{}
	decoder := json.NewDecoder(conn)
	for {
		err := decoder.Decode(&temp)
		/* In case of empty string, disconnect the peer */
		if err == io.EOF {
			peer.acceptDisconnect(conn)
			return
		}
		/* In case of an error, crash the peer */
		if err != nil {
			log.Println(err.Error())
			return
		}
		/* Forward the map to the handleRead method */
		peer.handleRead(temp)
	}
}

/* Handleread method */
func (peer *Peer) handleRead(temp map[string]interface{}) {
	/* Reads the type of the object received and activates appropriate switch-statement */
	jsonString, _ := json.Marshal(temp)
	objectType, _ := temp["Type"]
	switch objectType {
	case "peerlist":
		peers := &Msg_peerslist{}
		json.Unmarshal(jsonString, &peers)
		peer.handlePeerList(*peers)
		return
	case "transaction":
		transaction := &ledger.SignedTransaction{}
		json.Unmarshal(jsonString, &transaction)
		peer.handleTransaction(*transaction)
		return
	case "newpeer":
		newPeer := &Msg_newpeer{}
		json.Unmarshal(jsonString, &newPeer)
		peer.handleNewPeer(*newPeer)
	default:
		fmt.Println("Error... Type conversion could not be performed...")
		return
	}
}

/* Handle peer list method */
func (peer *Peer) handlePeerList(list Msg_peerslist) {
	/* If peer already has a list, return */
	if len(peer.peers.List) != 0 {
		return
	}
	/* Otherwise store the received list */
	peer.peers = list
	/* If more there are more than 10 peers on list,
	connect to the 10 peers before itself */
	if MAX_CON < len(peer.peers.List) {
		for index := len(peer.peers.List) - 10; index == len(peer.peers.List); index++ {
			peer.connect(peer.peers.List[index])
		}
		/* Otherwise connect to all peers on the list */
	} else {
		for _, address := range peer.peers.List {
			peer.connect(address)
		}
	}
	/* Then append itself */
	peer.peers.List = append(peer.peers.List, peer.inIP+":"+peer.inPort)
	/* As the peer only handles a list of peers, it is new on the network,
	it broadcasts its presence after having connectde to the previous 10 peers */
	newPeer := &Msg_newpeer{Type: "newpeer"}
	newPeer.Address = peer.inIP + ":" + peer.inPort
	jsonString, _ := json.Marshal(newPeer)
	peer.broadcast <- jsonString
}

/* Handle new peer method */
func (peer *Peer) handleNewPeer(newPeer Msg_newpeer) {
	for _, address := range peer.peers.List {
		/* If the peer is already on the local list of peers, do nothing */
		if newPeer.Address == address {
			return
			/* If not, append it to the list of peers */
		} else {
			peer.peers.List = append(peer.peers.List, address)
		}
	}
}

/* Received when a transaction is made */
func (peer *Peer) handleTransaction(transaction ledger.SignedTransaction) {
	/* If the transaction has not been processed, then */
	if peer.locateTransaction(transaction) == false {
		/* add it to the list of transactionsMade and broadcast it */
		peer.addTransaction(transaction)
		peer.ledger.Transaction(transaction)
		defer peer.ledger.PrintLedger()
		jsonString, _ := json.Marshal(transaction)
		peer.broadcast <- jsonString
	}
	/* If the transaction has been processed, do nothing */
	return
}

/* Write method for client */
func (peer *Peer) write() {
	fmt.Println("Please make transactions in the format: AMOUNT FROM TO followed by an empty character!")
	var i int
	for {
		/* Read transaction string from user */
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		m, err := reader.ReadString('\n')
		if err != nil || m == "quit\n" {
			return
		}
		/* Split the string into the amount, from and to */
		inpString := strings.Split(m, " ")
		/* Make transaction object from the details, */
		amount, _ := strconv.Atoi(inpString[0])
		transaction := &ledger.SignedTransaction{Type: "transaction"}
		transaction.ID = inpString[1] + strconv.Itoa(i) + strconv.Itoa(rand.Intn(100))
		transaction.From = inpString[1]
		transaction.To = inpString[2]
		transaction.Amount = amount

		/* and broadcast it */
		jsonString, _ := json.Marshal(transaction)
		peer.broadcast <- jsonString
		i++
	}
}

/* Broadcast method */
func (peer *Peer) broadcastMsg() {
	for {
		jsonString := <-peer.broadcast
		for _, con := range peer.connections {
			con.Write(jsonString)
		}
	}
}

/* Print details method */
func (peer *Peer) printDetails() {
	ip, port, _ := net.SplitHostPort(peer.ln.Addr().String())
	fmt.Println("Listening on address " + ip + ":" + port)
}

/* Locate transaction method */
func (peer *Peer) locateTransaction(transaction ledger.SignedTransaction) bool {
	peer.lock.Lock()
	_, found := peer.transactionsMade[transaction.ID]
	peer.lock.Unlock()
	return found
}

/* Add transaction method */
func (peer *Peer) addTransaction(transaction ledger.SignedTransaction) {
	peer.lock.Lock()
	peer.transactionsMade[transaction.ID] = true
	peer.lock.Unlock()
}

/* TODO:
ACCEPTDISCONNECT: remove connection from the ledger and publish the crash of the node.
LATECOMER: when a peer is joining the network late, it must receive the list of transactionsMade.
*/

/* TODO:
/* UPDATE STRUCTS */
/*
1) Update transaction such that includes a signature string
2) Update peerlist such that it now contains a map, mapping peer addresses to public keys
3) Update newpeer to also include a public key

/* UPDATE HANDLER METHODS */
/*
1) When a peerlist is recieved, the handler needs to be corrected to accomodate the address to public key map
2) When a newpeer is announced, the handler needs to accommodate the address to public key map
3) When a transaction is instantiated by a client via. user input, the client inserts its PUBLIC KEY in FROM-field,
the receivers PUBLIC KEY in the TO-field, and its address in the SIGNATURE field. The client then ENCRYPTS all fields
using its PRIVATE KEY and broadcasts the transaction.
4) When a transaction is received, the signature is DECRYPTED using all PUBLIC KEYS available in the peerlist
map. IF a decryption results in an address corresponding to an address in the map, the transaction is verified,
and can be made.
*/

/* UPDATE PEER INITIALIZATION */
/* New variables and stuff.
 */
