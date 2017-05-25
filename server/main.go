package main

import (
	"net"
	"fmt"
	"log"
	"bufio"

	"strings"

	"encoding/json"
	"os"
)

type db map[string]map[string]bool

type Sibling struct {
	address string
	conn    net.Conn
}
type Update struct {
	topic string
	value string
}

var siblings []Sibling

func set(topic string, address string) string {
	return "SET " + topic + " " + address
}

func writeToSibligs(c chan Update) {

	for val := range c {
		fmt.Printf("==> Received a message to send to siblings %v\n", val)
		for _, s := range siblings {
			fmt.Printf("=> Sending to %v\n", s.address)
			go fmt.Fprint(s.conn, set(val.topic, val.value))
			fmt.Println("i don't get it")
		}
	}

	fmt.Println("closing .....")

}

func handleConnection(conn net.Conn, data db) {

	scanner := bufio.NewScanner(conn)

	toSiblings := make(chan Update)

	go writeToSibligs(toSiblings)

	for scanner.Scan() {

		a := scanner.Text()
		w := strings.Fields(a)

		if len(w) < 1 {
			fmt.Fprintf(conn, "%v\n", "==> Available commands GET and SET")
			continue
		}

		switch w[0] {
		case "SET":
			fmt.Printf("will set -> %v with value: %v\n", w[1], w[2])
			if len(w) != 3 {
				fmt.Fprint(conn, "==> SET must have a value\n")
				continue
			}
			if _, ok := data[w[1]]; !ok {
				data[w[1]] = map[string]bool{w[2]: true, }
				toSiblings <- Update{
					topic: w[1],
					value: w[2],
				}
				continue
			}

			mapS := data[w[1]]
			mapS[w[2]] = true
			toSiblings <- Update{
				topic: w[1],
				value: w[2],
			}

		case "GET":
			if len(w) != 2 {
				fmt.Fprint(conn, "==> GET must have a KEY \n")
				continue
			}
			i, ok := data[w[1]]
			fmt.Printf("nr of map elements - > %v\n", len(i))

			if ok == true {
				var t []string
				for k := range i {
					t = append(t, k)
				}

				str, _ := json.Marshal(t)
				fmt.Fprintf(conn, "%v\n", string(str))

			} else {
				fmt.Fprintf(conn, "--> %v\n", "No such key")
			}
		case "TOPICS":
			var keys []string
			for k := range data {
				keys = append(keys, k)
			}
			str, _ := json.Marshal(keys)
			fmt.Fprintf(conn, "%v\n", string(str))

		case "DEL":
			fmt.Fprintf(conn, "==> DELETED: %v\n", data[w[1]])
			delete(data, w[1])
		default:
			fmt.Fprintf(conn, "==> %v\n", "Command not found")
		}

	}

	defer conn.Close()
}

func main() {

	// args[0] => portServer
	// args[1] => all other instances addresses
	args := os.Args[1:]

	if len(args) > 1 {
		servers := strings.Split(args[1], ",")

		fmt.Println(servers)

		for _, s := range servers {

			conn, err := net.Dial("tcp", s)

			if err != nil {
				fmt.Println(err)
				continue
			}
			siblings = append(siblings, Sibling{
				address: s,
				conn:    conn,
			})
		}

		fmt.Println(siblings)
	}

	ln, err := net.Listen("tcp", ":"+args[0])
	if err != nil {
		log.Fatal(err)
		// handle error
	}

	data := make(db)

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Panic(err)

		}
		go handleConnection(conn, data)
	}

}
