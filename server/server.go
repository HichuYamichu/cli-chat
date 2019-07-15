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

// REVERSE KEYS

var (
	allClients      = make(map[net.Conn]string)
	newConnections  = make(chan client)
	deadConnections = make(chan net.Conn)
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
			allClients[c.conn] = c.name

			go func(conn net.Conn, clientName string) {
				reader := bufio.NewReader(conn)
				for {
					msg, err := reader.ReadString('\n')
					msg = strings.Replace(msg, "\n", "", 1)
					if err != nil {
						break
					}
					messages <- message{clientName, msg}
				}

				deadConnections <- conn

			}(c.conn, allClients[c.conn])

		case msg := <-messages:
			fmt.Print(msg.Value)
			for conn, clientName := range allClients {
				if clientName == msg.Author {
					continue
				}
				go func(conn net.Conn, msg message) {
					enc := gob.NewEncoder(conn)
					err := enc.Encode(msg)
					if err != nil {
						deadConnections <- conn
						log.Fatal(err)
					}
				}(conn, msg)
			}

		case conn := <-deadConnections:
			log.Printf("Client %s disconnected", allClients[conn])
			delete(allClients, conn)
		}
	}
}
