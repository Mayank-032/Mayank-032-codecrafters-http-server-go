package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

const (
	OK_RESPONSE  = "HTTP/1.1 200 OK\r\n\r\n"
	ERR_RESPONSE = "HTTP/1.1 404 Not Found\r\n\r\n"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	var directoryPath string
	flag.StringVar(&directoryPath, "directory", "", "Path to the directory")
	flag.Parse()

	fmt.Println("directory path: ", directoryPath)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go readFromConn(conn, directoryPath)
	}
}

func readFromConn(conn net.Conn, directoryPath string) {
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
		response = OK_RESPONSE
	case "/index.html":
		response = ERR_RESPONSE
	case "/user-agent":
		response = OK_RESPONSE + fmt.Sprintf("Content-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(header), header)
	default:
		remainingPathString, commandType, err := processPathToFetchString(path)
		if err != nil && err.Error() == "invalid path" {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			fmt.Println("rem_path_str: ", remainingPathString)
			fmt.Println("command_type: ", commandType)
			switch commandType {
			case "echo":
				response = OK_RESPONSE + fmt.Sprintf("Content-Type: text/plain\r\nContent-Length: %v\r\n\r\n%v", len(remainingPathString), remainingPathString)
			case "files":
				files, err := os.ReadDir(directoryPath)
				if err != nil {
					fmt.Println("unable to read directory: ", err.Error())
					os.Exit(1)
				}

				fmt.Println("here1, len_of_files: ", len(files))

				var flag = false
				for _, file := range files {
					if file.Name() == remainingPathString {
						flag = true
						break
					}
				}

				if !flag {
					response = ERR_RESPONSE
				} else {
					if string(directoryPath[len(directoryPath) - 1]) != "/" {
						remainingPathString = "/"+remainingPathString
					}
					filePath := directoryPath + remainingPathString
					file, err := os.Open(filePath)
					if err != nil {
						fmt.Println("unable to open file: ", err.Error())
						os.Exit(1)
					}
					defer file.Close()

					var content = make([]byte, 4096)
					fileContentSize, err := file.Read(content)
					if err != nil {
						fmt.Println("unable to read file content: " + err.Error())
						os.Exit(1)
					}

					response = OK_RESPONSE + fmt.Sprintf("Content-Type: application/octet-stream\r\nContent-Length: %v\r\n\r\n%v", fileContentSize, string(content))
				}
			default:
				fmt.Println("Invalid command type")
				os.Exit(1)
			}
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

func processPathToFetchString(path string) (string, string, error) {
	pathArr := strings.Split(path, "/")

	switch pathArr[1] {
	case "echo", "files":
		if len(pathArr) > 2 {
			ind := 2
			randomString := ""
			for ind < len(pathArr) {
				randomString += pathArr[ind] + "/"
				ind++
			}
			randomString = randomString[:len(randomString)-1]

			return randomString, pathArr[1], nil
		}
		return "", pathArr[1], errors.New("invalid path")
	}

	return "", pathArr[1], errors.New("invalid path")
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
