package main

import (
	"fmt"

	"github.com/zmj/sf-ingest/server"
)

func main() {
	srv := server.Server{}
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Printf("dead: %v\n", err)
	}
	fmt.Println("done")
}
