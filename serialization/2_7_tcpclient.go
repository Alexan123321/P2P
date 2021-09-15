package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type message struct {
	Type string // type key
	Msg  string // only exported variables are sent, so start the ...
}

type bill struct {
	Type   string  // type key
	Amount float64 // only exported variables are sent, so start the ...
}

var conn net.Conn

func main() {
	conn, _ = net.Dial("tcp", "127.0.0.1:18081")
	defer conn.Close()
	go read(conn)
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		if text == "quit\n" {
			return
		}
		conn.Write([]byte(text))
	}
}

func read(conn net.Conn) {
	var temp map[string]interface{}
	dec := json.NewDecoder(conn)
	for {
		dec.Decode(&temp)
		parse(temp)
		//var object  = parse(temp)
	}
}

func parse(temp map[string]interface{}) interface{} {
	value, _ := temp["Type"]
	switch value {
	case "message":
		msg := &message{}
		msg.Type = temp["Type"].(string)
		msg.Msg = temp["Msg"].(string)
		fmt.Println("Message is: " + msg.Msg)
		return msg

	case "bill":
		bill := &bill{}
		bill.Type = temp["Type"].(string)
		bill.Amount = temp["Amount"].(float64)
		fmt.Println("Amount is: " + fmt.Sprint(bill.Amount))
		return bill

	default:
		fmt.Println("Error... Type conversion could not be performed...")
		return 0
	}
}
