package main

import (
	"fmt"
	"time"

	"github.com/zmj/sf-ingest/server"
)

func main() {
	session, err := server.NewSession(9284)
	if err != nil {
		fmt.Printf("Error making server: %v\n", err)
		return
	}
	err = session.ListenOne()
	if err != nil {
		fmt.Printf("Error receiving msg: %v\n", err)
		return
	}
	fmt.Println("yay")
	time.Sleep(time.Second)
}
