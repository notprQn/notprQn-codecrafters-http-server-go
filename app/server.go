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

const (
	GET = "GET"

	POST = "POST"

	PUT = "PUT"

	DELETE = "DELETE"
)

type RequestParams struct {
	method string

	path string

	version string

	sender string

	headers map[string]string
}

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

		handleConnection(conn)

		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {

	request := make([]byte, 1024)

	_, err := conn.Read(request)

	if err != nil {

		return

	}

	reqParams := getReqParams(request)

	switch reqParams.method {

	case GET:

		_ = handleGetRequest(reqParams, conn)

		return

	default:

		return

	}

}

func handleGetRequest(reqParams RequestParams, conn net.Conn) error {

	switch reqParams.path {

	case "/":

		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

		return nil

	default:

		reqPathAndValue := strings.Split(reqParams.path, "/")

		switch reqPathAndValue[1] {

		case "echo":

			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",

				len(reqPathAndValue[2]), reqPathAndValue[2])))

			return nil

		case "user-agent":

			reqHeader := reqParams.headers["User-Agent"]

			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",

				len(reqHeader), reqHeader)))

			return nil

		default:

			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

			return nil

		}

	}

}

func getReqParams(request []byte) RequestParams {

	requestString := string(request)

	fields := strings.Split(strings.Split(requestString, "\r\n\r\n")[0], "\r\n")

	// extract request type and path and http version

	reqDetails := strings.Split(fields[0], " ")

	hostDetails := strings.Split(fields[1], " ")

	reqHeaders := getHeaders(fields)

	return RequestParams{

		method: reqDetails[0],

		path: reqDetails[1],

		version: reqDetails[2],

		sender: strings.TrimSpace(hostDetails[1]),

		headers: reqHeaders,
	}

}

func getHeaders(reqDetails []string) map[string]string {

	headers := make(map[string]string)

	for index, elem := range reqDetails {

		if index == 0 || index == 1 {

			continue

		} else if elem == "" || elem == " " {

			break

		}

		temp := strings.Split(elem, ":")

		headerName := strings.TrimSpace(temp[0])

		headerValue := strings.TrimSpace(temp[1])

		headers[headerName] = headerValue

	}

	return headers

}
