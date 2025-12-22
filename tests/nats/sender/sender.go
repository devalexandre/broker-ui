package main

import (
	"log"

	natscli "github.com/devalexandre/broker-ui/tests/nats"
)

func main() {

	nc, err := natscli.NewNats("nats://a428978a-7bce-4bcb-a082-a760643edd00@127.0.0.1:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	err = nc.Publish("local", []byte("hello "))
	if err != nil {
		log.Fatal("Error publishing message: ", err)
	}
}
