package main

import (
	"bytes"
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

		header := bytes.Split(buffer, []byte("\r\n"))
		startLine := bytes.Split(header[0], []byte(" "))
		// _method := startLine[0]
		path := startLine[1]
		// _httpVersion := startLine[2]
		res := ""

		if string(path) == "/" {
			setStatus(&res, 200)
			res += CRLF
			_, err = connection.Write([]byte(res))
		} else if strings.HasPrefix(string(path), "/echo/") {

			status := 200
			contentType := "text/plain"
			body := strings.TrimPrefix(string(path), "/echo/")
			setStatus(&res, status)
			setHeader(&res, "Content-Type", contentType)
			setHeader(&res, "Content-Length", fmt.Sprint(len(body)))
			res += CRLF
			res += body
			_, err = connection.Write([]byte(res))
		} else {
			_, err = connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

		}

		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}

		connection.Close()
	}

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
