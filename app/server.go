package main

import (
	"fmt"

	"net"

	"os"

	"strconv"

	"strings"
)

func main() {

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {

		fmt.Println("Failed to bind to port 4221")

		os.Exit(1)

	}

	for {

		conn, err := l.Accept()

		if err != nil {

			fmt.Println("Error accepting connection: ", err.Error())

			os.Exit(1)

		}

		go func(c net.Conn) {

			defer c.Close()

			// Read data from connection

			data := make([]byte, 1024)

			n, err := c.Read(data)

			if err != nil {

				fmt.Println("Error reading data:", err.Error())

				return

			}

			// Extract potential request path

			request := string(data[:n])

			lines := strings.Split(request, "\r\n")

			path := getRequestPath(lines[0])

			if path == "/" {

				conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

				return

			}

			if strings.HasPrefix(path, "/echo/") {

				text := strings.TrimPrefix(path, "/echo/")

				conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", len(text), text)))

				return

			}

			if path == "/user-agent" {

				fmt.Println(lines[2])

				resBody := strings.TrimPrefix(lines[2], "User-Agent: ")

				conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(resBody), resBody)))

				return

			}

			if strings.HasPrefix(path, "/files/") {

				directory := os.Args[2]

				fileName := strings.TrimPrefix(path, "/files/")

				data, err := os.ReadFile(directory + fileName)

				if err != nil {

					conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

				} else {

					conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(len(data)) + "\r\n\r\n" + string(data) + "\r\n\r\n"))

				}

			}

			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

		}(conn)

	}

}

func getRequestPath(line string) string {

	path := strings.Split(line, " ")[1]

	return path

}
