package providers

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
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

// MQTTProvider implements MessagingProvider for MQTT 3.1.1.
//
// Scope (intentionally minimal):
// - QoS 0 only (publish/subscribe)
// - Clean session
// - Username/password via URL
// - mqtt:// and mqtts://
type MQTTProvider struct {
	mu        sync.RWMutex
	connected bool

	conn   net.Conn
	reader *bufio.Reader
	writeM sync.Mutex

	stop chan struct{}

	nextPacketID uint16

	handlers map[string]messaging.MessageHandler
	ackWait  map[uint16]chan byte
}

func NewMQTTProvider() *MQTTProvider {
	return &MQTTProvider{
		handlers: make(map[string]messaging.MessageHandler),
		ackWait:  make(map[uint16]chan byte),
	}
}

func (m *MQTTProvider) Connect(rawURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return nil
	}

	cfg, err := parseMQTTConfig(rawURL)
	if err != nil {
		return err
	}

	conn, reader, err := dialMQTT(cfg)
	if err != nil {
		return err
	}

	// CONNECT
	clientID := cfg.clientID
	if clientID == "" {
		clientID = "broker-ui-" + randomHex(6)
	}
	if err := m.writePacket(conn, buildCONNECT(clientID, cfg.username, cfg.password, cfg.keepAliveSeconds)); err != nil {
		_ = conn.Close()
		return err
	}

	pktType, payload, err := readMQTTPacket(reader)
	if err != nil {
		_ = conn.Close()
		return err
	}
	if pktType != 0x20 { // CONNACK
		_ = conn.Close()
		return fmt.Errorf("unexpected mqtt packet during connect: 0x%02x", pktType)
	}
	if len(payload) < 2 {
		_ = conn.Close()
		return fmt.Errorf("invalid CONNACK payload")
	}
	rc := payload[1]
	if rc != 0 {
		_ = conn.Close()
		return fmt.Errorf("mqtt connack refused: rc=%d", rc)
	}

	m.conn = conn
	m.reader = reader
	m.connected = true
	m.stop = make(chan struct{})
	m.nextPacketID = 1

	go m.readLoop(m.stop)
	go m.keepAliveLoop(m.stop, time.Duration(cfg.keepAliveSeconds)*time.Second)
	return nil
}

func (m *MQTTProvider) Publish(subject string, data []byte) error {
	m.mu.RLock()
	conn := m.conn
	m.mu.RUnlock()
	if conn == nil || !m.IsConnected() {
		return fmt.Errorf("not connected to MQTT")
	}

	pkt := buildPUBLISH(subject, data)
	return m.writePacket(conn, pkt)
}

func (m *MQTTProvider) Subscribe(subjectPattern string, handler messaging.MessageHandler) error {
	m.mu.Lock()
	if !m.connected || m.conn == nil {
		m.mu.Unlock()
		return fmt.Errorf("not connected to MQTT")
	}
	if _, exists := m.handlers[subjectPattern]; exists {
		m.mu.Unlock()
		return fmt.Errorf("already subscribed to: %s", subjectPattern)
	}

	packetID := m.nextIDLocked()
	ch := make(chan byte, 1)
	m.ackWait[packetID] = ch
	m.handlers[subjectPattern] = handler
	conn := m.conn
	m.mu.Unlock()

	pkt := buildSUBSCRIBE(packetID, subjectPattern)
	if err := m.writePacket(conn, pkt); err != nil {
		m.cleanupAck(packetID)
		m.mu.Lock()
		delete(m.handlers, subjectPattern)
		m.mu.Unlock()
		return err
	}

	select {
	case code := <-ch:
		m.cleanupAck(packetID)
		if code == 0x80 {
			m.mu.Lock()
			delete(m.handlers, subjectPattern)
			m.mu.Unlock()
			return fmt.Errorf("mqtt subscribe refused for %s", subjectPattern)
		}
		return nil
	case <-time.After(5 * time.Second):
		m.cleanupAck(packetID)
		m.mu.Lock()
		delete(m.handlers, subjectPattern)
		m.mu.Unlock()
		return fmt.Errorf("mqtt subscribe timeout for %s", subjectPattern)
	}
}

func (m *MQTTProvider) Unsubscribe(subjectPattern string) error {
	m.mu.Lock()
	if !m.connected || m.conn == nil {
		m.mu.Unlock()
		return fmt.Errorf("not connected to MQTT")
	}
	if _, exists := m.handlers[subjectPattern]; !exists {
		m.mu.Unlock()
		return fmt.Errorf("no subscription found for: %s", subjectPattern)
	}

	packetID := m.nextIDLocked()
	ch := make(chan byte, 1)
	m.ackWait[packetID] = ch
	conn := m.conn
	m.mu.Unlock()

	pkt := buildUNSUBSCRIBE(packetID, subjectPattern)
	if err := m.writePacket(conn, pkt); err != nil {
		m.cleanupAck(packetID)
		return err
	}

	select {
	case <-ch:
		m.cleanupAck(packetID)
		m.mu.Lock()
		delete(m.handlers, subjectPattern)
		m.mu.Unlock()
		return nil
	case <-time.After(5 * time.Second):
		m.cleanupAck(packetID)
		return fmt.Errorf("mqtt unsubscribe timeout for %s", subjectPattern)
	}
}

