package main

import (
	"bufio"
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

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {
	message, _ := bufio.NewReader(conn).ReadString('\n')
	path, msg := ParsePath(message)

	if strings.Contains(path, "/echo/") {
		conn.Write([]byte(CreateResponse(msg)))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
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
	return path[0:5], path[5:]
}

func CreateResponse(msg string) string {
	s := fmt.Sprintf(
		`HTTP/1.1 200 OK\r\n
		Content-Type: text/plain\r\n
		Content-Length: %v\r\n
		\r\n
		%v\r\n\r\n`, len(msg), msg)
	return s
}
