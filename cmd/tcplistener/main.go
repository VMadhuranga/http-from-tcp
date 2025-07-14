package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	log.SetFlags(log.Lshortfile)

	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error announcing local network address: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error waiting for next connection: %v", err)
			break
		}

		log.Println("connection opened")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("error getting request from connection: %v", err)
			break
		}

		fmt.Printf(`Request Line:
- Method: %v
- Target: %v
- Version: %v
`,
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %v: %v\n", k, v)
		}

		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}
	log.Println("connection closed")
}
