package main

import (
	"fmt"

	"strings"

	//Uncomment this block to pass the first stage

	"net"

	"os"
)

func main() {

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {

		fmt.Println("Failed to bind to port 4221")

		os.Exit(1)

	}

	res, err := l.Accept()

	if err != nil {

		fmt.Println("Error accepting connection: ", err.Error())

		os.Exit(1)

	}

	requestBuffer := make([]byte, 4096)

	_, err = res.Read(requestBuffer)

	if err != nil {

		return

	}

	request := string(requestBuffer)

	requestPath := strings.Split(request, " ")[1]

	response := ""

	if requestPath == "/" {

		response = "HTTP/1.1 200 OK\r\n\r\n"

	} else if strings.HasPrefix(requestPath, "/echo/") {

		echoStr := strings.TrimPrefix(requestPath, "/echo/")

		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echoStr), echoStr)

	} else {

		response = "HTTP/1.1 404 Not Found\r\n\r\n"

	}

	_, err = res.Write([]byte(response))

	if err != nil {

		return

	}

}
