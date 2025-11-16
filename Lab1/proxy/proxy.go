package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

const maxConn int = 10

func main() {
	if len(os.Args) == 1 {
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

	server, err := net.Dial("tcp", req.Host)
	if err != nil {
		sendError(c, "502", "502: Bad Gateway")
		return
	}

	defer server.Close()

	err = sendMinimalGET(server, req)
	if err != nil {
		sendError(c, "502", "502: Bad Gateway")
	}

	_, err = io.Copy(c, server)
	if err != nil {
		sendError(c, "500", "500 Internal server error")
	}
}

func sendMinimalGET(c net.Conn, req *http.Request) error {
	request := fmt.Sprintf("GET %s HTTP/1.1\r\n", req.RequestURI)
	request += fmt.Sprintf("Host: %s\r\n", req.Host)
	request += "Connection: close\r\n\r\n"

	_, err := fmt.Fprintf(c, request)

	return err
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
