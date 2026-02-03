package main

import (
	"bytes"
	"fmt"
	"http-server/internal/request"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer f.Close()
		defer close(out)
		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}

			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				out <- str
				str = ""
			}
			str += string(data)
		}
		if len(str) != 0 {
			out <- str
		}
	}()
	return out
}
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
	}
}
