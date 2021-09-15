package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

type ToSend struct {
	Type string // type key
	Msg  string // only exported variables are sent, so start the ...
}

type bill struct {
	Type   string // type key
	Amount int    // only exported variables are sent, so start the ...
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return
		} else {
			fmt.Print("From Client:", string(msg))
			//titlemsg := strings.Title(msg)
			titlemsg := &ToSend{}
			titlemsg.Type = "message"
			titlemsg.Msg = msg
			fmt.Println(titlemsg)
			enc := json.NewEncoder(conn)
			enc.Encode(titlemsg)

			regning := &bill{}
			regning.Type = "bill"
			regning.Amount = 20
			enc.Encode(regning)

			//conn.Write([]byte(titlemsg))
		}
	}
}

func main() {
	fmt.Println("Listening for connection...")
	ln, _ := net.Listen("tcp", ":18081")
	defer ln.Close()
	conn, _ := ln.Accept()
	fmt.Println("Got a connection...")
	handleConnection(conn)
}