func (m *MQTTProvider) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil
	}
	m.connected = false
	if m.stop != nil {
		close(m.stop)
		m.stop = nil
	}
	if m.conn != nil {
		_ = m.writePacket(m.conn, []byte{0xE0, 0x00}) // DISCONNECT
		_ = m.conn.Close()
		m.conn = nil
	}
	m.reader = nil
	m.handlers = make(map[string]messaging.MessageHandler)
	for id, ch := range m.ackWait {
		_ = ch
		delete(m.ackWait, id)
	}
	return nil
}

func (m *MQTTProvider) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected && m.conn != nil
}

func (m *MQTTProvider) GetProviderType() messaging.ProviderType {
	return messaging.ProviderMQTT
}

func (m *MQTTProvider) nextIDLocked() uint16 {
	id := m.nextPacketID
	m.nextPacketID++
	if m.nextPacketID == 0 {
		m.nextPacketID = 1
	}
	return id
}

func (m *MQTTProvider) cleanupAck(packetID uint16) {
	m.mu.Lock()
	delete(m.ackWait, packetID)
	m.mu.Unlock()
}

func (m *MQTTProvider) readLoop(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		m.mu.RLock()
		reader := m.reader
		m.mu.RUnlock()
		if reader == nil {
			return
		}

		pktType, payload, err := readMQTTPacket(reader)
		if err != nil {
			return
		}

		switch pktType & 0xF0 {
		case 0x30: // PUBLISH (QoS 0/1/2)
			topic, msg, err := parsePUBLISH(pktType, payload)
			if err != nil {
				continue
			}
			m.dispatch(topic, msg)

		case 0x90: // SUBACK
			if len(payload) < 3 {
				continue
			}
			id := binary.BigEndian.Uint16(payload[0:2])
			code := payload[2]
			m.signalAck(id, code)

		case 0xB0: // UNSUBACK
			if len(payload) < 2 {
				continue
			}
			id := binary.BigEndian.Uint16(payload[0:2])
			m.signalAck(id, 0)

		case 0xD0: // PINGRESP
			// ignore
		default:
			// ignore
		}
	}
}

func (m *MQTTProvider) keepAliveLoop(stop <-chan struct{}, keepAlive time.Duration) {
	if keepAlive <= 0 {
		return
	}
	interval := keepAlive / 2
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if !m.IsConnected() {
				return
			}
			m.mu.RLock()
			conn := m.conn
			m.mu.RUnlock()
			if conn == nil {
				return
			}
			_ = m.writePacket(conn, []byte{0xC0, 0x00}) // PINGREQ
		}
	}
}

func (m *MQTTProvider) signalAck(packetID uint16, code byte) {
	m.mu.RLock()
	ch := m.ackWait[packetID]
	m.mu.RUnlock()
	if ch == nil {
		return
	}
	select {
	case ch <- code:
	default:
	}
}

func (m *MQTTProvider) dispatch(topic string, payload []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for filter, handler := range m.handlers {
		if handler == nil {
			continue
		}
		if mqttMatch(filter, topic) {
			handler(topic, payload)
		}
	}
}

type mqttConfig struct {
	addr             string
	tls              bool
	username         string
	password         string
	clientID         string
	keepAliveSeconds uint16
}

func parseMQTTConfig(rawURL string) (mqttConfig, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return mqttConfig{}, fmt.Errorf("mqtt url is empty")
	}

	u := rawURL
	if !strings.Contains(u, "://") {
		u = "mqtt://" + u
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return mqttConfig{}, fmt.Errorf("invalid mqtt url: %w", err)
	}
	switch parsed.Scheme {
	case "mqtt", "mqtts":
	default:
		return mqttConfig{}, fmt.Errorf("unsupported mqtt scheme: %s", parsed.Scheme)
	}
	host := parsed.Hostname()
	if host == "" {
		return mqttConfig{}, fmt.Errorf("mqtt host is empty")
	}
	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "mqtts" {
			port = "8883"
		} else {
			port = "1883"
		}
	}

	var username, password string
	if parsed.User != nil {
		username = parsed.User.Username()
		password, _ = parsed.User.Password()
	}

	keepAlive := uint16(30)
	if v := parsed.Query().Get("keepalive"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n >= 0 && n <= 65535 {
			keepAlive = uint16(n)
		}
	}
	clientID := parsed.Query().Get("client_id")

	return mqttConfig{
		addr:             net.JoinHostPort(host, port),
		tls:              parsed.Scheme == "mqtts",
		username:         username,
		password:         password,
		clientID:         clientID,
		keepAliveSeconds: keepAlive,
	}, nil
}

