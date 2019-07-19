package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/jroimartin/gocui"
)

type message struct {
	Author string
	Value  string
}

var addr = flag.String("addr", "localhost:3000", "tcp service address")
var name = flag.String("name", "jan", "Username.")
var conn net.Conn

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	typerY := maxY / 8
	chatY := maxY - (typerY) - 1

	if v, err := g.SetView("typer", 0, chatY, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Editable = true

		if _, err := g.SetCurrentView("typer"); err != nil {
			return err
		}
	}

	if v, err := g.SetView("chatBox", 0, 0, maxX-1, chatY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Autoscroll = true
		v.Wrap = true
	}

	return nil
}

func up(g *gocui.Gui, v *gocui.View) error {
	typer, err := g.View("chatBox")
	if err != nil {
		return err
	}
	typer.SetCursor(0, 1)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func send(g *gocui.Gui, v *gocui.View) error {
	typer, err := g.View("typer")
	if err != nil {
		return err
	}
	reader := bufio.NewReader(typer)
	text, _ := reader.ReadString('\n')
	_, err = conn.Write([]byte(text))
	if err != nil {
		return err
	}
	typer.Clear()
	typer.SetCursor(0, 0)
	return nil
}

func readConnection(conn net.Conn, g *gocui.Gui) {
	dec := json.NewDecoder(conn)
	var msg message
	for {
		dec.Decode(&msg)

		g.Update(func(g *gocui.Gui) error {
			chatBox, err := g.View("chatBox")
			if err != nil {
				return err
			}
			m := fmt.Sprintf("<%v> %v", msg.Author, msg.Value)
			_, err = fmt.Fprintln(chatBox, m)
			if err != nil {
				return err
			}
			return nil
		})
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorRed

	g.SetManagerFunc(layout)

	conn, err = net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Write([]byte(*name + "\n"))
	if err != nil {
		log.Fatal(err)
	}

	go readConnection(conn, g)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatal(err)
	}

	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, send); err != nil {
		log.Fatal(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, up); err != nil {
		log.Fatal(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatal(err)
	}
}
