/**
BY: Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 10-09-2021
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed chat implemented as structured P2P flooding network.
**/

package peer

import ( "fmt" ; "net" ; "encoding/gob" ; "io" ; "log" ; "bufio" ; "os" )

/**
/ MESSAGE STRUCT
/ This struct contains all the information contained within the messages sent between peers.
/
/ Attributes:
/ - Msg string: the message being sent between peers.
**/
type msgType struct {
	Msg string
}

/**
/ PEER STRUCT
/ This struct defines the peer / node on the P2P network. 
/ As this struct is, in essence, both a client and a server, the attributes and the methods of the peer
/ have been described briefly below, associated to each component.
/
/ Attributes:
/ 	Attributes client side:
/ 	- outIP string: the IP of the peer being dialled. 
/ 	- outPort string: the port of the peer being dialled.
/
/ 	Attributes server side: 
/ 	- inIP string: the IP of the peer itself. 
/ 	- inPort string: the port of the peer itself. 
/	- inbound chan string: channel used for communication with peer(s) connected to the peer itself.
/ 	- outbound chan string: channel used for communication with peer to which the peer itself is connected.
/	- ln net.Listener: listener that listens on a specified port. TCP, in the present case.
/	- SentMessages map[string]bool: map used to store all strings that have been sent by the server, and
/	  confirm if the messages has been sent by the use of the boolean variable.
/	- encLedger map[net.Conn]gob.Encoder: map used to store all connections made to the server of the peer
/	  and mapping these to their respective encoders.
/
/ Methods:
/ - InitializePeer(): initializes the peer by initializing the client and the server.
/ - StartPeer(): starts the peer by starting the server of the peer and connecting to the other peer or itself.
/ 	
/	Methods client side:
/	- initializeClient(): initializes the client by asking the user to provide the IP and the port of the 
/	  peer the user wishes to connect to.
/	- connect(ip string, port string): connects the peer to the peer with the IP and port provided by the user.
/	  The client can ONLY connect to one peer at a time.
/	  IF the destination is faulty, the peer connects to itself.
/	- printDetails(): prints the IP and the port of the peer itself. 
/	- readClient(): reads messages received from the network.
/	- writeClient(): writes messages from the user to the network.
/	- unicastMsg(): unicasts a message written by the user to a SINGLE peer.
/
/	Methods server side:
/	- initializeServer(): initializes the server by:
/		1) setting up a TCP listener,
/		2) generate a port for the listener, 
/		3) store the port, the hardcoded IP and the listener,
/		4) initialize the out- and inbound channels,
/		5) make the map of sent messages and of all pairs of connections and encoders.
/	- startServer(): starts the server by running the broadcastMsg() and acceptConnection() methods concurrently.
/	- acceptConnection(ln net.Listener): accepts all requests for connection made by other peers. The server 
/	  can connect to MULTIPLE peers at a time.
/	- acceptDisconnect(conn net.Conn):
/	- readServer(conn net.Conn): reads all messages being sent to the server, and ensures that the client does
/	  not see previously received messages. 
/	- broadcastMsg(): broadcasts a message to all clients connected to the peer. 
**/



type Peer struct {
	/* Client side of peer */
	outIP string
	outPort string

	/* Server side of peer */
	inIP string	
	inPort string																
	inbound chan string
	outbound chan string
	ln net.Listener																																							
	SentMessages map[string]bool
	encLedger map[net.Conn]gob.Encoder 	
}

/*******************************/
/******** Peer methods *********/
/*******************************/

/* Initialize peer */
func (p *Peer) InitializePeer() {
	p.initializeClient()
	p.initializeServer()
}

/* Start peer method */
func (p *Peer) StartPeer() {
	go p.startServer()
	defer p.connect(p.outIP, p.outPort)
}

/*******************************/
/* Client side of peer methods */
/*******************************/

/* Initialize client */
func (p *Peer) initializeClient() {
	fmt.Println("Please enter IP to connect to:")
	fmt.Scanln(&p.outIP)
	fmt.Println("Please enter port to connect to:")
	fmt.Scanln(&p.outPort)
}

/* Connect to peer method */
func (p *Peer) connect(ip string, port string) {
	conn, err := net.Dial("tcp", ip + ":" + port)
	if(err != nil) {
		fmt.Println("Error at peer destination... Connecting to own network...")
		defer p.connect(p.inIP, p.inPort)
		return
	}
	defer conn.Close()
	p.printDetails()
	go p.readClient(conn)
	go p.writeClient(conn)
	go p.unicastMsg(conn)
	for {}
}

/* Print details method of client */
func (p *Peer) printDetails() {
	name, _ := os.Hostname()
	addrs, _ := net.LookupHost(name)
	for _, addr := range addrs {
	fmt.Println("My IP address: " + addr)
	fmt.Println("My port: " + p.inPort)
	}
}

/* Read method of client */
func (p *Peer) readClient(conn net.Conn) {
	msg := &msgType{}
	dec := gob.NewDecoder(conn)
	for {
		err := dec.Decode(msg)
		if err != nil { 
			return 
		}
		if value, _ := p.SentMessages[msg.Msg]; value == false {
			p.inbound<- msg.Msg 
		}
		fmt.Println(msg.Msg)
	}
}

/* Write method for client */
func (p *Peer) writeClient(conn net.Conn) {
	for {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		m, err := reader.ReadString('\n')
		if err != nil || m == "quit\n" { 
			return 
		}
		p.outbound<- m 
	}
}

/* Unicast method */
func (p *Peer) unicastMsg(conn net.Conn) {
	msgToSend := &msgType{}
	enc := gob.NewEncoder(conn)
	for {
		msg := <-p.outbound
		msgToSend.Msg = msg
		enc.Encode(msgToSend)
	}
}

/*******************************/
/* Server side of peer methods */
/*******************************/

/* Initialize server */
func (p *Peer) initializeServer() {
	ln, _ := net.Listen("tcp", ":")										
	_, port, _ := net.SplitHostPort(ln.Addr().String())					
	p.inPort = port														
	p.inIP = "127.0.0.69"									
	p.ln = ln
	p.outbound = make(chan string)
	p.inbound = make(chan string)
	p.SentMessages = make(map[string]bool)
	p.encLedger = make(map[net.Conn]gob.Encoder)
}

/* Start new network method */
func (p *Peer) startServer() {
	go p.broadcastMsg()
	go p.acceptConnection(p.ln)
}

/* Accept connection method */
func (p *Peer) acceptConnection(ln net.Listener) {
	defer ln.Close()														
	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a connection...")
		p.encLedger[conn] = *gob.NewEncoder(conn)
		go p.readServer(conn)
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

/* Read method of server */
func (p *Peer) readServer(conn net.Conn) {
	defer conn.Close()
	msg := &msgType{}
	dec := gob.NewDecoder(conn)

	for {
		err := dec.Decode(msg)
		if (err == io.EOF) {
			p.acceptDisconnect(conn)
			return
		}

		if (err != nil) {
			log.Println(err.Error())
			return
		}

		if _, found := p.SentMessages[msg.Msg]; found == false {
			p.outbound<- msg.Msg 
			p.inbound<- msg.Msg 
			p.SentMessages[msg.Msg] = true
			fmt.Println("Message: " + msg.Msg + " is added to SentMessages.")
		}
	}
}

/* Broadcast handler */
func (p *Peer) broadcastMsg() {
	msgToSend := &msgType{}
	for {
		msgToSend.Msg = <-p.inbound
		for _, enc := range p.encLedger {
			enc.Encode(msgToSend)
		}
	}
}