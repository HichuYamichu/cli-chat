package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
)

var (
	allClients      = make(map[string]net.Conn)
	newConnections  = make(chan client)
	deadConnections = make(chan client)
	messages        = make(chan message)
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
			c := client{conn, name}
			newConnections <- c
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
					messages <- message{c.name, msg}
				}

				deadConnections <- c

			}(c)

		case msg := <-messages:
			fmt.Println(msg.Value)
			for clientName, conn := range allClients {
				go func(conn net.Conn, clientName string, msg message) {
					enc := gob.NewEncoder(conn)
					err := enc.Encode(msg)
					if err != nil {
						deadConnections <- client{conn, clientName}
						log.Fatal(err)
					}
				}(conn, clientName, msg)
			}

		case c := <-deadConnections:
			log.Printf("Client %s disconnected", c.name)
			delete(allClients, c.name)
		}
	}
}
