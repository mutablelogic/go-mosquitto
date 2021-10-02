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
	Error int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MOSQ_ERR_CONN_PENDING   = C.MOSQ_ERR_CONN_PENDING
	MOSQ_ERR_SUCCESS        = C.MOSQ_ERR_SUCCESS
	MOSQ_ERR_NOMEM          = C.MOSQ_ERR_NOMEM
	MOSQ_ERR_PROTOCOL       = C.MOSQ_ERR_PROTOCOL
	MOSQ_ERR_INVAL          = C.MOSQ_ERR_INVAL
	MOSQ_ERR_NO_CONN        = C.MOSQ_ERR_NO_CONN
	MOSQ_ERR_CONN_REFUSED   = C.MOSQ_ERR_CONN_REFUSED
	MOSQ_ERR_NOT_FOUND      = C.MOSQ_ERR_NOT_FOUND
	MOSQ_ERR_CONN_LOST      = C.MOSQ_ERR_CONN_LOST
	MOSQ_ERR_TLS            = C.MOSQ_ERR_TLS
	MOSQ_ERR_PAYLOAD_SIZE   = C.MOSQ_ERR_PAYLOAD_SIZE
	MOSQ_ERR_NOT_SUPPORTED  = C.MOSQ_ERR_NOT_SUPPORTED
	MOSQ_ERR_AUTH           = C.MOSQ_ERR_AUTH
	MOSQ_ERR_ACL_DENIED     = C.MOSQ_ERR_ACL_DENIED
	MOSQ_ERR_UNKNOWN        = C.MOSQ_ERR_UNKNOWN
	MOSQ_ERR_ERRNO          = C.MOSQ_ERR_ERRNO
	MOSQ_ERR_EAI            = C.MOSQ_ERR_EAI
	MOSQ_ERR_PROXY          = C.MOSQ_ERR_PROXY
	MOSQ_ERR_PLUGIN_DEFER   = C.MOSQ_ERR_PLUGIN_DEFER
	MOSQ_ERR_MALFORMED_UTF8 = C.MOSQ_ERR_MALFORMED_UTF8
	MOSQ_ERR_KEEPALIVE      = C.MOSQ_ERR_KEEPALIVE
	MOSQ_ERR_LOOKUP         = C.MOSQ_ERR_LOOKUP
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Error) Error() string {
	return C.GoString(C.mosquitto_strerror(C.int(e)))
}
