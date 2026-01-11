package main

import (
	"fmt"
	"net"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	fmt.Println("Started server...")
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		_, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		header, err := NewHeader(buf[:HEADER_SIZE])
		if err != nil {
			fmt.Printf("failed to create header: %s\n", err)
			continue
		}

		question, _, err := NewQuestion(buf[HEADER_SIZE:])
		if err != nil {
			fmt.Printf("failed to create question: %s\n", err)
			continue
		}

		answer := NewAnswer(question.QNAME, question.QTYPE, question.QCLASS, 30)

		header_bytes := header.Serialize()
		question_bytes := question.Serialize()
		answer_bytes := answer.Serialize()
		response_buf := make([]byte, len(header_bytes)+len(question_bytes)+len(answer_bytes))
		copy(response_buf, header_bytes)
		copy(response_buf[len(header_bytes):], question_bytes)
		copy(response_buf[len(header_bytes)+len(question_bytes):], answer_bytes)

		_, err = udpConn.WriteToUDP(response_buf, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
