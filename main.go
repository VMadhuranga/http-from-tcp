package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	linesChan := make(chan string)

	go func() {
		currentLineContent := ""
		for {
			fileContent := make([]byte, 8)
			_, err := f.Read(fileContent)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("error reading file: %s\n", err)
				break
			}

			currentLineContent += string(fileContent)
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
	file, err := os.Open("message.txt")
	if err != nil {
		log.Fatalf("error opening file: %s\n", err)
	}
	defer file.Close()

	linesChan := getLinesChannel(file)
	for line := range linesChan {
		fmt.Printf("read: %s\n", line)
	}
}
