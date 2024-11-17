package natscli

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

type Nats struct {
	urls string
	nc   *nats.Conn
}

func NewNats(natsURL string) (*Nats, error) {
	// Connect to the NATS server
	nc, err := nats.Connect(natsURL)
	if err != nil {
		fmt.Println("Error connecting to NATS server: ", err)
		fmt.Println("NATS URL: ", natsURL)
		return nil, err
	}

	return &Nats{
		urls: natsURL,
		nc:   nc,
	}, nil
}

// Publish publishes a message to the given subject
func (n *Nats) Publish(subject string, data []byte) error {
	fmt.Println("Publishing to subject: ", subject)
	fmt.Println("Data: ", string(data))

	return n.nc.Publish(subject, data)
}

// Subscribe subscribes to the given subject
func (n *Nats) Subscribe(subject string, cb nats.MsgHandler) error {
	fmt.Println("Subscribing to subject: ", subject)

	_, err := n.nc.Subscribe(subject, cb)
	if err != nil {
		fmt.Println("Error subscribing to subject: ", subject)
		return err
	}

	return nil
}

// Close closes the connection to the NATS server
func (n *Nats) Close() {
	n.nc.Close()
}
