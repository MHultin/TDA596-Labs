package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"io"
)

const maxConn int = 10

func main() {

	if len(os.Args) == 2 {
		panic("No port provided")
	}
	proxyPort := os.Args[1]

	ln, err := net.Listen("tcp", ":"+proxyPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("Proxy listening on port " + proxyPort)
	sem := make(chan struct{}, maxConn)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		sem <- struct{}{}

		go func(c net.Conn) {
			defer func() { 
				<-sem 
			}()

			handleConn(c)
		}(conn)
	}
}

func handleConn(c net.Conn) {
	defer c.Close()

	br := bufio.NewReader(c)

	req, err := http.ReadRequest(br)

	if err != nil {
		sendBadRequest(c)
		return
	}

	if req.Method != http.MethodGet {
		sendNotImplemented(c)
		return
	}

	path := req.URL.EscapedPath()

	server, err := net.Dial("tcp", req.Host)
	if err != nil {
		sendError(c, "502", "502: Bad Gateway")
		return
	}
	defer server.Close()

	// forward a minimal GET to the origin
	bw := bufio.NewWriter(server)
	fmt.Fprintf(bw, "GET %s HTTP/1.1\r\n", path)
	// Host header: use the origin host:port
	fmt.Fprintf(bw, "Host: %s\r\n", req.Host)
	// keep it simple; close semantics
	fmt.Fprintf(bw, "Connection: close\r\n")

	// end headers
	bw.WriteString("\r\n")

	if err := bw.Flush(); err != nil {
		sendError(c, "502", "502: Bad Gateway")
		return
	}

	// stream the origin response back to the client (no parsing/modification)
	_, err = io.Copy(c, server)
	if err != nil {
		sendError(c, "500", "500 Internal server error")
	}
}

func sendError(c net.Conn, status string, message string) {
	response := fmt.Sprintf("HTTP/1.1 %s\r\n", status)
	response += "Content-Type: text/plain\r\n"
	response += fmt.Sprintf("Content-Length: %d\r\n", len(message))
	response += "\r\n"
	response += message + "\n"
	
	fmt.Fprint(c, response)
}

func sendBadRequest(c net.Conn) {
	sendError(c, "400", "400: Bad Request")
}

func sendNotImplemented(c net.Conn) {
	sendError(c, "501", "501: Not Implemented")
}