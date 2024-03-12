package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	HandleConnection(conn)
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	request := ReadHttpRequest(conn)
	fmt.Printf("Request: %s", request)
	path, msg := ParsePath(request)

	if strings.Contains(path, "echo") {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		conn.Write([]byte("Content-Type: text/plain\r\n"))
		conn.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(msg))))
		conn.Write([]byte("\r\n"))
		conn.Write([]byte(fmt.Sprintf("%s\r\n", msg)))
	} else if path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func ReadHttpRequest(conn net.Conn) string {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Failed to read http request: ", err.Error())
		os.Exit(1)
	}

	return string(buffer)
}

func ParsePath(s string) (string, string) {
	parts := strings.Split(s, " ")
	if len(parts) < 2 {
		fmt.Println("No path to extract")
		os.Exit(1)
	}

	path := parts[1]
	for _, part := range parts {
		fmt.Println(part)
	}
	return ExtractMessage(path)
}

func ExtractMessage(path string) (string, string) {
	if len(path) < 5 {
		return path, ""
	}

	return path[0:5], path[6:]
}

func CreateResponse(msg string) string {
	s := fmt.Sprintf(`HTTP/1.1 200 OK\r\n`+
		`Content-Type: text/plain\r\n`+
		`Content-Length: %d\r\n`+
		`\r\n`+
		`%s\r\n`, len(msg), msg)
	fmt.Printf("Response: %s", s)
	return s
}
