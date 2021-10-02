package mosquitto

import (
	"net"
	"strconv"
	"time"
	// Packages
	// Namespace imports
	//. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	// Broker
	clientId string
	host     string
	port     uint

	// Timeouts
	keepalive time.Duration

	// Credentials
	user     string
	password string

	// TLS options
	capath     string
	certpath   string
	keypath    string
	certverify bool

	// Callbacks
	fn    EventFunc
	trace TraceFunc
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	defaultConfig = Config{
		keepalive: 60 * time.Second,
	}
)

////////////////////////////////////////////////////////////////////////////////
// CONFIGURATION OPTIONS

// Create a new empty configuration
func NewConfigWithBroker(host string) Config {
	return defaultConfig.WithHost(host)
}

func (c Config) WithClientId(v string) Config {
	c.clientId = v
	return c
}

func (c Config) WithCredentials(user, password string) Config {
	c.user = user
	c.password = password
	return c
}

func (c Config) WithTLS(capath, certpath, keypath string, verify bool) Config {
	c.capath = capath
	c.certpath = certpath
	c.keypath = keypath
	c.certverify = verify
	return c
}

func (c Config) WithHost(v string) Config {
	// Try host:port version first
	if host, port, err := net.SplitHostPort(v); err == nil {
		if port, err := strconv.ParseUint(port, 0, 16); err == nil {
			c.host = host
			c.port = uint(port)
			return c
		}
	}

	// fallback to interpreting as host only
	c.host = v
	c.port = 0

	// return config
	return c
}

func (c Config) WithKeepalive(d time.Duration) Config {
	c.keepalive = d
	return c
}

func (c Config) WithCallback(fn EventFunc) Config {
	c.fn = fn
	return c
}

func (c Config) WithTrace(fn TraceFunc) Config {
	c.trace = fn
	return c
}
