package main

import (
	"errors"
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
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

		go readFromConn(conn)
	}
}

func readFromConn(conn net.Conn) {
	var reqByte = make([]byte, 4096)
	reqByteSize, err := conn.Read(reqByte)
	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		os.Exit(1)
	}
	fmt.Printf("req_bytes_size: %v, req_bytes: %v\n", reqByteSize, string(reqByte))

	var response string

	path := extractPath(reqByte)
	fmt.Println("path: ", path)

	header := extractHeader(reqByte)
	fmt.Println("header: ", header)

	switch path {
	case "/":
		response = "HTTP/1.1 200 OK\r\n\r\n"
	case "/index.html":
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	case "/user-agent":
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(header), header)
	default:
		randomString, err := processPathToFetchRandomString(path)
		if err != nil && err.Error() == "invalid path" {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(randomString), randomString)
		}
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing on connection: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("successfully written on connection: ", response)
}

func extractPath(reqByte []byte) string {
	reqBodySplitArr := strings.Split(string(reqByte), " ")
	var path string
	if len(reqBodySplitArr) > 1 {
		path = reqBodySplitArr[1]
	}
	return path
}

func processPathToFetchRandomString(path string) (string, error) {
	pathArr := strings.Split(path, "/")
	if pathArr[1] == "echo" {
		if len(pathArr) > 2 {
			ind := 2
			randomString := ""
			for ind < len(pathArr) {
				randomString += pathArr[ind] + "/"
				ind++
			}
			randomString = randomString[:len(randomString)-1]
			return randomString, nil
		}
		return "", nil
	}
	return "", errors.New("invalid path")
}

func extractHeader(reqByte []byte) string {
	reqBodySplitArr := strings.Split(string(reqByte), "\r\n")

	if len(reqBodySplitArr) > 2 {
		headerKeyValStr := reqBodySplitArr[2]

		headerArr := strings.Split(headerKeyValStr, ": ")
		if len(headerArr) > 1 {
			return headerArr[1]
		}
		return ""
	}

	return ""
}
