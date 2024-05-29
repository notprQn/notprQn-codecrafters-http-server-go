package main

import (
	"fmt"

	"net"

	"os"

	"strconv"

	"strings"
)

const (
	OK = "HTTP/1.1 200 OK\r\n\r\n"

	NOT_FOUND = "HTTP/1.1 404 Not Found\r\n\r\n"

	TCP_HOST = "0.0.0.0"

	TCP_PORT = "4221"
)

type httpRequest struct {
	method string

	Headers map[string]string

	path string

	Body string
}

func (req *httpRequest) parseRequest(request string) *httpRequest {

	requestLines := strings.Split(request, "\r\n")

	req.method = strings.Split(requestLines[0], " ")[0]

	req.Headers = make(map[string]string)

	req.path = strings.Split(requestLines[0], " ")[1]

	for i := 1; i < len(requestLines); i++ {

		if requestLines[i] == "" {

			req.Body = requestLines[i+1]

			break

		}

		header := strings.Split(requestLines[i], ": ")

		req.Headers[header[0]] = header[1]

	}

	return req

}

func (req *httpRequest) parsePath() string {

	return strings.TrimPrefix(req.path, "/echo/")

}

func (req *httpRequest) generateReponse(status string, body string, headers ...string) string {

	var sb strings.Builder

	sb.WriteString("HTTP/1.1 ")

	sb.WriteString(status)

	for _, header := range headers {

		sb.WriteString(header)

		sb.WriteString("\r\n")

	}

	sb.WriteString("\r\n")

	sb.WriteString(body)

	return sb.String()

}

func main() {

	// You can use print statements as follows for debugging, they'll be visible when running tests.

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {

		fmt.Println("Failed to bind to port 4221")

		os.Exit(1)

	}

	conn, err := l.Accept()

	if err != nil {

		fmt.Println("Error accepting connection: ", err.Error())

		os.Exit(1)

	}

	var buf []byte = make([]byte, 1024)

	conn.Read(buf)

	req := new(httpRequest)

	req.parseRequest(string(buf))

	if req.path == "/" {

		conn.Write([]byte(OK))

	} else if strings.HasPrefix(req.path, "/echo/") {

		msg := req.parsePath()

		conn.Write([]byte(req.generateReponse("200 OK\r\n", req.parsePath(), "Content-type: text/plain", "Content-Length: "+strconv.Itoa(len(msg)))))

	} else if req.path == "/user-agent" {

		conn.Write([]byte(req.generateReponse("200 OK\r\n", req.Headers["User-Agent"], "Content-type: text/plain", "Content-Length: "+strconv.Itoa(len(req.Headers["User-Agent"])))))

	} else {

		conn.Write([]byte(NOT_FOUND))

	}

}
