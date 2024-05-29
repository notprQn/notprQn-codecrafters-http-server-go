package main

import (
	"bytes"

	"compress/gzip"

	"fmt"

	"net"

	"os"

	"path"

	"strings"
)

type (
	Request struct {
		httpversion string

		method string

		path string

		headers map[string]string

		body []byte
	}

	ResponseWriter struct {
		conn net.Conn
	}
)

func main() {

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

		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {

	writter := ResponseWriter{conn: conn}

	buf := make([]byte, 4096)

	_, err := conn.Read(buf)

	if err != nil {

		return

	}

	defer conn.Close()

	req, ok := parseRequest(string(buf))

	if !ok {

		writter.write(400, nil, nil)

	}

	if req.path == "/" {

		writter.write(200, nil, nil)

	} else if req.path == "/user-agent" {

		writter.write(

			200,

			map[string]string{

				"Content-Type": "text/plain",

				"Content-Length": fmt.Sprintf("%d", len(req.headers[strings.ToUpper("User-Agent")])),
			},

			[]byte(req.headers[strings.ToUpper("User-Agent")]),
		)

	} else if strings.HasPrefix(req.path, "/echo/") {

		param, _ := strings.CutPrefix(req.path, "/echo/")

		respbody := []byte(param)

		respstatus := 200

		respheaders := make(map[string]string)

		respheaders["Content-Type"] = "text/plain"

		respheaders["Content-Length"] = fmt.Sprintf("%d", len(param))

		if contentEncodings, found := req.headers[strings.ToUpper("Accept-Encoding")]; found {

			if strings.Contains(contentEncodings, "gzip") {

				respheaders["Content-Encoding"] = "gzip"

				buffer := new(bytes.Buffer)

				writter := gzip.NewWriter(buffer)

				writter.Write(respbody)

				writter.Close()

				respbody = buffer.Bytes()

			}

		}

		respheaders["Content-Type"] = "text/plain"

		respheaders["Content-Length"] = fmt.Sprintf("%d", len(respbody))

		writter.write(

			respstatus,

			respheaders,

			respbody,
		)

	} else if req.method == "GET" && strings.HasPrefix(req.path, "/files/") {

		dir := os.Args[2]

		filename := strings.TrimPrefix(req.path, "/files/")

		data, err := os.ReadFile(path.Join(dir, filename))

		if err != nil {

			writter.write(404, nil, nil)

		} else {

			writter.write(

				200,

				map[string]string{

					"Content-Type": "application/octet-stream",

					"Content-Length": fmt.Sprintf("%d", len(data)),
				},

				[]byte(data),
			)

		}

	} else if req.method == "POST" && strings.HasPrefix(req.path, "/files/") {

		dir := os.Args[2]

		filename := strings.TrimPrefix(req.path, "/files/")

		data := bytes.Trim(req.body, "\x00") // trim excess bytes

		err := os.WriteFile(path.Join(dir, filename), data, 0644)

		if err != nil {

			writter.write(404, nil, nil)

		} else {

			writter.write(201, nil, nil)

		}

	} else {

		writter.write(404, nil, nil)

	}

}

func parseRequest(s string) (req *Request, ok bool) {

	method, after1, ok1 := strings.Cut(s, " ")

	path, after2, ok2 := strings.Cut(after1, " ")

	httpversion, rawHeadersAndBody, ok3 := strings.Cut(after2, "\r\n")

	if !ok1 || !ok2 || !ok3 {

		return nil, false

	}

	split := strings.Split(rawHeadersAndBody, "\r\n\r\n")

	rawheaders := split[0]

	rawbody := split[1]

	req = &Request{}

	req.httpversion = httpversion

	req.method = method

	req.path = path

	req.headers = mapheaders(strings.Split(rawheaders, "\r\n"))

	req.body = []byte(rawbody)

	return req, true

}

func mapheaders(ss []string) map[string]string {

	headers := make(map[string]string)

	for _, s := range ss {

		split := strings.Split(s, ": ")

		headers[strings.ToUpper(split[0])] = split[1]

	}

	return headers

}

func httpstatus(code int) string {

	switch code {

	case 200:

		return "OK"

	case 201:

		return "Created"

	case 400:

		return "Bad Request"

	case 404:

		return "Not Found"

	default:

		return "I'm a teapot"

	}

}

func (w ResponseWriter) write(status int, headers map[string]string, body []byte) error {

	h := strings.Builder{}

	if headers != nil && 0 < len(headers) {

		for k, v := range headers {

			_, err := h.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))

			if err != nil {

				panic(err)

			}

		}

	}

	statusline := fmt.Sprintf("HTTP/1.1 %d %s", status, httpstatus(status))

	resp := statusline + "\r\n" + h.String() + "\r\n" + string(body)

	_, err := w.conn.Write([]byte(resp))

	if err != nil {

		fmt.Printf("%v", err)

		return err

	}

	return nil

}
