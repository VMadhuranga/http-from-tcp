package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("error resolving udp address: %v", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("error dialing udp: %v", err)
	}
	defer udpConn.Close()

	bufReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		data, err := bufReader.ReadString('\n')
		if err != nil {
			log.Printf("error reading string: %v", err)
			break
		}

		_, err = udpConn.Write([]byte(data))
		if err != nil {
			log.Printf("error writing data: %v", err)
			break
		}
	}

}
