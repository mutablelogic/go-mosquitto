package mosquitto

import (
	"context"
	"fmt"
	"strings"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Flags  uint
	Option uint
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Client implements an MQTT client
type Client interface {
	// Connect to MQTT broker with options
	Connect(...Opt) error

	// Disconnect from MQTT broker
	Disconnect() error

	// Subscribe to topic with wildcard and return request-id
	Subscribe(string, ...Opt) (int, error)

	// Unsubscribe from topic with wildcard and return request-id
	Unsubscribe(string) (int, error)

	// Publish []byte data to topic and return request-id
	Publish(string, []byte, ...Opt) (int, error)

	// Publish JSON data to topic and return request-id
	PublishJSON(string, interface{}, ...Opt) (int, error)

	// Publish measurements in influxdata line protocol format and return request-id
	PublishInflux(string, string, map[string]interface{}, ...Opt) (int, error)

	// Wait for a specific request-id or 0 for a connect or disconnect event
	// with context (for timeout)
	WaitFor(context.Context, int) (Event, error)

	// Implements gopi.Unit
	gopi.Unit
}

// Event implements an MQTT event
type Event interface {
	ReturnCode() int // For CONNECT and DISCONNECT

	// Message information
	Id() int
	Type() Flags
	Topic() string
	Data() []byte

	// Implements gopi.Event
	gopi.Event
}

// Function options
type Opt struct {
	Type      Option
	Int       int
	Bool      bool
	Flags     Flags
	String    string
	Timestamp time.Time
}

////////////////////////////////////////////////////////////////////////////////
// MQTT Options

func OptQOS(value int) Opt           { return Opt{Type: MOSQ_OPTION_QOS, Int: value} }
func OptRetain(value bool) Opt       { return Opt{Type: MOSQ_OPTION_RETAIN, Bool: value} }
func OptFlags(value Flags) Opt       { return Opt{Type: MOSQ_OPTION_FLAGS, Flags: value} }
func OptKeepaliveSecs(value int) Opt { return Opt{Type: MOSQ_OPTION_KEEPALIVE, Int: value} }
func OptTag(name, value string) Opt {
	return Opt{Type: MOSQ_OPTION_TAG, String: fmt.Sprintf("%s=%s", strings.TrimSpace(name), strings.TrimSpace(value))}
}
func OptTimestamp(value time.Time) Opt { return Opt{Type: MOSQ_OPTION_TIMESTAMP, Timestamp: value} }

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MOSQ_FLAG_EVENT_CONNECT Flags = 1 << iota
	MOSQ_FLAG_EVENT_DISCONNECT
	MOSQ_FLAG_EVENT_SUBSCRIBE
	MOSQ_FLAG_EVENT_UNSUBSCRIBE
	MOSQ_FLAG_EVENT_PUBLISH
	MOSQ_FLAG_EVENT_MESSAGE
	MOSQ_FLAG_EVENT_LOG
	MOSQ_FLAG_EVENT_NONE Flags = 0
	MOSQ_FLAG_EVENT_ALL        = MOSQ_FLAG_EVENT_CONNECT | MOSQ_FLAG_EVENT_DISCONNECT | MOSQ_FLAG_EVENT_SUBSCRIBE | MOSQ_FLAG_EVENT_UNSUBSCRIBE | MOSQ_FLAG_EVENT_PUBLISH | MOSQ_FLAG_EVENT_MESSAGE
	MOSQ_FLAG_EVENT_MIN        = MOSQ_FLAG_EVENT_CONNECT
	MOSQ_FLAG_EVENT_MAX        = MOSQ_FLAG_EVENT_LOG
)

const (
	MOSQ_OPTION_NONE      Option = iota
	MOSQ_OPTION_QOS              // IntValue
	MOSQ_OPTION_RETAIN           // BoolValue
	MOSQ_OPTION_FLAGS            // FlagsValue
	MOSQ_OPTION_KEEPALIVE        // IntValue
	MOSQ_OPTION_TAG              // StringValue
	MOSQ_OPTION_TIMESTAMP        // TimeValue
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f Flags) String() string {
	if f == MOSQ_FLAG_EVENT_NONE {
		return f.StringFlag()
	}
	str := ""
	for v := MOSQ_FLAG_EVENT_MIN; v <= MOSQ_FLAG_EVENT_MAX; v <<= 1 {
		if f&v == v {
			str += v.StringFlag() + "|"
		}
	}
	return strings.TrimSuffix(str, "|")
}

func (f Flags) StringFlag() string {
	switch f {
	case MOSQ_FLAG_EVENT_NONE:
		return "MOSQ_FLAG_EVENT_NONE"
	case MOSQ_FLAG_EVENT_CONNECT:
		return "MOSQ_FLAG_EVENT_CONNECT"
	case MOSQ_FLAG_EVENT_DISCONNECT:
		return "MOSQ_FLAG_EVENT_DISCONNECT"
	case MOSQ_FLAG_EVENT_SUBSCRIBE:
		return "MOSQ_FLAG_EVENT_SUBSCRIBE"
	case MOSQ_FLAG_EVENT_UNSUBSCRIBE:
		return "MOSQ_FLAG_EVENT_UNSUBSCRIBE"
	case MOSQ_FLAG_EVENT_PUBLISH:
		return "MOSQ_FLAG_EVENT_PUBLISH"
	case MOSQ_FLAG_EVENT_MESSAGE:
		return "MOSQ_FLAG_EVENT_MESSAGE"
	case MOSQ_FLAG_EVENT_LOG:
		return "MOSQ_FLAG_EVENT_LOG"
	default:
		return "[?? Invalid Flags value]"
	}
}

func (o Option) String() string {
	switch o {
	case MOSQ_OPTION_NONE:
		return "MOSQ_OPTION_NONE"
	case MOSQ_OPTION_QOS:
		return "MOSQ_OPTION_QOS"
	case MOSQ_OPTION_RETAIN:
		return "MOSQ_OPTION_RETAIN"
	case MOSQ_OPTION_FLAGS:
		return "MOSQ_OPTION_FLAGS"
	case MOSQ_OPTION_KEEPALIVE:
		return "MOSQ_OPTION_KEEPALIVE"
	case MOSQ_OPTION_TAG:
		return "MOSQ_OPTION_TAG"
	case MOSQ_OPTION_TIMESTAMP:
		return "MOSQ_OPTION_TIMESTAMP"
	default:
		return "[?? Invalid Option value]"
	}
}
