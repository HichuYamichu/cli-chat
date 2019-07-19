package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/gomodule/redigo/redis"
)

var (
	allClients      = make(map[string]net.Conn)
	newConnections  = make(chan client)
	deadConnections = make(chan client)
	messages        = make(chan message)
	commands        = make(chan message)
	setName         = make(chan string)
)

type client struct {
	conn net.Conn
	name string
}

type message struct {
	Author string
	Value  string
}

var addr = flag.String("addr", "localhost:3000", "tcp service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	c, err := redis.Dial("tcp", "172.20.88.87:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	server, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				log.Fatal(err)
			}
			reader := bufio.NewReader(conn)
			name, err := reader.ReadString('\n')
			name = strings.TrimSuffix(name, "\n")
			clientConn := client{conn, name}
			newConnections <- clientConn
		}
	}()

	for {
		select {
		case c := <-newConnections:
			fmt.Println("connected")
			allClients[c.name] = c.conn

			go func(c client) {
				reader := bufio.NewReader(c.conn)
				for {
					msg, err := reader.ReadString('\n')
					msg = strings.Replace(msg, "\n", "", 1)
					msg = strings.Replace(msg, "\r", "", 1)
					if err != nil {
						break
					}
					if strings.HasPrefix(msg, "/") {
						commands <- message{c.name, msg}
					} else {
						messages <- message{c.name, msg}
					}
				}

				deadConnections <- c

			}(c)

		case msg := <-messages:
			fmt.Println(msg.Value)
			jsonMsg, err := json.Marshal(msg)

			if err != nil {
				log.Fatal(err)
			}
			_, err = c.Do("RPUSH", "messages", jsonMsg)
			if err != nil {
				log.Fatal(err)
			}
			for clientName, conn := range allClients {
				go func(conn net.Conn, clientName string, jsonMsg []byte) {
					conn.Write(jsonMsg)
					if err != nil {
						deadConnections <- client{conn, clientName}
						log.Fatal(err)
					}
				}(conn, clientName, jsonMsg)
			}
		case cmd := <-commands:
			if strings.HasPrefix(cmd.Value, "/fetch") {
				fmt.Println(cmd.Value)
				arg := strings.Replace(cmd.Value, "/fetch ", "", 1)
				res, err := redis.ByteSlices(c.Do("LRANGE", "messages", "-"+arg, -1))

				if err != nil {
					fmt.Println(err)
					return
				}
				for _, jsonMsg := range res {
					allClients[cmd.Author].Write(jsonMsg)
					// msg := message{}
					// err = json.Unmarshal([]byte(stringMsg), &msg)
					// enc := gob.NewEncoder(allClients[cmd.Author])
					// err := enc.Encode(msg)

					if err != nil {
						deadConnections <- client{allClients[cmd.Author], cmd.Author}
						log.Fatal(err)
					}

				}
			}

		case c := <-deadConnections:
			log.Printf("Client %s disconnected", c.name)
			delete(allClients, c.name)
		}
	}
}
