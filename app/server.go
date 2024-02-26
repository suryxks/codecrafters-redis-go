package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	var response = []byte("+PONG\r\n")
	listner, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	var connection net.Conn
	buf := make([]byte, 1024)
	connection, err = listner.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer connection.Close()

	handleConnection(connection, buf, response)

}

func handleConnection(conn net.Conn, buf []byte, response []byte) {
	defer conn.Close()
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from the connection")
	}
	_, err = conn.Write(response)
	if err != nil {
		fmt.Println("Error while writing response")
	}
}
