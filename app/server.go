package main

import (
	"flag"

	"fmt"

	"io"

	"net"

	"os"

	"strings"
)

var directory string

func main() {

	flag.StringVar(&directory, "directory", ".", "the directory to serve files from")

	flag.Parse()

	if directory != "" {

		if _, err := os.Stat(directory); os.IsNotExist(err) {

			panic(err)

		}

	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {

		panic(err)

	}

	for {

		con, err := l.Accept()

		if err != nil {

			panic(err)

		}

		go handle(con)

	}

}

func handle(c net.Conn) {

	defer c.Close()

	first, err := readUntil(c, []byte("\r\n"))

	if err != nil {

		panic(err)

	}

	ua := ""

	cl := ""

	ae := ""

	cln := 0

	for {

		line, err := readUntil(c, []byte("\r\n"))

		if err != nil {

			panic(err)

		}

		if len(line) == 0 {

			break

		}

		if strings.HasPrefix(string(line), "User-Agent: ") {

			ua = string(line)[len("User-Agent: "):]

		}

		if strings.HasPrefix(string(line), "Content-Length: ") {

			cl = string(line)[len("Content-Length: "):]

		}

		if strings.HasPrefix(string(line), "Accept-Encoding: ") {

			ae = string(line)[len("Accept-Encoding: "):]

		}

	}

	if cl != "" {

		fmt.Sscanf(cl, "%d", &cln)

	}

	body := make([]byte, cln)

	if cln > 0 {

		_, err = io.ReadFull(c, body)

		if err != nil {

			panic(err)

		}

	}

	parts := strings.Split(string(first), " ")

	method := parts[0]

	path := parts[1]

	if path == "/" {

		c.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	} else if strings.HasPrefix(path, "/echo/") && len(path) > len("/echo/") {

		echo := path[len("/echo/"):]

		if ae == "gzip" || strings.HasPrefix(ae, "gzip, ") || strings.HasSuffix(ae, ", gzip") || strings.Contains(ae, ", gzip,") {

			fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nContent-Encoding: gzip\r\n\r\n%s", len(echo), echo)

		} else {

			fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)

		}

	} else if string(path) == "/user-agent" {

		fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(ua), ua)

	} else if strings.HasPrefix(path, "/files/") && len(path) > len("/files/") {

		if method == "GET" {

			file := path[len("/files/"):]

			f, err := os.Open(directory + "/" + file)

			if err != nil {

				c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

				return

			}

			fi, err := f.Stat()

			if err != nil {

				c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

				return

			}

			fmt.Fprintf(c, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", fi.Size())

			io.Copy(c, f)

		} else if method == "POST" {

			file := path[len("/files/"):]

			f, err := os.Create(directory + "/" + file)

			if err != nil {

				c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))

				return

			}

			defer f.Close()

			_, err = f.Write(body)

			if err != nil {

				c.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))

				return

			}

			c.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))

		}

	} else {

		c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

	}

}

func readUntil(r io.Reader, delim []byte) ([]byte, error) {

	var buf []byte

	for {

		b := make([]byte, 1)

		_, err := r.Read(b)

		if err != nil {

			return nil, err

		}

		buf = append(buf, b...)

		if strings.HasSuffix(string(buf), string(delim)) {

			break

		}

	}

	// return buf, nil

	bufMinusDelim := buf[:len(buf)-len(delim)]

	return bufMinusDelim, nil

}
