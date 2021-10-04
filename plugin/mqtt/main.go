package main

import (
	"context"
	"fmt"
	"time"

	// Packages
	"github.com/hashicorp/go-multierror"
	"github.com/mutablelogic/go-mosquitto/pkg/mosquitto"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-mosquitto"
	. "github.com/mutablelogic/go-server"
	. "github.com/mutablelogic/go-sqlite"

	// Hack some dependencies
	_ "github.com/djthorpe/go-marshaler"
	_ "gopkg.in/yaml.v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Broker    string        `yaml:"broker"`    // Host:Port or just Host
	ClientId  string        `yaml:"clientid"`  // Client ID (optional)
	Timeout   time.Duration `yaml:"timeout"`   // Connection timeout (optional)
	KeepAlive time.Duration `yaml:"keepalive"` // KeepAlive delta (optional)
	User      string        `yaml:"user"`      // Username (optional)
	Password  string        `yaml:"password"`  // Password (required if user set)
	CertAuth  string        `yaml:"certauth"`  // Certificate Authority path or file (optional)
	CertFile  string        `yaml:"cert"`      // TLS Certificate (required if CertAuth is set)
	KeyFile   string        `yaml:"key"`       // TLS Key (required if CertAuth is set)
	Insecure  bool          `yaml:"insecure"`  // Don't verify broker certificates (optional)
	Topics    []string      `yaml:"topics"`    // Topics to subscribe to (optional)
	Database  string        `yaml:"database"`  // Database name for storage of messages
	Retain    time.Duration `yaml:"retention"` // Retain time for messages (optional)
}

type plugin struct {
	pool
	cfg       Config
	client    *mosquitto.Client
	ch        chan *mosquitto.Event
	connected time.Time
	topics    *topics
}

