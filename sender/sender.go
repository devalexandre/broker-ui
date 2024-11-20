package main

import (
	"log"

	"github.com/devalexandre/brokers-ui/messaging"
)

func main() {

	nc, err := messaging.NewNats("nats://b4e57197-87b0-4717-b1af-d0e6ecb3fb1f@127.0.0.1:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	err = nc.Publish("dev", []byte("hello "))
	if err != nil {
		log.Fatal("Error publishing message: ", err)
	}
}
