package mosquitto

import (
	"fmt"
	"strings"
	"time"
	// Frameworks
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
type MQClient interface {
	// Connect to MQTT broker with hostname, port and options
	Connect(string, uint, ...MQOpt) error

	// Disconnect from MQTT broker
	Disconnect() error

	// Subscribe to topic with wildcard and return request-id
	Subscribe(string, ...MQOpt) (int, error)

	// Unsubscribe from topic with wildcard and return request-id
	Unsubscribe(string) (int, error)

	// Publish []byte data to topic and return request-id
	Publish(string, []byte, ...MQOpt) (int, error)

	// Publish JSON data to topic and return request-id
	PublishJSON(string, interface{}, ...MQOpt) (int, error)

	// Publish measurements in influxdata line protocol format and return request-id
	PublishInflux(string, string, map[string]interface{}, ...MQOpt) (int, error)
}

// Function options
type MQOpt struct {
	Type      Option
	Int       int
	Bool      bool
	Flags     Flags
	String    string
	Timestamp time.Time
}

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

////////////////////////////////////////////////////////////////////////////////
// MQTT Options

func OptQOS(value int) MQOpt           { return MQOpt{Type: MOSQ_OPTION_QOS, Int: value} }
func OptRetain(value bool) MQOpt       { return MQOpt{Type: MOSQ_OPTION_RETAIN, Bool: value} }
func OptFlags(value Flags) MQOpt       { return MQOpt{Type: MOSQ_OPTION_FLAGS, Flags: value} }
func OptKeepaliveSecs(value int) MQOpt { return MQOpt{Type: MOSQ_OPTION_KEEPALIVE, Int: value} }
func OptTag(name, value string) MQOpt {
	return MQOpt{Type: MOSQ_OPTION_TAG, String: fmt.Sprintf("%s=%s", strings.TrimSpace(name), strings.TrimSpace(value))}
}
func OptTimestamp(value time.Time) MQOpt { return MQOpt{Type: MOSQ_OPTION_TIMESTAMP, Timestamp: value} }

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
