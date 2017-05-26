package admin

import (
	"net"
	"fmt"
	"encoding/gob"
	"log"
	"strings"
	//"time"
	"time"
)

type Db map[string]map[string]bool
type Node struct {
	ManagementAddress string
	ClientAddress     string
}

// initial values

type State struct {
	Self     Node `json:"self"`
	Nodes    []Node `json:"nodes"`
	Database Db `json:"database"`
	Tick     int `json:"tick"`
}

func GenerateListOfNodes(ns []string) []Node {

	nodes := []Node{}

	for _, n := range ns {
		addresses := strings.Split(n, ",")
		nodes = append(nodes, Node{
			ClientAddress:     addresses[0],
			ManagementAddress: addresses[1],
		})
	}

	return nodes

}

func NewState(self string, selfClient string) *State {

	return &State{
		Self: Node{
			ManagementAddress: self,
			ClientAddress:     selfClient,
		},
		Nodes:    []Node{},
		Database: make(Db),
		Tick:     0,
	}

}

func (s *State) assesMissingNodes(newState State) {
	nodesReceived := newState.Nodes
	nodesReceived = append(nodesReceived, newState.Self)

	//fmt.Printf("==> received a new self %v\n", newState.Self)
	//fmt.Printf("---> A view of the received total nodes %v\n", nodesReceived)

	s3 := append(s.Nodes, newState.Nodes...)
	s3 = append(s3, s.Self)
	s3 = append(s3, newState.Self)

	set := make(map[string]Node)

	for _, el := range s3 {

		set[el.ManagementAddress] = el
	}

	s4 := []Node{}

	for key := range set {
		if key != s.Self.ManagementAddress {
			s4 = append(s4, set[key])
		}
	}

	fmt.Printf("---> [New topography] Everything except me %v\n", s4)
	s.Nodes = s4

}

func handle(conn net.Conn, state *State) {
	dec := gob.NewEncoder(conn)
	dec.Encode(state)

	var s State
	rec := gob.NewDecoder(conn)
	rec.Decode(&s)

	//fmt.Printf("==> %v has the following nodes %v\n", s.Self, s.Nodes)
	state.assesMissingNodes(s)


}

func GetConfigFromAvailableServers( state *State) {

	channel := make(chan State)


	for _, server := range state.Nodes {

		fmt.Printf("==> Sending request to %v\n", server.ManagementAddress)

		ticker := time.NewTicker(5 * time.Second)
		quit := make(chan struct{})

		go func(ch1 chan State, state *State) {
			for {
				select {
				case <-ticker.C:

					go func(ch chan State,s string, state *State) {


						fmt.Printf("--> Available servers to fetch info %v\n",servers)
						fmt.Printf("--> getting latest information from %v\n\n", s)

						conn, err := net.Dial("tcp", s)
						if err != nil {
							return
						}
						defer conn.Close()

						// Receive the state from all connections
						var receivedState State
						encoder := gob.NewDecoder(conn)
						encoder.Decode(&receivedState)

						// Send my initial state to all nodes
						//fmt.Printf("--> Sent my state - %v\n", state)
						send := gob.NewEncoder(conn)
						send.Encode(state)

						ch <- receivedState

					}(ch1,server.ManagementAddress, state)
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}(channel, state)


	}

	for s := range channel {

		fmt.Printf("---> RECEIVED STATE <---\n")
		fmt.Printf("%v\n",s)
		fmt.Printf("---> END RECEIVED STATE <---\n")

		if s.Tick > state.Tick {
			state.Tick = s.Tick
			state.Database = s.Database
			fmt.Printf("--> Updating to tick %v\n", s.Tick)
		}
		state.assesMissingNodes(s)


	}
}

func RunAdminServer(address string, state *State) {
	ln, err := net.Listen("tcp", ":"+address)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==> Starting Management TCP server and listening for connections...")

	for {
		conn, err := ln.Accept() // this blocks until connection or error
		if err != nil {
			// handle error
			log.Fatal(err)

		}
		go handle(conn, state) // a goroutine handles conn so that the loop can accept other connections

	}

}
