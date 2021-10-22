package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
)

type Client chan<- string

var (
	incomingClients = make(chan Client)
	leavingClients  = make(chan Client)
	messages        = make(chan string)
)

var (
	host = flag.String("h", "localhost", "host")
	port = flag.Int("p", 3090, "port")
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	message := make(chan string)
	go MessageWrite(conn, message)
	clientName := conn.RemoteAddr().String()
	message <- fmt.Sprintf("Welcome to the server, your name: %s\n", clientName)
	// Everybody will receive a message announcing the new user connection
	messages <- fmt.Sprintf("%s connected to the chat\n", clientName)
	incomingClients <- message
	inputMessage := bufio.NewScanner(conn)
	for inputMessage.Scan() {
		messages <- fmt.Sprintf("%s: %s", clientName, inputMessage.Text())
	}
	// If something interrupts the message, then pass the message to leavingClients channel
	leavingClients <- message
	messages <- fmt.Sprintf("%s left the chat", clientName)
}

func MessageWrite(conn net.Conn, messages <-chan string) {
	// Iterate over every message and print them
	for message := range messages {
		fmt.Fprintln(conn, message)
	}
}

func Broadcast() {
	// track every client connected
	clients := make(map[Client]bool)
	for {
		select {
		// broadcast every message in messages to every client
		case message := <-messages:
			for client := range clients {
				client <- message
			}
		case newClient := <-incomingClients:
			clients[newClient] = true
		case leavingClient := <-leavingClients:
			delete(clients, leavingClient)
			close(leavingClient)
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatal(err)
	}
	go Broadcast()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			// continue is needed instead of return,
			//otherwise the chat server will finish its execution
			continue
		}
		go HandleConnection(conn)
	}
}
