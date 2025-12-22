package main

import (
	"fmt"
	"log"

	natscli "github.com/devalexandre/broker-ui/tests/nats"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := natscli.NewNats("nats://a428978a-7bce-4bcb-a082-a760643edd00@127.0.0.1:4222")
	if err != nil {
		log.Fatal(err)
	}

	defer nc.Close()

	fmt.Println("Connected to NATS server")
	err = nc.Subscribe("local", func(msg *nats.Msg) {
		fmt.Println("Received message: ", string(msg.Data))
	})

	select {}
}
