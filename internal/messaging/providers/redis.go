package providers

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/devalexandre/broker-ui/internal/messaging"
)

// RedisProvider implements MessagingProvider for Redis Pub/Sub.
//
// Notes:
// - Redis connections that issue SUBSCRIBE enter "pubsub mode" and cannot run arbitrary commands.
// - To keep Publish and Subscribe working concurrently, we keep two connections: one for publishing and one for subscribing.
type RedisProvider struct {
	mu        sync.RWMutex
	connected bool

	pubConn   net.Conn
	pubReader *bufio.Reader
	pubMu     sync.Mutex

	subConn   net.Conn
	subReader *bufio.Reader
	subMu     sync.Mutex

	stop chan struct{}

	handlers  map[string]messaging.MessageHandler
	isPattern map[string]bool
}

func NewRedisProvider() *RedisProvider {
	return &RedisProvider{
		handlers:  make(map[string]messaging.MessageHandler),
		isPattern: make(map[string]bool),
	}
}

func (r *RedisProvider) Connect(rawURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.connected {
		return nil
	}

	cfg, err := parseRedisConfig(rawURL)
	if err != nil {
		return err
	}

	pubConn, pubReader, err := dialRedis(cfg)
	if err != nil {
		return err
	}
	subConn, subReader, err := dialRedis(cfg)
	if err != nil {
		_ = pubConn.Close()
		return err
	}

	r.pubConn = pubConn
	r.pubReader = pubReader
	r.subConn = subConn
	r.subReader = subReader
	r.connected = true
	r.stop = make(chan struct{})

	go r.subReadLoop(r.stop)
	return nil
}

func (r *RedisProvider) Publish(subject string, data []byte) error {
	r.mu.RLock()
	if !r.connected || r.pubConn == nil || r.pubReader == nil {
		r.mu.RUnlock()
		return fmt.Errorf("not connected to Redis")
	}
	conn := r.pubConn
	reader := r.pubReader
	r.mu.RUnlock()

	r.pubMu.Lock()
	defer r.pubMu.Unlock()

	if err := writeRESPArray(conn, [][]byte{[]byte("PUBLISH"), []byte(subject), data}); err != nil {
		return fmt.Errorf("redis publish write failed: %w", err)
	}
	resp, err := readRESP(reader)
	if err != nil {
		return fmt.Errorf("redis publish read failed: %w", err)
	}
	if _, ok := resp.(int64); !ok {
		return fmt.Errorf("unexpected redis publish response: %T", resp)
	}
	return nil
}

func (r *RedisProvider) Subscribe(subjectPattern string, handler messaging.MessageHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.connected || r.subConn == nil {
		return fmt.Errorf("not connected to Redis")
	}
	if _, exists := r.handlers[subjectPattern]; exists {
		return fmt.Errorf("already subscribed to: %s", subjectPattern)
	}

	cmd := "SUBSCRIBE"
	isPattern := looksLikeRedisPattern(subjectPattern)
	if isPattern {
		cmd = "PSUBSCRIBE"
	}

	r.subMu.Lock()
	err := writeRESPArray(r.subConn, [][]byte{[]byte(cmd), []byte(subjectPattern)})
	r.subMu.Unlock()
	if err != nil {
		return fmt.Errorf("redis subscribe failed: %w", err)
	}

	r.handlers[subjectPattern] = handler
	r.isPattern[subjectPattern] = isPattern
	return nil
}

func (r *RedisProvider) Unsubscribe(subjectPattern string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.connected || r.subConn == nil {
		return fmt.Errorf("not connected to Redis")
	}
	if _, exists := r.handlers[subjectPattern]; !exists {
		return fmt.Errorf("no subscription found for: %s", subjectPattern)
	}

	cmd := "UNSUBSCRIBE"
	if r.isPattern[subjectPattern] {
		cmd = "PUNSUBSCRIBE"
	}

	r.subMu.Lock()
	err := writeRESPArray(r.subConn, [][]byte{[]byte(cmd), []byte(subjectPattern)})
	r.subMu.Unlock()
	if err != nil {
		return fmt.Errorf("redis unsubscribe failed: %w", err)
	}

	delete(r.handlers, subjectPattern)
	delete(r.isPattern, subjectPattern)
	return nil
}

func (r *RedisProvider) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.connected {
		return nil
	}
	r.connected = false
	if r.stop != nil {
		close(r.stop)
		r.stop = nil
	}

	if r.pubConn != nil {
		_ = r.pubConn.Close()
		r.pubConn = nil
	}
	if r.subConn != nil {
		_ = r.subConn.Close()
		r.subConn = nil
	}
	r.pubReader = nil
	r.subReader = nil
	r.handlers = make(map[string]messaging.MessageHandler)
	r.isPattern = make(map[string]bool)
	return nil
}

func (r *RedisProvider) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connected && r.pubConn != nil && r.subConn != nil
}

func (r *RedisProvider) GetProviderType() messaging.ProviderType {
	return messaging.ProviderRedis
}

