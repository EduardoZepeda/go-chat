package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	port = flag.Int("p", 3090, "port")
	host = flag.String("h", "localhost", "host")
)

func main() {
	flag.Parse()

	// Connecting to chat host:port
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("connected")
	done := make(chan struct{})

	// Copy the data received from conn to Stdout
	go func() {
		io.Copy(os.Stdout, conn)
		done <- struct{}{}
	}()

	// Copy what is in Stdin to conn
	CopyContent(conn, os.Stdin)
	conn.Close()
	<-done

}

// Copy content from src to dst
func CopyContent(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		fmt.Fprintf(os.Stderr, "io.Copy: %v\n", err)
		os.Exit(1)
	}
}
