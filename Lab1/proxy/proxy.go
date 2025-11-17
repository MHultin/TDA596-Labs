package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

// maximum concurrent connections
const maxConn int = 10

func main() {
	if len(os.Args) <= 1 {
		panic("No port provided")
	}
	proxyPort := os.Args[1]

	// start listening for incoming connections
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

// processes the client connection
func handleConn(c net.Conn) {
	defer c.Close()

	br := bufio.NewReader(c)

	req, err := http.ReadRequest(br)

	if err != nil {
		sendBadRequest(c)
		return
	}

	// only support GET method
	if req.Method != http.MethodGet {
		sendNotImplemented(c)
		return
	}

	// append default port if not specified
	if _, _, err := net.SplitHostPort(req.Host); err != nil {
		req.Host = net.JoinHostPort(req.Host, "80")
	}

	// connect to the target server
	server, err := net.Dial("tcp", req.Host)
	if err != nil {
		sendError(c, "502", "502: Bad Gateway")
		return
	}

	defer server.Close()

	// send minimal GET request to the server
	err = sendMinimalGET(server, req)
	if err != nil {
		sendError(c, "502", "502: Bad Gateway")
	}

	_, err = io.Copy(c, server)
	if err != nil {
		sendError(c, "500", "500 Internal server error")
	}
}

// sends a minimal GET request to the target server
func sendMinimalGET(c net.Conn, req *http.Request) error {
	request := fmt.Sprintf("GET %s HTTP/1.1\r\n", req.RequestURI)
	request += fmt.Sprintf("Host: %s\r\n", req.Host)
	request += "Connection: close\r\n\r\n"

	_, err := fmt.Fprintf(c, "%s", request)

	return err
}

// sends an error response to the client
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