func (r *RedisProvider) subReadLoop(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		r.mu.RLock()
		reader := r.subReader
		r.mu.RUnlock()
		if reader == nil {
			return
		}

		v, err := readRESP(reader)
		if err != nil {
			return
		}

		arr, ok := v.([]any)
		if !ok || len(arr) == 0 {
			continue
		}

		msgType := toString(arr[0])
		switch msgType {
		case "message":
			if len(arr) < 3 {
				continue
			}
			channel := toString(arr[1])
			payload := toBytes(arr[2])

			handler := r.getHandler(channel)
			if handler != nil {
				handler(channel, payload)
			}

		case "pmessage":
			// ["pmessage", pattern, channel, message]
			if len(arr) < 4 {
				continue
			}
			pattern := toString(arr[1])
			channel := toString(arr[2])
			payload := toBytes(arr[3])

			handler := r.getHandler(pattern)
			if handler != nil {
				handler(channel, payload)
			}
		default:
			// ignore subscribe/psubscribe/unsubscribe/punsubscribe and any other events
		}
	}
}

func (r *RedisProvider) getHandler(key string) messaging.MessageHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.handlers[key]
}

type redisConfig struct {
	addr     string
	username string
	password string
	db       int
	tls      bool
}

func parseRedisConfig(rawURL string) (redisConfig, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return redisConfig{}, fmt.Errorf("redis url is empty")
	}

	u := rawURL
	if !strings.Contains(u, "://") {
		u = "redis://" + u
	}

	parsed, err := url.Parse(u)
	if err != nil {
		return redisConfig{}, fmt.Errorf("invalid redis url: %w", err)
	}

	switch parsed.Scheme {
	case "redis", "rediss":
	default:
		return redisConfig{}, fmt.Errorf("unsupported redis scheme: %s", parsed.Scheme)
	}

	host := parsed.Hostname()
	if host == "" {
		return redisConfig{}, fmt.Errorf("redis host is empty")
	}
	port := parsed.Port()
	if port == "" {
		port = "6379"
	}

	db := 0
	if parsed.Path != "" && parsed.Path != "/" {
		p := strings.TrimPrefix(parsed.Path, "/")
		if p != "" {
			n, err := strconv.Atoi(p)
			if err != nil {
				return redisConfig{}, fmt.Errorf("invalid redis db: %w", err)
			}
			db = n
		}
	}

	var username, password string
	if parsed.User != nil {
		username = parsed.User.Username()
		password, _ = parsed.User.Password()
	}

	return redisConfig{
		addr:     net.JoinHostPort(host, port),
		username: username,
		password: password,
		db:       db,
		tls:      parsed.Scheme == "rediss",
	}, nil
}

func dialRedis(cfg redisConfig) (net.Conn, *bufio.Reader, error) {
	dialer := &net.Dialer{Timeout: 5 * time.Second}

	var conn net.Conn
	var err error
	if cfg.tls {
		conn, err = tls.DialWithDialer(dialer, "tcp", cfg.addr, &tls.Config{MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.Dial("tcp", cfg.addr)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to redis at %s: %w", cfg.addr, err)
	}

	reader := bufio.NewReader(conn)

	// AUTH (optional)
	if cfg.password != "" {
		var cmd [][]byte
		if cfg.username != "" {
			cmd = [][]byte{[]byte("AUTH"), []byte(cfg.username), []byte(cfg.password)}
		} else {
			cmd = [][]byte{[]byte("AUTH"), []byte(cfg.password)}
		}
		if err := writeRESPArray(conn, cmd); err != nil {
			_ = conn.Close()
			return nil, nil, err
		}
		if _, err := readRESP(reader); err != nil {
			_ = conn.Close()
			return nil, nil, err
		}
	}

	// SELECT (optional)
	if cfg.db != 0 {
		if err := writeRESPArray(conn, [][]byte{[]byte("SELECT"), []byte(strconv.Itoa(cfg.db))}); err != nil {
			_ = conn.Close()
			return nil, nil, err
		}
		if _, err := readRESP(reader); err != nil {
			_ = conn.Close()
			return nil, nil, err
		}
	}

	// PING
	if err := writeRESPArray(conn, [][]byte{[]byte("PING")}); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	if _, err := readRESP(reader); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	return conn, reader, nil
}

func looksLikeRedisPattern(s string) bool {
	return strings.ContainsAny(s, "*?[]")
}

func writeRESPArray(w io.Writer, parts [][]byte) error {
	if _, err := fmt.Fprintf(w, "*%d\r\n", len(parts)); err != nil {
		return err
	}
	for _, p := range parts {
		if _, err := fmt.Fprintf(w, "$%d\r\n", len(p)); err != nil {
			return err
		}
		if _, err := w.Write(p); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\r\n")); err != nil {
			return err
		}
	}
	return nil
}

func readRESP(r *bufio.Reader) (any, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch prefix {
	case '+': // simple string
		s, err := readLine(r)
		if err != nil {
			return nil, err
		}
		return s, nil
	case '-': // error
		s, err := readLine(r)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("redis error: %s", s)
	case ':': // integer
		s, err := readLine(r)
		if err != nil {
			return nil, err
		}
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	case '$': // bulk string
		s, err := readLine(r)
		if err != nil {
			return nil, err
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		if n == -1 {
			return nil, nil
		}
		buf := make([]byte, n+2) // + CRLF
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return buf[:n], nil
	case '*': // array
		s, err := readLine(r)
		if err != nil {
			return nil, err
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		if n == -1 {
			return nil, nil
		}
		out := make([]any, 0, n)
		for i := 0; i < n; i++ {
			v, err := readRESP(r)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported resp prefix: %q", prefix)
	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line, nil
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toBytes(v any) []byte {
	switch t := v.(type) {
	case []byte:
		return t
	case string:
		return []byte(t)
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}
