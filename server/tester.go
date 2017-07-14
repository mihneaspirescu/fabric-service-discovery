package main

import (
	"net"
	"log"
	"encoding/gob"
)

type db map[string]map[string]bool

type state struct {
	database db `json:"database"`
	tick     int `json:"tick"`
}


func main(){

	conn, err := net.Dial("tcp", ":3000")

	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()


	currentState := state{
		database: make(db),
		tick:     0,
	}

	encoder := gob.NewEncoder(conn)
	encoder.Encode(currentState)

}