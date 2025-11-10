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

// TODO: Implement check for valid http request
// TODO: Write tests

const maxConn int = 10

var acceptedExtensions = map[string]string{
	".html": "text/html",
	".css":  "text/css",
	".txt":  "text/plain",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
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

	switch req.Method {
		
	case http.MethodGet:
		handleGetRequest(c, path)
		
	case http.MethodPost:
		handlePostRequest(c, req)
		
	default:
		sendNotImplemented(c)
	}
}

func handleGetRequest(c net.Conn, path string) {
	if !isAcceptedExtension(path) {
		sendBadRequest(c)
		return
	}
	data, err := getFileBytes(path)
	if err != nil {
		sendError(c, "404", fmt.Sprintf("404: %s not found", path))
		return
	}

	sendResponse(c, "200", data, acceptedExtensions[filepath.Ext(path)])
}

func handlePostRequest(c net.Conn, req *http.Request) {
	if req.Body == nil {
		sendBadRequest(c)
		return
	}

	// convert the HTTP req into a multipart reader
	mr, err := req.MultipartReader() // magic
	if err != nil {
		sendBadRequest(c)
		return
	}

	// only one file expected here. get first part
	part, err := mr.NextPart() // magic
	if err != nil || part.FileName() == "" {
		sendBadRequest(c)
		return
	}

	// ex: curl -F file=@test.txt  -> test.txt
	fileName := filepath.Base(part.FileName())

	extension := strings.ToLower(filepath.Ext(fileName))
	if !isAcceptedExtension(extension) {
		sendBadRequest(c)
		return
	}

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
		sendError(c, "500", "500: Internal server error") // change?
		return
	}
	sendOk(c)
}

func isAcceptedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := acceptedExtensions[ext]
	return ok
}

func sendError(c net.Conn, status string, message string) {
	response := fmt.Sprintf("HTTP/1.1 %s\r\n", status)
	response += "Content-Type: text/plain\r\n"
	response += fmt.Sprintf("Content-Length: %d\r\n", len(message))
	response += "\r\n"
	response += message + "\n"

	fmt.Fprint(c, response)
}

func sendResponse(c net.Conn, status string, body []byte, contentType string) {
	response := fmt.Sprintf("HTTP/1.1 %s\r\n", status)
	response += fmt.Sprintf("Content-Type: %s\r\n", contentType)
	response += fmt.Sprintf("Content-Length: %d\r\n", len(body))
	response += "\r\n"
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

func parseHTTPRequest(br *bufio.Reader) (*http.Request, error) {
	req, err := http.ReadRequest(br)
	if err != nil {
		return nil, err
	}

	return req, nil
}
