package main

import (
	"log"

	"github.com/devalexandre/nats-ui/natscli"
)

func main() {

	nc, err := natscli.NewNats("nats://a428978a-7bce-4bcb-a082-a760643edd00@127.0.0.1:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	err = nc.Publish("dev", []byte("hello "))
	if err != nil {
		log.Fatal("Error publishing message: ", err)
	}
}
