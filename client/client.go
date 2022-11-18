package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const width = 40

func read(conn net.Conn, showUser chan bool) {
	//TODO In a continuous loop, read a message from the server and display it.
	for {
		msg, _ := bufio.NewReader(conn).ReadString('\n')
		spaces := width - len(msg) - 1
		if spaces < 0 {
			spaces = 0
		}
		fmt.Println("\r" + msg[:len(msg)-1] + strings.Repeat(" ", spaces))
		showUser <- true
	}
}

func write(conn net.Conn, showUser chan bool) {
	//TODO Continually get input from the user and send messages to the server.
	r := bufio.NewReader(os.Stdin)
	msg, _ := r.ReadString('\n')
	fmt.Fprintf(conn, msg)
	showUser <- true
}

func username(username string, showUser chan bool) {
	for {
		<-showUser
		fmt.Print(username + ": ")
	}
}

func cleanup(conn net.Conn) {
	fmt.Fprintf(conn, "<DISCONNECT>")
}

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	pUsername := flag.String("uname", "default_user", "Your alias")
	pAddr := flag.String("server", "pichat.ddns.net:8030", "The chat server")
	flag.Parse()
	conn, err := net.Dial("tcp", *pAddr)
	if err != nil {
		fmt.Printf("Error Dialing server: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(conn, *pUsername)
	go func() {
		<-c
		cleanup(conn)
		os.Exit(1)
	}()

	showUser := make(chan bool)
	fmt.Print(*pUsername + ": ")

	go read(conn, showUser)
	go username(*pUsername, showUser)

	for {
		write(conn, showUser)
	}
}
