package mosquitto

import (
	"fmt"
	"os"
	"unsafe"
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
	Client C.struct_mosquitto
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New is called to create a new empty client object
func New(clientId string, clean bool, userInfo unsafe.Pointer) (*Client, error) {
	var cClientId *C.char
	if clientId != "" {
		cClientId = C.CString(clientId)
	}
	defer C.free(unsafe.Pointer(cClientId))
	if handle := C.mosquitto_new(cClientId, C.bool(clean), userInfo); handle == nil {
		return nil, fmt.Errorf("mosquitto_new failed")
	} else {
		return (*Client)(handle), nil
	}
}

// Destroy is called when you have finished using a client
func (this *Client) Destroy() error {
	C.mosquitto_destroy((*C.struct_mosquitto)(this))
	return nil
}

// Reinitalize a client object
func (this *Client) Reinitialise(clientId string, clean bool, userInfo unsafe.Pointer) error {
	var cClientId *C.char
	if clientId != "" {
		cClientId = C.CString(clientId)
	}
	defer C.free(unsafe.Pointer(cClientId))
	if err := Error(C.mosquitto_reinitialise((*C.struct_mosquitto)(this), cClientId, C.bool(clean), userInfo)); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// CONNECT & DISCONNECT

// Set username and password for connecting to a broker. Call this before Connect
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

// Connect to a broker using host and port, setting the keepalive time in seconds
// and use 'true' for the async parameter to connect asyncronously
func (this *Client) Connect(host string, port int, keepalive int, async bool) error {
	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))

	if async {
		if err := Error(C.mosquitto_connect_async((*C.struct_mosquitto)(this), cHost, C.int(port), C.int(keepalive))); err != MOSQ_ERR_SUCCESS {
			return err
		} else {
			return nil
		}
	} else {
		if err := Error(C.mosquitto_connect((*C.struct_mosquitto)(this), cHost, C.int(port), C.int(keepalive))); err != MOSQ_ERR_SUCCESS {
			return err
		} else {
			return nil
		}
	}
}

// Connect to a broker using host and port, setting the keepalive time in seconds
// and use 'true' for the async parameter to connect asyncronously. Connects to
// a specific interface.
func (this *Client) ConnectBind(host, bindAddress string, port int, keepalive int, async bool) error {
	cHost, cBindAddress := C.CString(host), C.CString(bindAddress)
	defer C.free(unsafe.Pointer(cHost))
	defer C.free(unsafe.Pointer(cBindAddress))

	if async {
		if err := Error(C.mosquitto_connect_bind_async((*C.struct_mosquitto)(this), cHost, C.int(port), C.int(keepalive), cBindAddress)); err != MOSQ_ERR_SUCCESS {
			return err
		} else {
			return nil
		}
	} else {
		if err := Error(C.mosquitto_connect_bind((*C.struct_mosquitto)(this), cHost, C.int(port), C.int(keepalive), cBindAddress)); err != MOSQ_ERR_SUCCESS {
			return err
		} else {
			return nil
		}
	}
}

// Reconnect to a broker when disconnect has occured.
func (this *Client) Reconnect(async bool) error {
	if async {
		if err := Error(C.mosquitto_reconnect_async((*C.struct_mosquitto)(this))); err != MOSQ_ERR_SUCCESS {
			return err
		} else {
			return nil
		}
	} else {
		if err := Error(C.mosquitto_reconnect((*C.struct_mosquitto)(this))); err != MOSQ_ERR_SUCCESS {
			return err
		} else {
			return nil
		}
	}
}

