package mosquitto

import (
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Flags uint
)

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
