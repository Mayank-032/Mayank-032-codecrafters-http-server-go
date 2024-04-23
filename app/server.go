package main

import (
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

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

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
	
	switch path {
	case "/":
		response = "HTTP/1.1 200 OK\r\n\r\n"
	case "/index.html":
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	default:
		randomString := processPathToFetchRandomString(path)
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(randomString), randomString)
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

func processPathToFetchRandomString(path string) string {
	pathArr := strings.Split(path, "/")
	if len(pathArr) > 1 {
		return pathArr[2]
	}
	return ""
}