func dialMQTT(cfg mqttConfig) (net.Conn, *bufio.Reader, error) {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	var conn net.Conn
	var err error
	if cfg.tls {
		conn, err = tls.DialWithDialer(dialer, "tcp", cfg.addr, &tls.Config{MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.Dial("tcp", cfg.addr)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to mqtt at %s: %w", cfg.addr, err)
	}
	return conn, bufio.NewReader(conn), nil
}

func (m *MQTTProvider) writePacket(w io.Writer, packet []byte) error {
	m.writeM.Lock()
	defer m.writeM.Unlock()
	_, err := w.Write(packet)
	return err
}

func buildCONNECT(clientID, username, password string, keepAliveSeconds uint16) []byte {
	var vh []byte
	// Protocol name length + "MQTT" + protocol level (4) + connect flags + keepalive
	vh = append(vh, 0x00, 0x04)
	vh = append(vh, 'M', 'Q', 'T', 'T')
	vh = append(vh, 0x04) // MQTT 3.1.1

	flags := byte(0)
	flags |= 1 << 1 // Clean Session
	if username != "" {
		flags |= 1 << 7
	}
	if password != "" {
		flags |= 1 << 6
	}
	vh = append(vh, flags)
	ka := make([]byte, 2)
	binary.BigEndian.PutUint16(ka, keepAliveSeconds)
	vh = append(vh, ka...)

	var pl []byte
	pl = append(pl, encodeUTF8(clientID)...)
	if username != "" {
		pl = append(pl, encodeUTF8(username)...)
	}
	if password != "" {
		pl = append(pl, encodeUTF8(password)...)
	}

	return buildPacket(0x10, append(vh, pl...))
}

func buildPUBLISH(topic string, payload []byte) []byte {
	var vh []byte
	vh = append(vh, encodeUTF8(topic)...)
	body := append(vh, payload...)
	return buildPacket(0x30, body) // QoS0, no retain, no dup
}

func buildSUBSCRIBE(packetID uint16, topicFilter string) []byte {
	var body []byte
	pid := make([]byte, 2)
	binary.BigEndian.PutUint16(pid, packetID)
	body = append(body, pid...)
	body = append(body, encodeUTF8(topicFilter)...)
	body = append(body, 0x00)      // QoS0
	return buildPacket(0x82, body) // SUBSCRIBE requires flags 0010
}

func buildUNSUBSCRIBE(packetID uint16, topicFilter string) []byte {
	var body []byte
	pid := make([]byte, 2)
	binary.BigEndian.PutUint16(pid, packetID)
	body = append(body, pid...)
	body = append(body, encodeUTF8(topicFilter)...)
	return buildPacket(0xA2, body) // UNSUBSCRIBE requires flags 0010
}

func buildPacket(firstByte byte, body []byte) []byte {
	rem := encodeRemainingLength(len(body))
	out := make([]byte, 0, 1+len(rem)+len(body))
	out = append(out, firstByte)
	out = append(out, rem...)
	out = append(out, body...)
	return out
}

func encodeUTF8(s string) []byte {
	b := []byte(s)
	out := make([]byte, 2+len(b))
	binary.BigEndian.PutUint16(out[0:2], uint16(len(b)))
	copy(out[2:], b)
	return out
}

func encodeRemainingLength(n int) []byte {
	var out []byte
	for {
		digit := byte(n % 128)
		n /= 128
		if n > 0 {
			digit |= 0x80
		}
		out = append(out, digit)
		if n == 0 {
			break
		}
	}
	return out
}

func readMQTTPacket(r *bufio.Reader) (byte, []byte, error) {
	first, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	rem, err := decodeRemainingLength(r)
	if err != nil {
		return 0, nil, err
	}
	payload := make([]byte, rem)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}
	return first, payload, nil
}

func decodeRemainingLength(r *bufio.Reader) (int, error) {
	mult := 1
	value := 0
	for i := 0; i < 4; i++ {
		digit, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		value += int(digit&127) * mult
		mult *= 128
		if digit&128 == 0 {
			return value, nil
		}
	}
	return 0, fmt.Errorf("invalid remaining length")
}

func parsePUBLISH(firstByte byte, payload []byte) (string, []byte, error) {
	if len(payload) < 2 {
		return "", nil, fmt.Errorf("invalid publish")
	}
	topicLen := int(binary.BigEndian.Uint16(payload[0:2]))
	if len(payload) < 2+topicLen {
		return "", nil, fmt.Errorf("invalid publish topic")
	}
	topic := string(payload[2 : 2+topicLen])
	idx := 2 + topicLen

	qos := (firstByte >> 1) & 0x03
	if qos > 0 {
		// Skip packet identifier for QoS1/2
		if len(payload) < idx+2 {
			return "", nil, fmt.Errorf("invalid publish packet id")
		}
		idx += 2
	}
	return topic, payload[idx:], nil
}

func mqttMatch(filter, topic string) bool {
	// Exact match fast path
	if filter == topic {
		return true
	}

	f := strings.Split(filter, "/")
	t := strings.Split(topic, "/")

	for i := 0; i < len(f); i++ {
		if i >= len(t) {
			return false
		}
		switch f[i] {
		case "#":
			return true
		case "+":
			continue
		default:
			if f[i] != t[i] {
				return false
			}
		}
	}
	return len(f) == len(t)
}

func randomHex(nBytes int) string {
	if nBytes <= 0 {
		nBytes = 6
	}
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b)
}
