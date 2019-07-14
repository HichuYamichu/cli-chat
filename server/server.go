package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

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
	author string
	value  string
}

var addr = flag.String("addr", "localhost:3000", "tcp service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	server, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
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
					incoming, err := reader.ReadString('\n')
					if err != nil {
						break
					}
					messages <- message{clientName, incoming}
				}

				deadConnections <- conn

			}(c.conn, allClients[c.conn])

		case msg := <-messages:
			fmt.Print(msg.value)
			for conn, clientName := range allClients {
				if clientName == msg.author {
					continue
				}
				go func(conn net.Conn, msg message) {
					m := fmt.Sprintf("<%s> %s", msg.author, msg.value)
					_, err := conn.Write([]byte(m))
					if err != nil {
						deadConnections <- conn
					}
				}(conn, msg)
			}

		case conn := <-deadConnections:
			log.Printf("Client %s disconnected", allClients[conn])
			delete(allClients, conn)
		}
	}
}
