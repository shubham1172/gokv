package main

import (
	"fmt"
	"github.com/shubham1172/gokv/api/v1/server"
)

// Port to start the server on.
const Port = 8080

func main() {
	server.Start(fmt.Sprintf("%s:%d", "", Port))
}
