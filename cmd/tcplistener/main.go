package main

import (
	"fmt"
	"http-server/internal/request"
	"log"
	"net"
)

func main() {
	listner, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Fatal("error :", err)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error :", err)
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		r.Header.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n,v)
		})
	}
}
