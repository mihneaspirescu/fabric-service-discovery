package admin

import (
	"net"
	"fmt"
	"encoding/gob"
	"log"

	//"time"
	"time"
	"encoding/json"
)

type Db map[string]map[string]bool
type NodeRaw struct {
	ClientAddr     string `json:"client_addr"`
	ManagementAddr string `json:"management_addr"`
	Name           string `json:"name"`
}

type Node struct {
	ManagementAddress net.TCPAddr `json:"management_addr"`
	ClientAddress     net.TCPAddr `json:"client_addr"`
	Name              string `json:"name"`
}

// initial values

type State struct {
	Self     Node `json:"self"`
	Nodes    []Node `json:"nodes"`
	Database Db `json:"database"`
	Tick     int `json:"tick"`
}



func ToNode(node NodeRaw) Node {

	managementAddr, err := net.ResolveTCPAddr("tcp", node.ManagementAddr)
	if err != nil {
		fmt.Printf(err.Error())
	}
	if managementAddr.IP == nil {
		managementAddr = &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 5001,
		}
	}

	//TODO to checks these values and cases.
	if managementAddr.Port == 0 {
		managementAddr = &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 5001,
		}
	}
	clientAddr, _ := net.ResolveTCPAddr("tcp", node.ClientAddr)

	return Node{
		ManagementAddress: *managementAddr,
		ClientAddress:     *clientAddr,
		Name:              node.Name,

	}

}

func NewState(j []byte) *State {

	var a struct {
		Self  NodeRaw
		Nodes []NodeRaw
	}

	json.Unmarshal([]byte(j), &a)

	serverState := State{
		Self:  ToNode(a.Self),
		Nodes:    []Node{},
		Database: make(Db),
		Tick:     0,
	}

	for _, el := range a.Nodes {
		serverState.Nodes = append(serverState.Nodes, ToNode(el))
	}

	return &serverState

}



func (s *State) assesMissingNodes(newState State) {
	nodesReceived := newState.Nodes
	nodesReceived = append(nodesReceived, newState.Self)

	s3 := append(s.Nodes, newState.Nodes...)
	s3 = append(s3, s.Self)
	s3 = append(s3, newState.Self)

	set := make(map[string]Node)

	for _, el := range s3 {

		set[el.ManagementAddress.String()] = el
	}

	s4 := []Node{}

	for key := range set {
		if key != s.Self.ManagementAddress.String() {
			s4 = append(s4, set[key])
		}
	}

	s.Nodes = s4

}

func handle(conn net.Conn, state *State) {
	dec := gob.NewEncoder(conn)
	dec.Encode(state)

	var s State
	rec := gob.NewDecoder(conn)
	rec.Decode(&s)

	state.assesMissingNodes(s)

}

func GetConfigFromAvailableServers(state *State) {

	channel := make(chan State)

	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})

	go func(ch1 chan State, state *State) {
		for {
			select {
			case <-ticker.C:

				//fmt.Printf("--> Available servers to fetch info %v\n", state.Nodes)

				for _, server := range state.Nodes {
					go func(ch chan State, s string, state *State) {

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

						send := gob.NewEncoder(conn)
						send.Encode(state)

						ch <- receivedState

					}(ch1, server.ManagementAddress.String(), state)
				}

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(channel, state)

	for s := range channel {

		if s.Tick > state.Tick {
			state.Tick = s.Tick
			state.Database = s.Database
			fmt.Printf("--> Updating to tick %v\n", s.Tick)
		}
		state.assesMissingNodes(s)

	}
}

func RunAdminServer(address string, state *State) {
	ln, err := net.Listen("tcp", address)
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
