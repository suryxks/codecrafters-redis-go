package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type RedisKey string
type ResdisValue struct {
	value     string
	px        time.Duration
	savedTime time.Time
}

type RedisDB struct {
	Store map[RedisKey]ResdisValue
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	// var response = []byte("+PONG\r\n")
	listner, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	db := RedisDB{
		Store: make(map[RedisKey]ResdisValue),
	}
	for {
		var connection net.Conn
		// buf := make([]byte, 1024)
		connection, err = listner.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
			break
		}
		defer connection.Close()

		go handleConnection(connection, &db)
	}

}

func handleConnection(conn net.Conn, db *RedisDB) {
	defer conn.Close()
	for {
		resp := NewResp(conn)

		value, err := resp.Read()

		if err != nil {
			fmt.Println(err.Error())
		}
		writer := NewWriter(conn)

		switch value.typ {
		case "array":
			{
				switch strings.ToLower(value.array[0].bulk) {
				case "echo":
					writer.Write(Value{typ: "bulk", bulk: value.array[1].bulk})
				case "ping":
					writer.Write(Value{typ: "bulk", bulk: "PONG"})
				case "set":
					{
						lenOfArr := len(value.array)
						if lenOfArr < 5 {
							db.Store[RedisKey(value.array[1].bulk)] = ResdisValue{
								value:     (value.array[2].bulk),
								savedTime: time.Now(),
							}
						} else {
							timeout, err := strconv.Atoi(value.array[4].bulk)
							if err != nil {
								fmt.Println(err.Error())
								os.Exit(1)
							}
							db.Store[RedisKey(value.array[1].bulk)] = ResdisValue{
								value:     (value.array[2].bulk),
								px:        time.Duration(timeout),
								savedTime: time.Now(),
							}
						}

						writer.Write(Value{typ: "string", str: "OK"})
					}
				case "get":

					{
						key := RedisKey(value.array[1].bulk)
						val := db.Store[key]
						if time.Since(val.savedTime) > val.px {
							delete(db.Store, key)
						}
						if val.value == "" {
							writer.Write(Value{typ: "error", str: "not found"})
						} else {
							writer.Write(Value{typ: "bulk", bulk: string(val.value)})
						}

					}
				}
			}
		case "string":
			{
				switch strings.ToLower(value.str) {
				case "ping":
					writer.Write(Value{typ: "bulk", bulk: "PONG"})
				}
			}
		default:
			writer.Write(Value{typ: "string", str: "OK"})
		}

	}

}
