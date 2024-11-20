package messaging

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type Kafka struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
	brokers  []string
	config   *sarama.Config
}

func NewKafka(brokers []string, config *sarama.Config) (*Kafka, error) {
	if config == nil {
		config = sarama.NewConfig()
	}
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Printf("Error creating Kafka producer: %v", err)
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Printf("Error creating Kafka consumer: %v", err)
		producer.Close() // Close the producer if consumer creation fails
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &Kafka{
		producer: producer,
		consumer: consumer,
		brokers:  brokers,
		config:   config,
	}, nil
}

// Publish publishes a message to the given topic
func (k *Kafka) Publish(topic string, data []byte) error {
	fmt.Printf("Publishing to topic: %s\n", topic)
	fmt.Printf("Data: %s\n", string(data))

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Error publishing to topic %s: %v\n", topic, err)
		return fmt.Errorf("failed to publish message to topic %s: %w", topic, err)
	}

	fmt.Printf("Message published to topic %s (partition: %d, offset: %d)\n", topic, partition, offset)
	return nil
}

// Subscribe subscribes to the given topic
func (k *Kafka) Subscribe(topic string, cb func(msg []byte)) error {
	fmt.Println("Subscribing to topic:", topic)

	partitions, err := k.consumer.Partitions(topic)
	if err != nil {
		log.Printf("Error getting partitions for topic %s: %v", topic, err)
		return err
	}

	for _, partition := range partitions {
		pc, err := k.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Error subscribing to topic %s, partition %d: %v", topic, partition, err)
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			for message := range pc.Messages() {
				cb(message.Value)
			}
		}(pc)
	}

	return nil
}

// Close closes the Kafka producer and consumer
func (k *Kafka) Close() error {
	var producerErr, consumerErr error

	if err := k.producer.Close(); err != nil {
		log.Println("Error closing Kafka producer:", err)
		producerErr = fmt.Errorf("failed to close Kafka producer: %w", err)
	}

	if err := k.consumer.Close(); err != nil {
		log.Println("Error closing Kafka consumer:", err)
		consumerErr = fmt.Errorf("failed to close Kafka consumer: %w", err)
	}

	if producerErr != nil || consumerErr != nil {
		return fmt.Errorf("errors closing Kafka connections: %v, %v", producerErr, consumerErr)
	}

	return nil
}
