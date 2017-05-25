package main

import (
	"net"
	"fmt"
	"log"
	"bufio"

	"strings"

	"encoding/json"
)


type db map[string]map[string]bool

func handleConnection(conn net.Conn, data db) {



	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {

		a := scanner.Text()
		w := strings.Fields(a)

		if len(w) < 1 {
			fmt.Fprintf(conn, "%v\n", "==> Available commands GET and SET")
			continue
		}

		switch w[0] {
		case "SET":
			fmt.Printf("will set -> %v with value: %v\n", w[1], w[2] )
			if len(w) != 3 {
				fmt.Fprint(conn, "==> SET must have a value\n")
				continue
			}
			if _, ok := data[w[1]]; !ok {
				data[w[1]] = map[string]bool{w[2]: true, }
				continue
			}

			mapS := data[w[1]]
			mapS[w[2]] = true

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
		case "EXIT":
			conn.Close()
		default:
			fmt.Fprintf(conn, "==> %v\n", "Command not found")
		}

	}

	defer conn.Close()
}

func main() {

	ln, err := net.Listen("tcp", ":3500")
	if err != nil {
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
