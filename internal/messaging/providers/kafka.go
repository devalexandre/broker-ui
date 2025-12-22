package providers

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/devalexandre/broker-ui/internal/messaging"
	"github.com/segmentio/kafka-go"
)

// KafkaProvider implements MessagingProvider for Kafka using segmentio/kafka-go.
//
// Mapping:
// - subject/topic: Kafka topic name
// - Subscribe(subjectPattern): consumes from that topic using a consumer group (configurable via server URL)
//
// URL examples:
// - kafka://localhost:19092?group=broker-ui&start=latest
// - kafka://host1:9092,host2:9092?group=my-group&start=earliest
type KafkaProvider struct {
	mu        sync.RWMutex
	connected bool

	brokers     []string
	groupID     string
	startOffset int64

	writer *kafka.Writer

	subs   map[string]*kafka.Reader
	cancel map[string]context.CancelFunc
}

func NewKafkaProvider() *KafkaProvider {
	return &KafkaProvider{
		subs:   make(map[string]*kafka.Reader),
		cancel: make(map[string]context.CancelFunc),
	}
}

func (k *KafkaProvider) Connect(rawURL string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.connected {
		return nil
	}

	cfg, err := parseKafkaConfig(rawURL)
	if err != nil {
		return err
	}

	k.brokers = cfg.brokers
	k.groupID = cfg.groupID
	k.startOffset = cfg.startOffset

	k.writer = &kafka.Writer{
		Addr:                   kafka.TCP(k.brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		RequiredAcks:           kafka.RequireOne,
		WriteTimeout:           5 * time.Second,
		ReadTimeout:            5 * time.Second,
	}

	k.connected = true
	return nil
}

func (k *KafkaProvider) Publish(topic string, data []byte) error {
	k.mu.RLock()
	if !k.connected || k.writer == nil {
		k.mu.RUnlock()
		return fmt.Errorf("not connected to Kafka")
	}
	writer := k.writer
	k.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: data,
		Time:  time.Now(),
	})
}

func (k *KafkaProvider) Subscribe(topic string, handler messaging.MessageHandler) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.connected || len(k.brokers) == 0 {
		return fmt.Errorf("not connected to Kafka")
	}
	if _, exists := k.subs[topic]; exists {
		return fmt.Errorf("already subscribed to topic: %s", topic)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        append([]string(nil), k.brokers...),
		GroupID:        k.groupID,
		Topic:          topic,
		StartOffset:    k.startOffset,
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second, // auto-commit periodically (when GroupID set)
	})

	ctx, cancel := context.WithCancel(context.Background())
	k.subs[topic] = reader
	k.cancel[topic] = cancel

	go func() {
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				return
			}
			handler(msg.Topic, msg.Value)
		}
	}()

	return nil
}

func (k *KafkaProvider) Unsubscribe(topic string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	reader, ok := k.subs[topic]
	if !ok {
		return fmt.Errorf("no subscription found for topic: %s", topic)
	}
	if cancel, ok := k.cancel[topic]; ok {
		cancel()
		delete(k.cancel, topic)
	}
	delete(k.subs, topic)
	return reader.Close()
}

func (k *KafkaProvider) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.connected {
		return nil
	}
	k.connected = false

	for topic, cancel := range k.cancel {
		cancel()
		delete(k.cancel, topic)
	}
	for topic, r := range k.subs {
		_ = r.Close()
		delete(k.subs, topic)
	}
	if k.writer != nil {
		_ = k.writer.Close()
		k.writer = nil
	}
	return nil
}

func (k *KafkaProvider) IsConnected() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.connected && k.writer != nil
}

func (k *KafkaProvider) GetProviderType() messaging.ProviderType {
	return messaging.ProviderKafka
}

type kafkaConfig struct {
	brokers     []string
	groupID     string
	startOffset int64
}

func parseKafkaConfig(raw string) (kafkaConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return kafkaConfig{}, fmt.Errorf("kafka url is empty")
	}

	u := raw
	if !strings.Contains(u, "://") {
		u = "kafka://" + u
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return kafkaConfig{}, fmt.Errorf("invalid kafka url: %w", err)
	}
	if parsed.Scheme != "kafka" {
		return kafkaConfig{}, fmt.Errorf("unsupported kafka scheme: %s", parsed.Scheme)
	}

	brokers := splitCSV(parsed.Host)
	if q := parsed.Query().Get("brokers"); q != "" {
		brokers = splitCSV(q)
	}
	if len(brokers) == 0 {
		return kafkaConfig{}, fmt.Errorf("kafka brokers are empty")
	}

	groupID := parsed.Query().Get("group")
	if groupID == "" {
		groupID = "broker-ui"
	}

	startOffset := kafka.LastOffset
	switch strings.ToLower(parsed.Query().Get("start")) {
	case "", "latest":
		startOffset = kafka.LastOffset
	case "earliest", "first":
		startOffset = kafka.FirstOffset
	}

	return kafkaConfig{
		brokers:     brokers,
		groupID:     groupID,
		startOffset: startOffset,
	}, nil
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
