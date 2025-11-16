package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxConn int = 10

// map of accepted file extensions and their content types
var acceptedExtensions = map[string]string{
	".html": "text/html",
	".css":  "text/css",
	".txt":  "text/plain",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
}

func main() {

	// check if server started with required arguments
	if !(len(os.Args) > 1) {
		panic("No port provided")
	}

	// port specified by user
	port := os.Args[1]

	// create a listener
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	fmt.Println("Server started and listening on port " + port)
	// Utilize channel as semaphore
	sem := make(chan struct{}, maxConn)

	// forever loop handling each client request
	for {
		// create connection for to the listener
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		sem <- struct{}{}

		// Starts a goroutine for the connection c 
		go func(c net.Conn) {
			defer func() {
				<-sem
				fmt.Println("Done handling connection")
			}()
			fmt.Println("Handling connection")
			handleConn(c)
		}(conn)
	}
}

// Function for parsing, reading and deligating the funtionality of a HTTP request. 
func handleConn(c net.Conn) {
	defer c.Close()

	br := bufio.NewReader(c)

	req, err := parseHTTPRequest(br)	

	if err != nil {
		sendBadRequest(c)
		return
	}

	path := req.URL.Path
	if path == "" {
		sendBadRequest(c)
		return
	}

	// determine the type of request from the client
	switch req.Method {
		
	case http.MethodGet:
		handleGetRequest(c, path)

	case http.MethodPost:
		handlePostRequest(c, req)

	// invalid or missing request type
	default:
		sendNotImplemented(c)
	}
}

// Function for handling GET requests from the client.
func handleGetRequest(c net.Conn, path string) {
	if !isAcceptedExtension(path) {
		sendBadRequest(c)
		return
	}

	// read file bytes from disk
	data, err := getFileBytes(path)
	if err != nil {
		sendError(c, "404", fmt.Sprintf("404: %s not found", path))
		return
	}

	sendResponse(c, "200", data, acceptedExtensions[filepath.Ext(path)])
}

// Function for handling POST requests from the client.
func handlePostRequest(c net.Conn, req *http.Request) {
	if req.Body == nil {
		sendBadRequest(c)
		return
	}

	// create multipart reader obeject
	mr, err := req.MultipartReader() 
	if err != nil {
		sendBadRequest(c)
		return
	}

	// select the file part of the multipart reader
	part, err := mr.NextPart()
	if err != nil || part.FileName() == "" {
		sendBadRequest(c)
		return
	}

	// ex: curl -X POST -F "file=@test.txt" localhost:port -> test.txt
	fileName := filepath.Base(part.FileName())

	// get the extension of the file to upload
	extension := strings.ToLower(filepath.Ext(fileName))
	if !isAcceptedExtension(extension) {
		sendBadRequest(c)
		return
	}
	
	// create destination file in public directory
	dst, err := os.Create(filepath.Join("public", fileName))
	if err != nil {
		sendBadRequest(c)
		return
	}

	// copy stream directly. does NOT buffer entire file into memory.
	// part is like a stream reader of file data
	_, err = io.Copy(dst, part)
	dst.Close()

	if err != nil {
		sendError(c, "500", "500: Internal server error")
		return
	}
	sendOk(c)
}

// Function to check if the file extension is in acceptedExtensions[].
func isAcceptedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := acceptedExtensions[ext]
	return ok
}

// Function to send error responses to the client.
func sendError(c net.Conn, status string, message string) {
	response := fmt.Sprintf("HTTP/1.1 %s\r\n", status)
	response += "Content-Type: text/plain\r\n"
	response += fmt.Sprintf("Content-Length: %d\r\n", len(message))
	response += "\r\n"
	response += message + "\n"

	fmt.Fprint(c, response)
}

// Function to send responses to the client.
func sendResponse(c net.Conn, status string, body []byte, contentType string) {
	response := fmt.Sprintf("HTTP/1.1 %s\r\n", status)
	response += fmt.Sprintf("Content-Type: %s\r\n", contentType)
	response += fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	response += string(body)

	fmt.Fprint(c, response)
}

func sendOk(c net.Conn) {
	sendResponse(c, "200", []byte("ok"), "text/plain")
}

func sendBadRequest(c net.Conn) {
	sendError(c, "400", "400: Bad Request")
}

func sendNotImplemented(c net.Conn) {
	sendError(c, "501", "501: Not implemented")
}

// give a filename from the public directory, read it, and return its content
func getFileBytes(path string) ([]byte, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	p := filepath.Join(cwd, "public", path)

	data, err := os.ReadFile(p)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// convert a bufio.Reader object into a http.Request object
func parseHTTPRequest(br *bufio.Reader) (*http.Request, error) {
	req, err := http.ReadRequest(br)
	if err != nil {
		return nil, err
	}

	return req, nil
}
