package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	dir := flag.String("directory", ".", "Directory")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go HandleConnection(conn, *dir)
	}
}

func HandleConnection(conn net.Conn, dir string) {
	defer conn.Close()

	request := ReadHttpRequest(conn)
	fmt.Printf("Request: %s", request)
	path, msg := ParsePath(request)

	fmt.Printf("Path: %s Msg: %s\n", path, msg)

	if strings.Contains(path, "echo") {
		fmt.Println("ECHO BRANCH")
		conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		conn.Write([]byte("Content-Type: text/plain\r\n"))
		conn.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(msg))))
		conn.Write([]byte("\r\n"))
		conn.Write([]byte(fmt.Sprintf("%s\r\n", msg)))
	} else if strings.Contains(path, "user-agent") {
		fmt.Println("USER-AGENT BRANCH")
		body := ParseUserAgent(request)
		fmt.Printf("Body: %s len: %v\n", body, len(body))
		conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		conn.Write([]byte("Content-Type: text/plain\r\n"))
		conn.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body))))
		conn.Write([]byte("\r\n"))
		conn.Write([]byte(fmt.Sprintf("%s", body)))
	} else if strings.Contains(path, "files") {
		method := ParseMethod(request)
		if method == "GET" {
			fmt.Printf("GET BRANCH\n")
			fmt.Printf("Opening file %s%s", dir, msg)
			file, err := os.ReadFile(dir + msg)
			if err == nil {
				body := string(file)
				fmt.Printf("File body:\n%s\n", body)
				conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
				conn.Write([]byte("Content-Type: application/octet-stream\r\n"))
				conn.Write([]byte(fmt.Sprintf("Content-Length: %d\r\n", len(body))))
				conn.Write([]byte("\r\n"))
				conn.Write([]byte(fmt.Sprintf("%s\r\n", body)))
			} else {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			}
		} else {
			fmt.Printf("POST BRANCH\n")
			body := ParseBody(request)
			fmt.Printf("Parsed body: %s\n (len: %d)", body, len(body))
			err := os.WriteFile(dir+msg, []byte(body), 0777)
			os.Truncate(dir+msg, int64(len(body)))
			if err != nil {
				fmt.Printf("Write to %s%s failed: %s", dir, msg, err.Error())
				os.Exit(1)
			}
			conn.Write([]byte("HTTP/1.1 201 CREATED\r\n\r\n"))
		}
	} else if path == "/" {
		fmt.Println("/ BRANCH")
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		fmt.Println("404 BRANCH")
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

func ParseMethod(s string) string {
	lines := strings.Split(s, "\n")
	parts := strings.Split(lines[0], " ")
	fmt.Printf("Method parsed: %s\n", parts[0])
	return parts[0]
}

func ParseBody(s string) string {
	lines := strings.Split(s, "\r\n")
	for i, line := range lines {
		if line != "" {
			continue
		}

		b := []byte(lines[i+1])

		return string(bytes.Trim(b, "\x00"))
	}
	return ""
}

func ParseUserAgent(s string) string {
	parts := strings.Split(s, "\n")
	if len(parts) < 3 {
		fmt.Println("Malformed HTTP request")
		os.Exit(1)
	}

	for _, part := range parts {
		if !strings.Contains(part, "User-Agent:") {
			continue
		}

		userAgent := strings.Split(part, ":")
		return strings.TrimSpace(userAgent[1])
	}
	return ""
}

func ParsePath(s string) (string, string) {
	parts := strings.Split(s, " ")
	if len(parts) < 2 {
		fmt.Println("No path to extract")
		os.Exit(1)
	}
	path := parts[1]
	return ExtractMessage(path)
}

func ExtractMessage(path string) (string, string) {
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		return ("/" + parts[1]), ""
	}

	return ("/" + parts[1]), strings.Join(parts[2:], "/")
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
