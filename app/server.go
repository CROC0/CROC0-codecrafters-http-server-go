package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

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

	if string(path) == "/" {
		_, err = connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		_, err = connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

	}

	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}

}