type pool interface {
	Get() SQConnection
	Put(conn SQConnection)
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultConnectTimeout = 30 * time.Second
	defaultKeepAlive      = 60 * time.Second
	defaultRetain         = 7 * 24 * time.Hour
	defaultCapacity       = 10000
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Load configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, err)
		return nil
	} else if cfg.Broker == "" {
		provider.Print(ctx, "Missing required 'broker' configuration")
		return nil
	} else {
		p.cfg = cfg
	}

	// Get sqlite database
	if pool, ok := provider.GetPlugin(ctx, "sqlite3").(pool); !ok {
		provider.Print(ctx, "Missing required 'sqlite3' plugin")
		return nil
	} else if p.cfg.Database == "" {
		provider.Print(ctx, "Missing required 'database' configuration")
		return nil
	} else if err := HasSchema(pool, p.cfg.Database); err != nil {
		provider.Print(ctx, err)
		return nil
	} else {
		p.pool = pool
	}

	// Set message retain (with minimum of one minute)
	if p.cfg.Retain < time.Minute {
		p.cfg.Retain = defaultRetain
	}

	// Create a channel to receive events
	p.ch = make(chan *mosquitto.Event, defaultCapacity)

	// Create a topics object to track subscriptions
	p.topics = NewTopics()

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<mqtt"
	if p.connected.IsZero() {
		str += " disconnected"
	} else {
		str += fmt.Sprint(" connected=", time.Since(p.connected))
	}
	if p.client != nil {
		str += fmt.Sprint(" ", p.client)
	}
	if p.pool != nil {
		str += fmt.Sprint(" ", p.pool)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "mqtt"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	var result error

	// Set up a ticker to check for connection and topic subscription
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	// Set up a ticker to delete older messages beyond the retain time
	retain := time.NewTimer(5 * time.Second)
	defer retain.Stop()

	// Set up schema
	if err := p.AddSchema(ctx); err != nil {
		return err
	}

	// Add REST API handlers
	if err := p.AddHandlers(ctx, provider); err != nil {
		return err
	}

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case <-timer.C:
			// Re-connect client as necessary
			if p.client == nil {
				provider.Printf(ctx, "Connect: %q", p.cfg.Broker)
				if client, err := p.connect(ctx, provider); err != nil {
					provider.Printf(ctx, "Connection error: %v", err)
				} else {
					p.client = client
					p.connected = time.Now()
				}
			} else {
				if err := p.subscribeToTopics(); err != nil {
					provider.Printf(ctx, "Subscribe error: %v", err)
				}
			}
			// Reset the timer
			timer.Reset(defaultConnectTimeout)
		case <-retain.C:
			if n, err := p.RetainCycle(ctx); err != nil {
				provider.Printf(ctx, "Retain cycle error: %v", err)
			} else if n > 0 {
				provider.Printf(ctx, "Retain cycle deleted %d oldest messages", n)
			}
			// Reset the timer
			retain.Reset(p.cfg.Retain / 4)
		case evt := <-p.ch:
			// Handle message
			if evt.Type == MOSQ_FLAG_EVENT_MESSAGE {
				if err := p.AddMessage(ctx, evt); err != nil {
					provider.Printf(ctx, "Message error: %v", err)
				}
			} else if evt.Type == MOSQ_FLAG_EVENT_SUBSCRIBE || evt.Type == MOSQ_FLAG_EVENT_UNSUBSCRIBE {
				p.topics.Event(evt.Type, evt.Id)
			} else {
				provider.Printf(ctx, "Event: %v", evt)
			}
		}
	}

	// Disconnect client if connected
	if p.client != nil {
		if err := p.client.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Close the event channel
	close(p.ch)

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Subscribe to a topic
func (p *plugin) Subscribe(topic string) error {
	if p.client == nil {
		return ErrOutOfOrder.With("Client not connected")
	}
	if req, err := p.client.Subscribe(topic); err != nil {
		return err
	} else {
		p.topics.Subscribe(topic, req)
	}

	// Return success
	return nil
}

// Unsubscribe from a topic
func (p *plugin) Unubscribe(topic string) error {
	if p.client == nil {
		return ErrOutOfOrder.With("Client not connected")
	}
	if req, err := p.client.Unsubscribe(topic); err != nil {
		return err
	} else {
		// TODO: Remove topic from p.cfg.Topics
		p.topics.Unsubscribe(topic, req)
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *plugin) connect(ctx context.Context, provider Provider) (*mosquitto.Client, error) {
	// Create config
	cfg := mosquitto.NewConfigWithBroker(p.cfg.Broker)
	if p.cfg.ClientId != "" {
		cfg = cfg.WithClientId(p.cfg.ClientId)
	}
	cfg = cfg.WithCallback(func(evt *mosquitto.Event) {
		p.callback(ctx, provider, evt)
	})
	if p.cfg.KeepAlive > 0 {
		cfg = cfg.WithKeepalive(p.cfg.KeepAlive)
	} else {
		cfg = cfg.WithKeepalive(defaultKeepAlive)
	}
	if p.cfg.CertAuth != "" {
		cfg = cfg.WithTLS(p.cfg.CertAuth, p.cfg.CertFile, p.cfg.KeyFile, !p.cfg.Insecure)
	}
	if p.cfg.User != "" {
		cfg = cfg.WithCredentials(p.cfg.User, p.cfg.Password)
	}

	// Create a context for the timeout
	connectTimeout := defaultConnectTimeout
	if p.cfg.Timeout > 0 {
		connectTimeout = p.cfg.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	// Connect and return any errors
	return mosquitto.NewWithConfig(ctx, cfg)
}

func (p *plugin) callback(ctx context.Context, provider Provider, evt *mosquitto.Event) {
	switch evt.Type {
	case MOSQ_FLAG_EVENT_CONNECT:
		provider.Print(ctx, evt)
		if evt.Err != nil {
			provider.Printf(ctx, "Connection error: %v", evt.Err)
			return
		}
		// Subscribe to topics
		if err := p.subscribeToTopics(); err != nil {
			provider.Printf(ctx, "Subscribe error: %v", err)
		}
	case MOSQ_FLAG_EVENT_DISCONNECT:
		provider.Print(ctx, evt)
		p.client = nil
		p.connected = time.Time{}
		if evt.Err != nil {
			provider.Printf(ctx, "Disconnection error: %v", evt.Err)
			return
		}
	default:
		select {
		case p.ch <- evt:
			break
		default:
			provider.Printf(ctx, "Message dropped in topic %q, too many messages", evt.Topic)
		}
	}
}

func (p *plugin) subscribeToTopics() error {
	for _, topic := range p.cfg.Topics {
		if !p.topics.Has(topic) {
			if err := p.Subscribe(topic); err != nil {
				return err
			}
		}
	}
	// Return success
	return nil
}
