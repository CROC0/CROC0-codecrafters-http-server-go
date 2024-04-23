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

var statusCodes = map[int]string{
	200: "OK",
	201: "Created",
	400: "Bad Request",
	404: "Not Found",
}

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

func handleRequest(conn net.Conn) {

	req, err := parseRequest(conn)
	res := createResponse(conn)

	if err != nil {
		fmt.Println("Could not parse request: ", err.Error())
		res.status = 400
		res.sendResponse()
		return
	}

	if req.path == "/" {
		res.status = 200
		res.sendResponse()

	} else if strings.HasPrefix(req.path, "/echo/") {
		res.status = 200
		res.headers["Content-Type"] = "text/plain"
		res.body = strings.TrimPrefix(req.path, "/echo/")
		res.headers["Content-Length"] = fmt.Sprint(len(res.body))
		res.sendResponse()

	} else if req.path == "/user-agent" {
		body, ok := req.headers["User-Agent"]
		if ok {
			res.status = 200
			res.headers["Content-Type"] = "text/plain"
			res.headers["Content-Length"] = fmt.Sprint(len(body))
			res.body = body
			res.sendResponse()
		} else {
			res.status = 404
			res.sendResponse()
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
					res.status = 404
					res.sendResponse()
				} else {
					res.status = 200
					res.headers["Content-Type"] = "application/octet-stream"
					res.headers["Content-Length"] = fmt.Sprint(len(string(body)))
					res.body = string(body)
					res.sendResponse()
				}
			} else {
				if err := os.WriteFile(dir+fileName, []byte(req.body), os.ModeAppend); err != nil {
					fmt.Println("Error writing to file: ", err.Error())
					os.Exit(1)
				}
				res.status = 201
				res.sendResponse()
			}

		}
	} else {
		res.status = 404
		res.sendResponse()

	}

}

type Request struct {
	method      string
	path        string
	httpVersion string
	body        string
	headers     map[string]string
}

func parseRequest(conn net.Conn) (Request, error) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		os.Exit(1)
	}
	req := Request{}
	lines := bytes.Split(buffer, []byte("\r\n"))
	startLine := bytes.Split(lines[0], []byte(" "))
	if len(startLine) == 3 {
		req.method = string(startLine[0])
		req.path = string(startLine[1])
		req.httpVersion = string(startLine[2])
		req.headers = make(map[string]string)
	} else {
		return req, errors.New("http request not correct")
	}

	for _, v := range lines {
		line := bytes.Split(v, []byte(" "))
		if len(line) == 2 {
			name := strings.Trim(string(line[0]), ":")
			value := string(line[1])
			req.headers[name] = value
		}
	}
	str := string(buffer)
	req.body = str[strings.Index(str, "\r\n\r\n")+4 : strings.Index(str, "\x00")]
	return req, nil

}

type Response struct {
	status  int
	headers map[string]string
	body    string
	conn    net.Conn
}

func createResponse(conn net.Conn) Response {
	r := Response{}
	r.conn = conn
	r.headers = make(map[string]string)
	return r
}

func (res *Response) sendResponse() error {
	str := ""
	res.setStatus(&str)
	res.setHeaders(&str)
	str += CRLF
	res.setBody(&str)
	_, err := res.conn.Write([]byte(str))
	res.conn.Close()
	return err
}

func (res *Response) setStatus(str *string) {
	*str += fmt.Sprintf("%s %d %s%s", httpType, res.status, statusCodes[res.status], CRLF)
}

func (res *Response) setHeaders(str *string) {
	for n, v := range res.headers {
		*str += fmt.Sprintf("%s: %s%s", n, v, CRLF)
	}
}

func (res *Response) setBody(str *string) {
	*str += res.body
	*str += CRLF

}
