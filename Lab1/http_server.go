package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const MAX_CONNECTIONS int = 10

var acceptedExtensions = map[string]bool{
	".html": true,
	".css":  true,
	".js":   true,
	".jpg":  true,
	".png":  true,
}

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

	line, err := br.ReadString('\n')

	if err != nil {
		fmt.Printf("Error reading request line: %v\n", err)
		return
	}

	fmt.Println(line)

	// TODO: Implement check for valid http request
	// TODO: Implement 404 if requested file doesn't exist
	// TODO: Implement 501 if the method is not GET or POST
	// TODO: Write tests

	splitRequest := strings.Split(line, " ")
	method, path := splitRequest[0], splitRequest[1]

	// TODO: Implement check for accepted file types (400 Bad Request)
	extension := strings.ToLower(filepath.Ext(path))
	if extension != "" && !acceptedExtensions[extension] {
		fmt.Printf("Unsupported file type requested: %s\n", extension)
		// If the extension is not in our map, send a 400 Bad Request
		sendError(c, "400 Bad Request", fmt.Sprintf("Unsupported file type: %s", extension))
		return
	}

	fmt.Println(method)
	fmt.Println(path)

	fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len("Test"), "Test")
}

// Helper function to write an HTTP error response
func sendError(c net.Conn, status string, message string) {
	// the response includesthe HTTP version, status code, and a body.
	response := fmt.Sprintf("HTTP/1.1 %s\r\n", status)
	response += "Content-Type: text/plain\r\n"
	response += fmt.Sprintf("Content-Length: %d\r\n", len(message))
	response += "Connection: close\r\n"
	response += "\r\n"
	response += message + "\n"

	c.Write([]byte(response))
	fmt.Printf("Sent error: %s - %s\n", status, message)
}
