package tracing

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// NATS Logging.NATS implements the ILogger interface
type NATS struct {
	conn     *nats.Conn
	subject  string
	clientID string
}

// NATSOption is a functional option for configuring the NATS logger
type NATSOption func(*NATS)

// WithClientID sets the client ID for the NATS logger
func WithClientID(clientID string) NATSOption {
	return func(n *NATS) {
		n.clientID = clientID
	}
}

// WithSubject sets the subject for publishing log messages
func WithSubject(subject string) NATSOption {
	return func(n *NATS) {
		n.subject = subject
	}
}

// WithCredentials sets username and password for NATS authentication
func WithCredentials(username, password string) NATSOption {
	return func(n *NATS) {
		// This option doesn't modify the NATS struct directly
		// Instead, it's used when establishing the connection
	}
}

func NewMetricsNATS(url string, options ...NATSOption) (*NATS, error) {
	logger := &NATS{
		subject:  "tracing",
		clientID: "internal-tracing-broker",
	}

	for _, opt := range options {
		opt(logger)
	}

	natsOpts := []nats.Option{
		nats.Name(logger.clientID),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10),
	}

	nc, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	logger.conn = nc
	return logger, nil
}

func NewMetricsNATSWithAuth(url string, username, password string, options ...NATSOption) (*NATS, error) {
	logger := &NATS{
		subject:  "tracing",
		clientID: "internal-tracing-broker",
	}

	for _, opt := range options {
		opt(logger)
	}

	nc, err := nats.Connect(url,
		nats.Name(logger.clientID),
		nats.UserInfo(username, password),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(10),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	logger.conn = nc
	return logger, nil
}

func (n *NATS) Log(mm Trace) error {
	if n.conn == nil || n.conn.IsClosed() {
		return fmt.Errorf("NATS connection is closed or not initialized")
	}

	jsonBytes, err := json.Marshal(mm)
	if err != nil {
		return err
	}

	err = n.conn.Publish(n.subject, jsonBytes)
	if err != nil {
		return err
	}

	err = n.conn.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (n *NATS) Close() {
	if n.conn != nil && !n.conn.IsClosed() {
		n.conn.Close()
	}
}
