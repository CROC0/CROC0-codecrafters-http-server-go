package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	CRLF     = "\r\n"
	httpType = "HTTP/1.1"
)

func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		buffer := make([]byte, 1028)
		_, err = connection.Read(buffer)
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}

		req := Request{}

		if err = req.parseRequest(&buffer); err != nil {
			fmt.Println("Could not parse file: ", err.Error())
			break
		}

		res := ""
		if req.path == "/" {
			setStatus(&res, 200)
			res += CRLF

		} else if strings.HasPrefix(req.path, "/echo/") {
			status := 200
			contentType := "text/plain"
			body := strings.TrimPrefix(req.path, "/echo/")
			setStatus(&res, status)
			setHeader(&res, "Content-Type", contentType)
			setHeader(&res, "Content-Length", fmt.Sprint(len(body)))
			res += CRLF
			res += body
		} else if req.path == "/user-agent" {
			body, ok := req.headers["User-Agent"]
			if ok {
				status := 200
				contentType := "text/plain"
				setStatus(&res, status)
				setHeader(&res, "Content-Type", contentType)
				setHeader(&res, "Content-Length", fmt.Sprint(len(body)))
				res += CRLF
				res += body
			} else {
				res = "HTTP/1.1 400 Bad Request\r\n\r\n"
			}
		} else {
			res = "HTTP/1.1 404 Not Found\r\n\r\n"
		}

		_, err = connection.Write([]byte(res))

		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}

		connection.Close()
	}

}

type Request struct {
	method      string
	path        string
	httpVersion string
	headers     map[string]string
}

func (req *Request) parseRequest(buffer *[]byte) error {
	lines := bytes.Split(*buffer, []byte("\r\n"))
	startLine := bytes.Split(lines[0], []byte(" "))
	if len(startLine) == 3 {
		req.method = string(startLine[0])
		req.path = string(startLine[1])
		req.httpVersion = string(startLine[2])
		req.headers = make(map[string]string)
	} else {
		return errors.New("invalid Request received")
	}

	for _, v := range lines {
		line := bytes.Split(v, []byte(" "))
		if len(line) == 2 {
			name := strings.Trim(string(line[0]), ":")
			value := string(line[1])
			req.headers[name] = value
		}
	}
	return nil
}

func setStatus(res *string, status int) {
	statusCodes := map[int]string{
		200: "OK",
		404: "Not Found",
	}
	*res += fmt.Sprintf("%s %d %s%s", httpType, status, statusCodes[status], CRLF)

}

func setHeader(res *string, name string, value string) {
	*res += fmt.Sprintf("%s: %s%s", name, value, CRLF)
}
