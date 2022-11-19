package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"pichat/shared"
	"strings"
	"sync"
	"time"
)

type Message struct {
	sender  int
	message string
}

type Client struct {
	ID          int
	Conn        *net.Conn
	Name        string
	Acknowledge *chan bool
}

//func send(conn *net.Conn, s string, connLock)

func handleError(header string, err error) {
	// TODO: all
	// Deal with an error event.

	fmt.Printf("Error %s: %s\n", header, err)
}

func acceptConns(ln net.Listener, conns chan net.Conn) {
	for {
		//fmt.Println(ln)
		conn, err := ln.Accept()
		if err != nil {
			handleError("couldn't accept connection", err)
		}
		conns <- conn
	}
}

//normal disconnect (announced)
func disconnectClient(client Client, disconnect chan Message) {
	disconnect <- Message{client.ID, "user '" + client.Name + "' disconnected"}
}

func removeClient(id int, clients *map[int]Client, clientLock *sync.Mutex) {
	clientLock.Lock()
	defer clientLock.Unlock()
	delete(*clients, id) //remove conn
}

//30 seconds to respond with ack or disconnected
func timeout(client *Client, clients map[int]Client, disconnect chan Message, clientLock *sync.Mutex) {
	slow := make(chan bool)
	//clientCopy := *client
	go func() {
		time.Sleep(time.Second * 15)
		slow <- true
	}()
	//fmt.Println("waiting for "+shared.CtrlCode("acknowledge")+" from '"+clientCopy.Name+"':", clientCopy.ID)
	select {
	case <-*(client.Acknowledge):
		//fmt.Println("acknowledged")
		return
	case <-slow:
		if userExists(client.ID, &clients, clientLock) {
			//fmt.Println("timed out")
			fmt.Fprintln(*client.Conn, shared.CtrlCode("disconnect"))
			disconnectClient(*client, disconnect)
		} else {
			//fmt.Println("client '"+clientCopy.Name+"':", client.ID, "already disconnected")
		}
	}
}

func userExists(id int, clients *map[int]Client, clientLock *sync.Mutex) bool {
	clientLock.Lock()
	defer clientLock.Unlock()
	return (*clients)[id] != Client{Conn: nil, Name: "", Acknowledge: nil}
}

func handleClient(client net.Conn, id int, msgs, unames, disconnect chan Message, clients *map[int]Client, clientLock *sync.Mutex) {
	// TODO:
	// So long as this connection is alive:
	// Read in new messages as delimited by '\n's
	// Tidy up each message and add it to the messages channel,
	// recording which client it came from.
	reader := bufio.NewReader(client)
	username, _ := reader.ReadString('\n')
	username = shared.Sanitize(username)
	unames <- Message{id, username}

	for {
		msg, _ := reader.ReadString('\n')
		msg = shared.Sanitize(msg)
		fmt.Printf("message %s\n", msg)
		if !userExists(id, clients, clientLock) {
			return
		}
		clientLock.Lock()
		if len(msg) == 0 || msg == shared.CtrlCode("disconnect") {
			disconnectClient((*clients)[id], disconnect)
			clientLock.Unlock()
			return
		} else if msg == shared.CtrlCode("acknowledge") {
			*((*clients)[id].Acknowledge) <- true //we dont announce acknowledges else we would have a loop
		} else if msg != "" && msg != "\000" {
			msgs <- Message{id, msg}
		}
		clientLock.Unlock()
	}
}

func announce(message string, sender int, clients *map[int]Client, disconnect chan Message, clientLock *sync.Mutex) {
	clientLock.Lock()
	client := (*clients)[sender]
	go timeout(&client, *clients, disconnect, clientLock)
	for key, entry := range *clients {
		if key != sender {
			fmt.Fprintln(*entry.Conn, "\000"+message)
			fmt.Println("sent '"+message+"' to", (*clients)[key].Name)
		} else {
			fmt.Fprintln(*entry.Conn, shared.CtrlCode("acknowledge")) //letting client know we got their message
		}
	}
	clientLock.Unlock()
}

func main() {
	// Read in the network port we should listen on, from the commandline argument.
	// Default to port 8030
	portPtr := flag.String("port", "8030", "port to listen on")
	flag.Parse()
	ln, err := net.Listen("tcp", ":"+*portPtr)
	if err != nil {
		handleError("couldn't start listener", err)
	}

	conns := make(chan net.Conn)
	unames := make(chan Message)
	msgs := make(chan Message)
	disconnect := make(chan Message)
	clientLock := sync.Mutex{}
	clients := make(map[int]Client)

	go acceptConns(ln, conns)
	fmt.Println("accepting connections on local port " + *portPtr)

	cID := 0

	for {
		select {
		case conn := <-conns:
			go handleClient(conn, cID, msgs, unames, disconnect, &clients, &clientLock)
			ackCh := make(chan bool)
			clientLock.Lock()
			clients[cID] = Client{
				ID:          cID,
				Conn:        &conn,
				Name:        "",
				Acknowledge: &ackCh,
			}
			clientLock.Unlock()
			cID++

		case msg := <-unames: //the first message recieved from a client
			U := shared.CtrlCode("username")
			if strings.Contains(msg.message, U) {

				clientLock.Lock()
				c := clients[msg.sender]
				name := strings.Split(msg.message, U)[1]
				c.Name = name
				clients[msg.sender] = c
				clientLock.Unlock()

				fmt.Println(U+" '"+name+"' joined with id", msg.sender)
				announce("user '"+name+"' joined", msg.sender, &clients, disconnect, &clientLock)

				//time.Sleep(500 * time.Millisecond)
				go func() {
					onlineMessage := "users connected: "
					for key, entry := range clients {
						if msg.sender != key {
							onlineMessage += entry.Name + " "
						}
					}
					fmt.Fprintf(*c.Conn, onlineMessage+"\n")
					fmt.Println("sent online users")
				}()

			} else {

				fmt.Println(shared.CtrlCode("disconnect") + " (forced) on user '" + msg.message + "'")
				err := (*clients[msg.sender].Conn).Close()
				if err != nil {
					handleError("couldn't close connection to user", err)
				}
				removeClient(msg.sender, &clients, &clientLock)
			}

		case msg := <-msgs:
			if userExists(msg.sender, &clients, &clientLock) {
				client := clients[msg.sender]
				fmt.Println("user '" + client.Name + "' says '" + msg.message + "'")
				line := client.Name + ": " + msg.message
				announce(line, msg.sender, &clients, disconnect, &clientLock)
			}

		case msg := <-disconnect:
			fmt.Println(shared.CtrlCode("disconnect") + " user '" + clients[msg.sender].Name + "'")
			announce(msg.message, msg.sender, &clients, disconnect, &clientLock)
			err := (*clients[msg.sender].Conn).Close()
			if err != nil {
				handleError("couldn't close connection to user", err)
			}
			removeClient(msg.sender, &clients, &clientLock)
		}
	}
}
