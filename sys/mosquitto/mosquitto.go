package mosquitto

import (
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libmosquitto
#include <stdlib.h>
#include <mosquitto.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Level int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MOSQ_DEFAULT_PORT        = 1883
	MOSQ_DEFAULT_SECURE_PORT = 8883
)

const (
	MOSQ_LOG_NONE        Level = 0
	MOSQ_LOG_INFO        Level = (1 << 0)
	MOSQ_LOG_NOTICE      Level = (1 << 1)
	MOSQ_LOG_WARNING     Level = (1 << 2)
	MOSQ_LOG_ERR         Level = (1 << 3)
	MOSQ_LOG_DEBUG       Level = (1 << 4)
	MOSQ_LOG_SUBSCRIBE   Level = (1 << 5)
	MOSQ_LOG_UNSUBSCRIBE Level = (1 << 6)
	MOSQ_LOG_WEBSOCKETS  Level = (1 << 7)
	MOSQ_LOG_MIN               = MOSQ_LOG_INFO
	MOSQ_LOG_MAX               = MOSQ_LOG_WEBSOCKETS
)

////////////////////////////////////////////////////////////////////////////////
// INIT & CLEANUP

// Version returns major, minor and revision of the mosquitto
// client library
func Version() (int, int, int) {
	var major, minor, revision C.int
	C.mosquitto_lib_version(&major, &minor, &revision)
	return int(major), int(minor), int(revision)
}

// Init initiaizes the library
func Init() error {
	if err := Error(C.mosquitto_lib_init()); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

// Cleanup should be called when use of library is finished
func Cleanup() error {
	if err := Error(C.mosquitto_lib_cleanup()); err != MOSQ_ERR_SUCCESS {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f Level) StringFlag() string {
	switch f {
	case MOSQ_LOG_NONE:
		return "MOSQ_LOG_NONE"
	case MOSQ_LOG_INFO:
		return "MOSQ_LOG_INFO"
	case MOSQ_LOG_NOTICE:
		return "MOSQ_LOG_NOTICE"
	case MOSQ_LOG_WARNING:
		return "MOSQ_LOG_WARNING"
	case MOSQ_LOG_ERR:
		return "MOSQ_LOG_ERR"
	case MOSQ_LOG_DEBUG:
		return "MOSQ_LOG_DEBUG"
	case MOSQ_LOG_SUBSCRIBE:
		return "MOSQ_LOG_SUBSCRIBE"
	case MOSQ_LOG_UNSUBSCRIBE:
		return "MOSQ_LOG_UNSUBSCRIBE"
	case MOSQ_LOG_WEBSOCKETS:
		return "MOSQ_LOG_WEBSOCKETS"
	default:
		return "[?? Invalid Level value]"
	}
}

func (f Level) String() string {
	if f == MOSQ_LOG_NONE {
		return f.StringFlag()
	}
	str := ""
	for v := MOSQ_LOG_MIN; v <= MOSQ_LOG_MAX; v <<= 1 {
		if f&v == v {
			str += v.StringFlag() + "|"
		}
	}
	return strings.TrimSuffix(str, "|")
}
