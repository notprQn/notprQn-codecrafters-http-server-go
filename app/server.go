package main

import (
	"fmt"

	"net"

	"os"

	"strings"
)

func main() {

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

	data := make([]byte, 4096)

	i, err := conn.Read(data)

	fmt.Println("data data", i)

	if err != nil {

		fmt.Println("Error reading data: ", err.Error())

		os.Exit(1)

	}

	request_parts := strings.Split(string(data), "\r\n")

	path := strings.Split(request_parts[0], " ")[1]

	var response string

	if path == "/" {

		response = "HTTP/1.1 200 OK\r\n\r\n"

	} else if path[0:6] == "/echo/" {

		echo := path[6:]

		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)

	} else {

		response = "HTTP/1.1 404 NOT FOUND \r\n\r\n"

	}

	_, err = conn.Write([]byte(response))

	if err != nil {

		fmt.Println("Error writing data: ", err.Error())

		os.Exit(1)

	}

}
