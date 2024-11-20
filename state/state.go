package state

var (
	// receivedMessages stores messages received for each subscription
	ReceivedMessages = make(map[string][]string)

	// sentMessages can also be added here if you have similar needs for sent messages
	SentMessages = make(map[string][]string)
)
