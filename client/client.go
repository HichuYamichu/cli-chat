package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type message struct {
	Author string
	Value  string
}

var addr = flag.String("addr", "localhost:3000", "tcp service address")
var name = flag.String("name", "jan", "Username.")

func main() {
	flag.Parse()
	log.SetFlags(0)

	fmt.Printf("Connecting to %s...\n", *addr)
	conn, err := net.Dial("tcp", *addr)

	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Write([]byte(*name + "\n"))
	if err != nil {
		fmt.Println("Error writing to stream.")
	}

	go readConnection(conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		pattern := "<" + *name + "> "
		fmt.Print(pattern)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, pattern, "", 1)

		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func readConnection(conn net.Conn) {
	var msg message
	for {
		dec := gob.NewDecoder(conn)
		err := dec.Decode(&msg)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		m := fmt.Sprintf("\n<%s> %s", msg.Author, msg.Value)
		fmt.Print(m)
	}
}
