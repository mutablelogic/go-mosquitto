package mosquitto

import (
	"errors"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libmosquitto
#include <stdlib.h>
#include <mosquitto.h>
#include <mosquitto_broker.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Error  int
	Client C.struct_mosquitto
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MOSQ_ERR_AUTH_CONTINUE      Error = -4
	MOSQ_ERR_NO_SUBSCRIBERS     Error = -3
	MOSQ_ERR_SUB_EXISTS         Error = -2
	MOSQ_ERR_CONN_PENDING       Error = -1
	MOSQ_ERR_SUCCESS            Error = 0
	MOSQ_ERR_NOMEM              Error = 1
	MOSQ_ERR_PROTOCOL           Error = 2
	MOSQ_ERR_INVAL              Error = 3
	MOSQ_ERR_NO_CONN            Error = 4
	MOSQ_ERR_CONN_REFUSED       Error = 5
	MOSQ_ERR_NOT_FOUND          Error = 6
	MOSQ_ERR_CONN_LOST          Error = 7
	MOSQ_ERR_TLS                Error = 8
	MOSQ_ERR_PAYLOAD_SIZE       Error = 9
	MOSQ_ERR_NOT_SUPPORTED      Error = 10
	MOSQ_ERR_AUTH               Error = 11
	MOSQ_ERR_ACL_DENIED         Error = 12
	MOSQ_ERR_UNKNOWN            Error = 13
	MOSQ_ERR_ERRNO              Error = 14
	MOSQ_ERR_EAI                Error = 15
	MOSQ_ERR_PROXY              Error = 16
	MOSQ_ERR_PLUGIN_DEFER       Error = 17
	MOSQ_ERR_MALFORMED_UTF8     Error = 18
	MOSQ_ERR_KEEPALIVE          Error = 19
	MOSQ_ERR_LOOKUP             Error = 20
	MOSQ_ERR_MALFORMED_PACKET   Error = 21
	MOSQ_ERR_DUPLICATE_PROPERTY Error = 22
	MOSQ_ERR_TLS_HANDSHAKE      Error = 23
	MOSQ_ERR_QOS_NOT_SUPPORTED  Error = 24
	MOSQ_ERR_OVERSIZE_PACKET    Error = 25
	MOSQ_ERR_OCSP               Error = 26
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

func Init() error {
	if err := Error(C.mosquitto_lib_init()); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func Cleanup() error {
	if err := Error(C.mosquitto_lib_cleanup()); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// NEW & DESTROY

func New(clientId string, clean bool, userInfo uintptr) (*Client, error) {
	cs := C.CString(clientId)
	defer C.free(unsafe.Pointer(cs))
	if handle := C.mosquitto_new(cs, C.bool(clean), unsafe.Pointer(userInfo)); handle == nil {
		return nil, errors.New("ERROR")
	} else {
		return (*Client)(handle), nil
	}
}

func (this *Client) Destroy() error {
	C.mosquitto_destroy((*C.struct_mosquitto)(this))
	return nil
}

func (this *Client) Reinitialise(clientId string, clean bool, userInfo uintptr) error {
	cs := C.CString(clientId)
	defer C.free(unsafe.Pointer(cs))
	if err := Error(C.mosquitto_reinitialise((*C.struct_mosquitto)(this), cs, C.bool(clean), unsafe.Pointer(userInfo))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// CONNECT

func (this *Client) SetCredentials(user, password string) error {
	cUser, cPassword := C.CString(user), C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))
	defer C.free(unsafe.Pointer(cUser))

	if err := Error(C.mosquitto_username_pw_set((*C.struct_mosquitto)(this), cUser, cPassword)); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) Connect(host string, port int, keepalive int) error {
	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))

	if err := Error(C.mosquitto_connect((*C.struct_mosquitto)(this), cHost,C.int(port),C.int(keepalive)); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}


func (this *Client) ConnectBind(host,bindAddress string, port int, keepalive int) error {
	cHost,cBindAddress := C.CString(host), C.CString(bindAddress)
	defer C.free(unsafe.Pointer(cHost))
	defer C.free(unsafe.Pointer(cBindAddress))

	if err := Error(C.mosquitto_connect_bind((*C.struct_mosquitto)(this), cHost,C.int(port),C.int(keepalive),cBindAddress); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// CLIENT

func (this *Client) String() string {
	return "<mosquitto.Client" +
		" client_id=??" +
		">"
}

////////////////////////////////////////////////////////////////////////////////
// ERRORS

func (e Error) Error() string {
	switch e {
	case MOSQ_ERR_AUTH_CONTINUE:
		return "MOSQ_ERR_AUTH_CONTINUE"
	case MOSQ_ERR_NO_SUBSCRIBERS:
		return "MOSQ_ERR_NO_SUBSCRIBERS"
	case MOSQ_ERR_SUB_EXISTS:
		return "MOSQ_ERR_SUB_EXISTS"
	case MOSQ_ERR_CONN_PENDING:
		return "MOSQ_ERR_CONN_PENDING"
	case MOSQ_ERR_SUCCESS:
		return "MOSQ_ERR_SUCCESS"
	case MOSQ_ERR_NOMEM:
		return "MOSQ_ERR_NOMEM"
	case MOSQ_ERR_PROTOCOL:
		return "MOSQ_ERR_PROTOCOL"
	case MOSQ_ERR_INVAL:
		return "MOSQ_ERR_INVAL"
	case MOSQ_ERR_NO_CONN:
		return "MOSQ_ERR_NO_CONN"
	case MOSQ_ERR_CONN_REFUSED:
		return "MOSQ_ERR_CONN_REFUSED"
	case MOSQ_ERR_NOT_FOUND:
		return "MOSQ_ERR_NOT_FOUND"
	case MOSQ_ERR_CONN_LOST:
		return "MOSQ_ERR_CONN_LOST"
	case MOSQ_ERR_TLS:
		return "MOSQ_ERR_TLS"
	case MOSQ_ERR_PAYLOAD_SIZE:
		return "MOSQ_ERR_PAYLOAD_SIZE"
	case MOSQ_ERR_NOT_SUPPORTED:
		return "MOSQ_ERR_NOT_SUPPORTED"
	case MOSQ_ERR_AUTH:
		return "MOSQ_ERR_AUTH"
	case MOSQ_ERR_ACL_DENIED:
		return "MOSQ_ERR_ACL_DENIED"
	case MOSQ_ERR_UNKNOWN:
		return "MOSQ_ERR_UNKNOWN"
	case MOSQ_ERR_ERRNO:
		return "MOSQ_ERR_ERRNO"
	case MOSQ_ERR_EAI:
		return "MOSQ_ERR_EAI"
	case MOSQ_ERR_PROXY:
		return "MOSQ_ERR_PROXY"
	case MOSQ_ERR_PLUGIN_DEFER:
		return "MOSQ_ERR_PLUGIN_DEFER"
	case MOSQ_ERR_MALFORMED_UTF8:
		return "MOSQ_ERR_MALFORMED_UTF8"
	case MOSQ_ERR_KEEPALIVE:
		return "MOSQ_ERR_KEEPALIVE"
	case MOSQ_ERR_LOOKUP:
		return "MOSQ_ERR_LOOKUP"
	case MOSQ_ERR_MALFORMED_PACKET:
		return "MOSQ_ERR_MALFORMED_PACKET"
	case MOSQ_ERR_DUPLICATE_PROPERTY:
		return "MOSQ_ERR_DUPLICATE_PROPERTY"
	case MOSQ_ERR_TLS_HANDSHAKE:
		return "MOSQ_ERR_TLS_HANDSHAKE"
	case MOSQ_ERR_QOS_NOT_SUPPORTED:
		return "MOSQ_ERR_QOS_NOT_SUPPORTED"
	case MOSQ_ERR_OVERSIZE_PACKET:
		return "MOSQ_ERR_OVERSIZE_PACKET"
	case MOSQ_ERR_OCSP:
		return "MOSQ_ERR_OCSP"
	default:
		return "[?? Invalid Error value]"
	}
}
