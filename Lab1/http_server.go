package lab1

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
			}()

			handleConn(c)
		}(conn)
	}
}

var acceptedExtensions = map[string]string{
	".html": "text/html",
	".css":  "text/css",
	".txt":  "text/plain",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
}

func handleConn(c net.Conn) {
	defer c.Close()

	br := bufio.NewReader(c)

	req, err := parseHTTPRequest(br)

	if err != nil {
		sendBadRequest(c)
		return
	}

	if !isAcceptedMethod(req.Method) {
		sendNotImplemented(c)
		return
	}
	if !isAcceptedExtension(req.RequestURI) {
		sendBadRequest(c)
		return
	}

	switch req.Method {
	case "GET":
		handleGetRequest(c, req.RequestURI)
	case "POST":
		handlePostRequest(c, req)
	}
}

func handlePostRequest(c net.Conn, req *http.Request) {
	if req.Body == nil {
		sendBadRequest(c)
		return
	}

	splitUrl := strings.Split(req.RequestURI, "/")

	if len(splitUrl) != 2 || splitUrl[1] == "" {
		sendBadRequest(c)
		return
	}

	fileName := splitUrl[1]

	data, err := io.ReadAll((req.Body))

	if err != nil {
		sendError(c, "500", "500: Internal server error")
		return
	}

	os.Chdir("public")

	err = os.WriteFile(fileName, data, 0644)

	if err != nil {
		sendError(c, "500", "500: Internal server error")
		return
	}

	os.Chdir("..")

	sendOk(c)
}

func handleGetRequest(c net.Conn, path string) {
	f, err := getFileBytes(path)
	if err != nil {
		sendError(c, "404", fmt.Sprintf("404: %s not found", path))
	}

	sendResponse(c, "200", f, acceptedExtensions[filepath.Ext(path)])
}

func isAcceptedExtension(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	return !(extension == "" || acceptedExtensions[extension] == "")
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

	f, err := os.ReadFile(p)

	if err != nil {
		return nil, err
	}

	return f, nil
}

func isAcceptedMethod(method string) bool {
	return method == "GET" || method == "POST"
}

func parseHTTPRequest(br *bufio.Reader) (*http.Request, error) {
	req, err := http.ReadRequest(br)
	if err != nil {
		return nil, err
	}

	return req, nil
}
