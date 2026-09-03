package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("Listening on port :6379")
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Failed to bind to port", err)

	}

	conn, err := listener.Accept()

	if err != nil {
		fmt.Println("Failed to accept connection", err)
	}
	defer conn.Close()

	for {
		resp := NewResp(conn)

		value, err := resp.Read()
		if err != nil {
			fmt.Println("Failed to read from connection", err)
			return
		}

		fmt.Println(value)

		conn.Write([]byte("+OK\r\n"))

	}

}
