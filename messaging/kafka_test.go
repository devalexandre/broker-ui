package messaging

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

const (
	unsecuredBroker = "localhost:9092"
	securedBroker   = "localhost:9093"
	testTopic       = "test-topic"
)

func TestNewKafka(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		config  *sarama.Config
		wantErr bool
	}{
		{
			name:    "Unsecured Kafka",
			brokers: []string{unsecuredBroker},
			config:  nil,
			wantErr: false,
		},
		{
			name:    "Secured Kafka",
			brokers: []string{securedBroker},
			config:  getSecuredConfig(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, err := NewKafka(tt.brokers, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, k)
				defer k.Close()
			}
		})
	}
}

func TestKafkaPublishSubscribe(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		config  *sarama.Config
	}{
		{
			name:    "Unsecured Kafka",
			brokers: []string{unsecuredBroker},
			config:  nil,
		},
		{
			name:    "Secured Kafka",
			brokers: []string{securedBroker},
			config:  getSecuredConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, err := NewKafka(tt.brokers, tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, k)
			defer k.Close()

			message := []byte("test message")
			err = k.Publish(testTopic, message)
			assert.NoError(t, err)

			receivedChan := make(chan []byte)
			err = k.Subscribe(testTopic, func(msg []byte) {
				receivedChan <- msg
			})
			assert.NoError(t, err)

			select {
			case received := <-receivedChan:
				assert.Equal(t, message, received)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for message")
			}
		})
	}
}

func getSecuredConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Net.SASL.Enable = true
	config.Net.SASL.User = os.Getenv("KAFKA_ADMIN_USERNAME")
	config.Net.SASL.Password = os.Getenv("KAFKA_ADMIN_PASSWORD")
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	return config
}
