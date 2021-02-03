package main

import (
	"h24gochallenge/webalayzer"
	"log"
)

func main() {
	if err := webalayzer.RunWebServer(":8080"); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
