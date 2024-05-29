package main

import (
	"fmt"

	"net"

	"os"

	"strings"
	// Uncomment this block to pass the first stage
	// "net"
	// "os"
)

type Request struct {
	Method string

	Headers map[string]string

	Path string

	Body string
}

func ParseRequest(rawReq string) *Request {

	req := &Request{Headers: make(map[string]string)}

	parts := strings.Split(rawReq, "\r\n")

	reqLine := strings.Fields(parts[0])

	req.Method = reqLine[0]

	req.Path = reqLine[1]

	part := 1

	for parts[part] != "" {

		header := strings.Split(parts[part], ": ")

		req.Headers[header[0]] = header[1]

		part++

	}

	part++

	req.Body = parts[part]

	return req

}

type Server struct {
	Addr string

	Port string

	Fs string
}

func main() {

	// You can use print statements as follows for debugging, they'll be visible when running tests.

	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	//

	s := &Server{

		Addr: "0.0.0.0",

		Port: "4221",
	}

	if len(os.Args) > 2 {

		if os.Args[1] == "--directory" {

			s.Fs = os.Args[2]

			err := os.Chdir(s.Fs)

			if err != nil {

				fmt.Fprintf(os.Stderr, "unable to serve FS %s error: %s", s.Fs, err)

				os.Exit(1)

			}

		}

	}

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.Addr, s.Port))

	if err != nil {

		fmt.Fprintf(os.Stderr, "Failed to bind to port %s", s.Port)

		os.Exit(1)

	}

	defer l.Close()

	for {

		conn, err := l.Accept()

		if err != nil {

			fmt.Fprintln(os.Stderr, "Error accepting connection: ", err.Error())

			os.Exit(1)

		}

		go func(conn net.Conn) {

			defer conn.Close()

			buf := make([]byte, 1024)

			n, err := conn.Read(buf)

			if err != nil {

				fmt.Fprintln(os.Stderr, "listening error occurred: ", err)

				os.Exit(1)

			}

			req := ParseRequest(string(buf[:n]))

			if req.Path == "/" {

				conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

			}

			if strings.HasPrefix(req.Path, "/echo/") {

				customHeaders := ""

				message := req.Path[len("/echo/"):]

				if v, ok := req.Headers["Accept-Encoding"]; ok {

					if v == "gzip" {

						customHeaders += "Content-Encoding: gzip\r\n"

					}

				}

				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n%s\r\n%s", len(message), customHeaders, message)

			}

			if strings.HasPrefix(req.Path, "/user-agent") {

				message := req.Headers["User-Agent"]

				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)

			}

			if strings.HasPrefix(req.Path, "/files/") && req.Method == "GET" {

				file := req.Path[len("/files/"):]

				data, err := os.ReadFile(file)

				if err != nil {

					conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

					return

				}

				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)

			}

			if strings.HasPrefix(req.Path, "/files/") && req.Method == "POST" {

				file := req.Path[len("/files/"):]

				err := os.WriteFile(file, []byte(req.Body), 0666)

				if err != nil {

					conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))

					return

				}

				fmt.Fprintf(conn, "HTTP/1.1 201 Created\r\n\r\n")

			} else {

				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

			}

		}(conn)

	}

}
