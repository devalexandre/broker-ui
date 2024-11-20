package main

import (
	"fmt"
	"log"

	"github.com/devalexandre/brokers-ui/messaging"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := messaging.NewNats("nats://b4e57197-87b0-4717-b1af-d0e6ecb3fb1f@127.0.0.1:4222")
	if err != nil {
		log.Fatal(err)
	}

	defer nc.Close()

	fmt.Println("Connected to NATS server")
	err = nc.Subscribe("dev", func(msg *nats.Msg) {
		fmt.Println("Received message: ", string(msg.Data))
	})

	select {}
}
