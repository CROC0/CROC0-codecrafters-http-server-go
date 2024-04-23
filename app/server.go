package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	method      string
	path        string
	httpVersion string
	body        string
	headers     map[string]string
}

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
		go handleRequest(connection)
	}

}

func handleRequest(connection net.Conn) {
	buffer := make([]byte, 1028)
	_, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}

	req := Request{}
	res := ""

	if err = req.parseRequest(&buffer); err != nil {
		fmt.Println("Could not parse request: ", err.Error())
		res = "HTTP/1.1 400 Bad Request\r\n\r\n"
		_, err = connection.Write([]byte(res))

		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
		return
	}

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
	} else if strings.HasPrefix(req.path, "/files/") {
		if len(os.Args) < 3 {
			fmt.Println("Please ensure a directory is provided. Usage: <program> --directory <directory> ")
			os.Exit(1)
		} else {
			dir := os.Args[2]
			fileName := strings.TrimPrefix(req.path, "/files/")

			if req.method == "GET" {
				body, err := os.ReadFile(dir + fileName)
				if err != nil {
					res = "HTTP/1.1 404 Not Found\r\n\r\n"
				} else {
					status := 200
					contentType := "application/octet-stream"
					setStatus(&res, status)
					setHeader(&res, "Content-Type", contentType)
					setHeader(&res, "Content-Length", fmt.Sprint(len(string(body))))
					res += CRLF
					res += string(body)
				}
			} else {
				if err := os.WriteFile(dir+fileName, []byte(req.body), os.ModeAppend); err != nil {
					fmt.Println("Error writing to file: ", err.Error())
					os.Exit(1)
				}
				res = "HTTP/1.1 201 Created\r\n\r\n"
			}

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
	str := string(*buffer)
	req.body = str[strings.Index(str, "\r\n\r\n")+4 : strings.Index(str, "\x00")]
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
