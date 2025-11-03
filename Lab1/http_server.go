package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const MAX_CONNECTIONS int = 10

func main() {

	if len(os.Args) == 1 {
		panic("No port provided")
	}

	port := os.Args[1]

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	fmt.Println("Server started and listening on port " + port)
	sem := make(chan struct{}, MAX_CONNECTIONS)

	for {
		sem <- struct{}{}

		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go func(c net.Conn) {
			defer func() {
				<-sem
				fmt.Println("Connection closed")
			}()

			handleConn(c)
		}(conn)
	}
}

func handleConn(c net.Conn) {
	defer c.Close()

	br := bufio.NewReader(c)

	line, _ := br.ReadString('\n')
	fmt.Println(line)

	// TODO: Implement check for accepted file types (400 Bad Request)
	// TODO: Implement check for valid http request
	// TODO: Implement 404 if requested file doesn't exist
	// TODO: Implement 501 if the method is not GET or POST
	// TODO: Write tests

	splitRequest := strings.Split(line, " ")
	method, path := splitRequest[0], splitRequest[1]

	fmt.Println(method)
	fmt.Println(path)

	fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len("Test"), "Test")
}
