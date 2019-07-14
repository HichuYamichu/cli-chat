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
		if len(text) == 1 {
			continue
		}

		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func readConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Print(message)
	}
}
