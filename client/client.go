package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"pichat/shared"
	"strings"
	"syscall"
)

const width = 40

func acknowledge(conn net.Conn) {
	fmt.Fprintln(conn, shared.CtrlCode("acknowledge"))
}

func read(conn net.Conn, showUser chan bool) {
	//TODO In a continuous loop, read a message from the server and display it.
	for {
		msg, _ := bufio.NewReader(conn).ReadString('\n')
		msg = shared.Sanitize(msg)
		//fmt.Println("MESSAGE:" + msg)
		if msg == shared.CtrlCode("disconnect") {
			os.Exit(1)
		} else if msg == shared.CtrlCode("acknowledge") {
			acknowledge(conn)
		} else if len(msg) > 0 {
			spaces := width - len(msg) - 1
			if spaces < 0 {
				spaces = 0
			}
			fmt.Println("\r\r" + msg + strings.Repeat(" ", spaces))
			showUser <- true
		} else {
			fmt.Println("\rerror 500: server has gone offline" + strings.Repeat(" ", width))
			os.Exit(1)
		}
	}
}

func write(conn net.Conn, showUser chan bool) {
	//TODO Continually get input from the user and send messages to the server.
	r := bufio.NewReader(os.Stdin)
	msg, _ := r.ReadString('\n')
	fmt.Fprintf(conn, "\000"+msg)
	//go func() {
	//	time.Sleep(5 * time.Second)
	//	acknowledge(conn)
	//	//cleanup(conn)
	//}()
	showUser <- true
}

func username(username string, showUser chan bool) {
	for {
		<-showUser
		fmt.Print(username + ": ")
	}
}

func cleanup(conn net.Conn) {
	fmt.Fprintf(conn, shared.CtrlCode("disconnect"))
	os.Exit(1)
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
	fmt.Fprintln(conn, shared.CtrlCode("username")+*pUsername)
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