// Disconnect from a broker
func (this *Client) Disconnect() error {
	if err := Error(C.mosquitto_disconnect((*C.struct_mosquitto)(this))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// LOOP

// Loop and perform actions on a regular basis.
func (this *Client) LoopForever(timeout_ms int) error {
	if err := Error(C.mosquitto_loop_forever((*C.struct_mosquitto)(this), C.int(timeout_ms), C.int(1))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

// Start the event loop thread, to be called before Connect
func (this *Client) LoopStart() error {
	if err := Error(C.mosquitto_loop_start((*C.struct_mosquitto)(this))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

// Stop the event loop thread, to be called after Disconnect has completed
func (this *Client) LoopStop(force bool) error {
	if err := Error(C.mosquitto_loop_stop((*C.struct_mosquitto)(this), C.bool(force))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

// Loop for a specific period of time
func (this *Client) Loop(timeout_ms int) error {
	if err := Error(C.mosquitto_loop((*C.struct_mosquitto)(this), C.int(timeout_ms), C.int(1))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// SUBSCRIBE & UNSUBSCRIBE

// Subscribe to one set of topics and return the id of the request
func (this *Client) Subscribe(topics string, qos int) (int, error) {
	var messageId C.int
	cTopics := C.CString(topics)
	defer C.free(unsafe.Pointer(cTopics))

	if err := Error(C.mosquitto_subscribe((*C.struct_mosquitto)(this), &messageId, cTopics, C.int(qos))); err != MOSQ_ERR_SUCCESS {
		return 0, err
	} else {
		return int(messageId), nil
	}
}

// Unsubscribe from one set of topics and return the id of the request
func (this *Client) Unsubscribe(topics string) (int, error) {
	var messageId C.int
	cTopics := C.CString(topics)
	defer C.free(unsafe.Pointer(cTopics))

	if err := Error(C.mosquitto_unsubscribe((*C.struct_mosquitto)(this), &messageId, cTopics)); err != MOSQ_ERR_SUCCESS {
		return 0, err
	} else {
		return int(messageId), nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLISH

// Publish a message to the broker in a topic and return the id of the request
func (this *Client) Publish(topic string, data []byte, qos int, retain bool) (int, error) {
	var messageId C.int
	cTopic := C.CString(topic)
	defer C.free(unsafe.Pointer(cTopic))
	payloadlen := len(data)
	payload := unsafe.Pointer(&data[0])
	if err := Error(C.mosquitto_publish((*C.struct_mosquitto)(this), &messageId, cTopic, C.int(payloadlen), unsafe.Pointer(payload), C.int(qos), C.bool(retain))); err != MOSQ_ERR_SUCCESS {
		return 0, err
	} else {
		return int(messageId), nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// CLIENT OPTIONS

// SetTLS sets certificate authority, cert and key  for TLS connections.
// The certificate authority can either be a file or path to files.
// This version does not accept a callback for password, use ClientEx
// for that.
func (c *Client) SetTLS(capath, certpath, keypath string) error {
	var cCaFile, cCaPath, cCertFile, cKeyFile *C.char

	// If capath is a directory, use directory form or else use file form.
	if stat, err := os.Stat(capath); err != nil {
		return err
	} else if stat.IsDir() {
		cCaPath = C.CString(capath)
	} else {
		cCaFile = C.CString(capath)
	}
	if certpath != "" {
		cCertFile = C.CString(certpath)
	}
	if keypath != "" {
		cKeyFile = C.CString(keypath)
	}
	defer C.free(unsafe.Pointer(cCaFile))
	defer C.free(unsafe.Pointer(cCaPath))
	defer C.free(unsafe.Pointer(cCertFile))
	defer C.free(unsafe.Pointer(cKeyFile))
	if err := Error(C.mosquitto_tls_set((*C.struct_mosquitto)(c), cCaFile, cCaPath, cCertFile, cKeyFile, nil)); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

// SetTLSInsecure configures verification of the server hostname in the server certificate.
// If value is set to true, it is impossible to guarantee that the host you are connecting
// to is not impersonating your server.
func (c *Client) SetTLSInsecure(v bool) error {
	if err := Error(C.mosquitto_tls_insecure_set((*C.struct_mosquitto)(c), C.bool(v))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

// Value must be set to either MQTT_PROTOCOL_V31, MQTT_PROTOCOL_V311, or MQTT_PROTOCOL_V5.  Must be set before the client connects.  Defaults to MQTT_PROTOCOL_V311.
func (this *Client) SetProtocol(protocol int) error {
	if err := Error(C.mosquitto_opts_set((*C.struct_mosquitto)(this), C.enum_mosq_opt_t(MOSQ_OPT_PROTOCOL_VERSION), unsafe.Pointer(uintptr(protocol)))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

// Control the behaviour of the client when it has unexpectedly disconnected
// The default behaviour if this function is not used is to repeatedly attempt to reconnect
// with a delay of 1 second until the connection succeeds.
func (this *Client) SetReconnectDelay(delay, max uint, exponential bool) error {
	if err := Error(C.mosquitto_reconnect_delay_set((*C.struct_mosquitto)(this), C.uint(delay), C.uint(max), C.bool(exponential))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) SetUserInfo(userInfo unsafe.Pointer) error {
	C.mosquitto_user_data_set((*C.struct_mosquitto)(this), userInfo)
	return nil
}
