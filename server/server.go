package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
)

type Message struct {
	sender  int
	message string
}

func handleError(err error) {
	// TODO: all
	// Deal with an error event.
	fmt.Printf("Error: %s\n", err)
}

func acceptConns(ln net.Listener, conns chan net.Conn) {
	for {
		//fmt.Println(ln)
		conn, err := ln.Accept()
		if err != nil {
			handleError(err)
		}
		conns <- conn
	}
}

func handleClient(client net.Conn, clientid int, msgs, unames, disconnect chan Message, names *map[int]string) {
	// TODO:
	// So long as this connection is alive:
	// Read in new messages as delimited by '\n's
	// Tidy up each message and add it to the messages channel,
	// recording which client it came from.
	reader := bufio.NewReader(client)
	username, _ := reader.ReadString('\n')

	unames <- Message{clientid, username[:len(username)-1]}
	for {
		msg, _ := reader.ReadString('\n')
		if msg == "<DISCONNECT>" {
			disconnect <- Message{clientid, "user '" + (*names)[clientid] + "' disconnected"}
		} else if msg != "" {
			msgs <- Message{clientid, msg[:len(msg)-1]}
		}
	}
}

func announce(message string, sender int, clients *map[int]net.Conn, names *map[int]string) {
	for key, entry := range *clients {
		if key != sender {
			fmt.Fprintln(entry, message)
			fmt.Println("sent '"+message+"' to", (*names)[key])
		}
	}
}

func main() {
	// Read in the network port we should listen on, from the commandline argument.
	// Default to port 8030
	portPtr := flag.String("port", "8030", "port to listen on")
	flag.Parse()
	ln, err := net.Listen("tcp", ":"+*portPtr)
	if err != nil {
		handleError(err)
	}

	conns := make(chan net.Conn)
	unames := make(chan Message)
	msgs := make(chan Message)
	disconnect := make(chan Message)
	clients := make(map[int]net.Conn)
	clientNames := make(map[int]string)

	go acceptConns(ln, conns)
	fmt.Println("accepting connections on local port " + *portPtr)

	cID := 0

	for {
		select {
		case conn := <-conns:
			go handleClient(conn, cID, msgs, unames, disconnect, &clientNames)
			clients[cID] = conn
			cID++
		case msg := <-unames: //the first message recieved from a client
			fmt.Println("user '"+msg.message+"' joined with id", msg.sender)
			clientNames[msg.sender] = msg.message
			announce("user '"+msg.message+"' joined", msg.sender, &clients, &clientNames)
			conn := clients[msg.sender]
			fmt.Fprintf(conn, "users connected: ")
			for key, entry := range clientNames {
				if msg.sender != key {
					fmt.Fprintf(conn, entry+" ")
				}
			}
			fmt.Fprintf(conn, "\n")
		case msg := <-msgs:
			fmt.Println("user '" + clientNames[msg.sender] + "' says '" + msg.message + "'")
			line := clientNames[msg.sender] + ": " + msg.message
			announce(line, msg.sender, &clients, &clientNames)
		case msg := <-disconnect:
			fmt.Println("<DISCONNECT> user '" + clientNames[msg.sender] + "'")
			delete(clients, msg.sender)     //remove conn
			delete(clientNames, msg.sender) //remove name
			announce(msg.message, msg.sender, &clients, &clientNames)
		}
	}
}
