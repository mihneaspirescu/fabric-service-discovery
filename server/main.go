package main

import (
	"net"
	"fmt"
	//"log"
	"bufio"

	"strings"

	"encoding/json"
	"os"
	"log"
	"encoding/gob"
	"github.com/mihneaspirescu/fabric-discovery/server/admin"
)


var serversAdmin []string

type Update struct {
	topic string
	value string
}

func set(topic string, address string) string {
	return "SETADMIN " + topic + " " + address
}

func writeToSibligs(servers []admin.Node,c chan Update) {

	for val := range c {
		fmt.Printf("==> Received a message to send to siblings %v\n", val)
		for _, s := range servers {
			fmt.Printf("=> Sending to %v\n", s.ClientAddress)

			conn, err := net.Dial("tcp", s.ClientAddress)

			if err != nil {
				fmt.Printf("--> Error writing replication to server %v", s)
				fmt.Printf("--> /t %v/n", err)
				continue
			}

			fmt.Fprint(conn, set(val.topic, val.value))
			conn.Close()
		}
	}

}

func handleMessage(conn net.Conn, currentState *admin.State, w []string) {

	data := currentState.Database
	toSiblings := make(chan Update)
	go writeToSibligs(currentState.Nodes,toSiblings)



	switch w[0] {
	case "SETADMIN":
		fmt.Printf("=====> ADMIN RECEIVE - %v - %v \n", w[1], w[2])
		if _, ok := data[w[1]]; !ok {
			data[w[1]] = map[string]bool{w[2]: true, }
			currentState.Tick++
			break

		}
		mapS := data[w[1]]
		mapS[w[2]] = true
		currentState.Tick++

	case "SET":
		fmt.Printf("will set -> %v with value: %v\n", w[1], w[2])
		if len(w) != 3 {
			fmt.Fprint(conn, "==> SET must have a value\n")

		}
		if _, ok := data[w[1]]; !ok {
			data[w[1]] = map[string]bool{w[2]: true, }
			currentState.Tick++
			toSiblings <- Update{
				topic: w[1],
				value: w[2],
			}
			break

		}

		mapS := data[w[1]]
		mapS[w[2]] = true
		currentState.Tick++
		toSiblings <- Update{
			topic: w[1],
			value: w[2],
		}

	case "GET":
		if len(w) != 2 {
			fmt.Fprint(conn, "==> GET must have a KEY \n")

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
		toSiblings <- Update{
			topic: w[1],
			value: w[2],
		}
		currentState.Tick++
	case "NODES":
		fmt.Fprintf(conn, "==> NODES: %v\n", currentState.Nodes)
	case "INIT":
		fmt.Println("received a init call")
		encoder := gob.NewEncoder(conn)
		encoder.Encode(currentState)

	default:
		fmt.Fprintf(conn, "==> %v\n", "Command not found")
	}

}

func handleConnection(conn net.Conn, currentState *admin.State) {

	scanner := bufio.NewScanner(conn)



	for scanner.Scan() {

		a := scanner.Text()
		w := strings.Fields(a)

		if len(w) < 1 {
			fmt.Fprintf(conn, "%v\n", "==> Available commands GET and SET")
			continue
		}

		go handleMessage(conn, currentState, w)

	}

	defer conn.Close()
}




func main() {


	// args[0] => portServer
	// args[1] => adminPort
	// args[2] ...  => pairs of :portServer,:adminServer
	args := os.Args[1:]

	portClient := args[0]
	portAdministration := args[1]


	nodes := []admin.Node{}

	if len(args) > 2 {
		nodes = admin.GenerateListOfNodes(args[2:])

	}


	// generate an initial state for the server
	// generates the self node
	currentState := admin.NewState(":"+portAdministration, ":"+portClient)

	fmt.Printf("*** the current node is %v with management %v and client %v \n", currentState.Self, currentState.Self.ManagementAddress, currentState.Self.ClientAddress)

	if len(nodes) != 0 {
		fmt.Printf("==> Searching for configurations %v\n", nodes)
		currentState.Nodes = nodes
		go admin.GetConfigFromAvailableServers(currentState)
	}

	// runs the admin server which runs on port args[1]
	// and parses the GOB for configuration
	go admin.RunAdminServer(args[1], currentState)

	//normal server that parses the input.
	ln, err := net.Listen("tcp", ":"+args[0])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("==> Starting TCP server and listening for connections...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Panic(err)

		}
		go handleConnection(conn, currentState)
	}

}
