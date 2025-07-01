package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(conn io.ReadCloser) <-chan string {
	linesChan := make(chan string)

	go func() {
		currentLineContent := ""
		for {
			connContent := make([]byte, 8)
			_, err := conn.Read(connContent)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("error reading connection content: %s\n", err)
				break
			}

			currentLineContent += string(connContent)
			currentLineContentSlice := strings.Split(currentLineContent, "\n")
			if len(currentLineContentSlice) > 1 {
				linesChan <- currentLineContentSlice[0]
				currentLineContent = ""
				currentLineContent += string(currentLineContentSlice[1])
			}
		}

		linesChan <- currentLineContent
		close(linesChan)
	}()

	return linesChan
}

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

		linesChan := getLinesChannel(conn)
		for line := range linesChan {
			fmt.Printf("%s\n", line)
		}

		log.Println("connection closed")
	}
}
