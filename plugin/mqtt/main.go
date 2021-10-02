package main

import (
	"context"
	"time"

	// Packages
	"github.com/djthorpe/go-mosquitto/pkg/mosquitto"
	"github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-mosquitto"
	. "github.com/mutablelogic/go-server"

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
	Topics    []string      `yaml:"topics"`    // Topics to subscribe to
}

type plugin struct {
	cfg    Config
	client *mosquitto.Client
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultConnectTimeout = 30 * time.Second
	defaultKeepAlive      = 60 * time.Second
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
		provider.Print(ctx, "Missing required 'broker'")
		return nil
	} else {
		p.cfg = cfg
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<mqtt"
	if p.client != nil {
		str += " " + p.client.String()
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

	// Set up a ticker to check for connection
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

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
				}
			}
			// Reset the timer
			timer.Reset(defaultConnectTimeout)
		}
	}

	// Disconnect client if connected
	if p.client != nil {
		if err := p.client.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Subscribe(topic string) error {
	if p.client == nil {
		return ErrOutOfOrder.With("Client not connected")
	}
	if _, err := p.client.Subscribe(topic); err != nil {
		return err
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
	provider.Print(ctx, evt)
	switch evt.Type {
	case MOSQ_FLAG_EVENT_CONNECT:
		if evt.Err != nil {
			provider.Printf(ctx, "Connection error: %v", evt.Err)
			return
		}
		// Subscribe to topics
		for _, topic := range p.cfg.Topics {
			if err := p.Subscribe(topic); err != nil {
				provider.Printf(ctx, "Subscribe error: %q: %v", topic, err)
			}
		}
	case MOSQ_FLAG_EVENT_DISCONNECT:
		p.client = nil
		if evt.Err != nil {
			provider.Printf(ctx, "Disconnection error: %v", evt.Err)
			return
		}
	case MOSQ_FLAG_EVENT_MESSAGE:
		provider.Printf(ctx, "Message: %q => %v", evt.Topic, string(evt.Data))
	default:
		provider.Print(ctx, evt)
	}
}
