package mosquitto

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libmosquitto
#include <mosquitto.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Option C.enum_mosq_opt_t
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MOSQ_OPT_PROTOCOL_VERSION      Option = 1
	MOSQ_OPT_SSL_CTX               Option = 2
	MOSQ_OPT_SSL_CTX_WITH_DEFAULTS Option = 3
	MOSQ_OPT_RECEIVE_MAXIMUM       Option = 4
	MOSQ_OPT_SEND_MAXIMUM          Option = 5
	MOSQ_OPT_TLS_KEYFORM           Option = 6
	MOSQ_OPT_TLS_ENGINE            Option = 7
	MOSQ_OPT_TLS_ENGINE_KPASS_SHA1 Option = 8
	MOSQ_OPT_TLS_OCSP_REQUIRED     Option = 9
	MOSQ_OPT_TLS_ALPN              Option = 10
)

const (
	MQTT_PROTOCOL_V31  = int(C.MQTT_PROTOCOL_V31)
	MQTT_PROTOCOL_V311 = int(C.MQTT_PROTOCOL_V311)
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v Option) String() string {
	switch v {
	case MOSQ_OPT_PROTOCOL_VERSION:
		return "MOSQ_OPT_PROTOCOL_VERSION"
	case MOSQ_OPT_SSL_CTX:
		return "MOSQ_OPT_SSL_CTX"
	case MOSQ_OPT_SSL_CTX_WITH_DEFAULTS:
		return "MOSQ_OPT_SSL_CTX_WITH_DEFAULTS"
	case MOSQ_OPT_RECEIVE_MAXIMUM:
		return "MOSQ_OPT_RECEIVE_MAXIMUM"
	case MOSQ_OPT_SEND_MAXIMUM:
		return "MOSQ_OPT_SEND_MAXIMUM"
	case MOSQ_OPT_TLS_KEYFORM:
		return "MOSQ_OPT_TLS_KEYFORM"
	case MOSQ_OPT_TLS_ENGINE:
		return "MOSQ_OPT_TLS_ENGINE"
	case MOSQ_OPT_TLS_ENGINE_KPASS_SHA1:
		return "MOSQ_OPT_TLS_ENGINE_KPASS_SHA1"
	case MOSQ_OPT_TLS_OCSP_REQUIRED:
		return "MOSQ_OPT_TLS_OCSP_REQUIRED"
	case MOSQ_OPT_TLS_ALPN:
		return "MOSQ_OPT_TLS_ALPN"
	default:
		return "[?? Invalid Option value]"
	}
}
