package admin

import (
	"net"
	"fmt"
	"encoding/gob"
	"log"
)

type Db map[string]map[string]bool

// initial values

type State struct {
	Database Db `json:"database"`
	Tick     int `json:"tick"`
}

func NewState() *State {

	return &State{
		Database: make(Db),
		Tick:     0,
	}

}
func handle(conn net.Conn, state *State) {
	dec := gob.NewEncoder(conn)
	dec.Encode(state)
	conn.Close()
}

func GetConfigFromAvailableServers(servers []string, state *State) {

	channel := make(chan State, len(servers))

	for _, server := range servers {

		fmt.Printf("==> Sending request to %v\n", server)
		go func(c chan State, s string) {

			conn, err := net.Dial("tcp", s)
			if err != nil {
				fmt.Printf("--> Error trying server: %v\n", s)
				fmt.Printf("--> \t%v\n", err)
				return
			}
			defer conn.Close()

			var state State

			encoder := gob.NewDecoder(conn)
			encoder.Decode(&state)

			c <- state

		}(channel, server)

	}

	for s := range channel {

		if s.Tick > state.Tick {
			state.Tick = s.Tick
			state.Database = s.Database
			fmt.Printf("--> Updating to tick %v\n", s.Tick)
		}

		fmt.Printf("--> Received state %v\n", s.Tick)
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
