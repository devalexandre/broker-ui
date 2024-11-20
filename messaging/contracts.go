package messaging

var ClientType = []string{"NATS", "Kafka"}

type MessagingClient interface {
	Publish(subject string, data []byte) error
	Subscribe(subject string, cb func(msg []byte)) error
	Close()
}
