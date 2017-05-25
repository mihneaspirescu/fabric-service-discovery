package main

import (
	"net"
	"log"
	"fmt"
	"os"
)

func set(topic string, address string) string {
	return "SET " + topic + " " + address
}

func main() {

	// args[0] => DS address
	// args[1] => service name
	// args[2] => address of current service
	args := os.Args[1:]

	conn, err := net.Dial("tcp", args[0])

	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	fmt.Fprint(conn, set(args[1], args[2]))
	fmt.Fprintf(os.Stdout, "Assigned the service %v to the SD found at %v\n", args[1],args[0])

}
