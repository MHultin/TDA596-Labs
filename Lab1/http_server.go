package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn)

	}
}

func handleConn(c net.Conn) {
	defer c.Close()

	br := bufio.NewReader(c)

	line, _ := br.ReadString('\n') // "GET / HTTP/1.1\r\n"
	fmt.Println(line)

	splitRequest := strings.Split(line, " ")
	method, path := splitRequest[0], splitRequest[1]

	fmt.Println(method)
	fmt.Println(path)

	fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len("Test"), "Test")
}
