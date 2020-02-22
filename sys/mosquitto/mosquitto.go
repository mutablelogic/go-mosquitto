package mosquitto

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libmosquitto
#include <stdlib.h>
#include <mosquitto.h>

extern void onConnect(struct mosquitto*, void*, int, int);
extern void onDisconnect(struct mosquitto*, void*, int);
extern void onPublish(struct mosquitto*, void*, int);
extern void onSubscribe(struct mosquitto*, void*, int,int,int*);
extern void onUnsubscribe(struct mosquitto*, void*, int);
extern void onMessage(struct mosquitto*, void*, struct mosquitto_message*);
extern void onLog(struct mosquitto*,void*,int,char*);

static void set_connect_callback(struct mosquitto*	client) {
	mosquitto_connect_with_flags_callback_set(client,onConnect);
}
static void set_disconnect_callback(struct mosquitto* client) {
	mosquitto_disconnect_callback_set(client,onDisconnect);
}
static void set_publish_callback(struct mosquitto* client) {
	mosquitto_publish_callback_set(client,onPublish);
}
static void set_unsubscribe_callback(struct mosquitto* client) {
	mosquitto_unsubscribe_callback_set(client,onUnsubscribe);
}
static void set_message_callback(struct mosquitto* client) {
	mosquitto_message_callback_set(client,(void (*)(struct mosquitto *, void *, const struct mosquitto_message *))(onMessage));
}
static void set_log_callback(struct mosquitto*	client) {
	mosquitto_log_callback_set(client,(void (*)(struct mosquitto *, void *, int, const char *))(onLog));
}
static void set_subscribe_callback(struct mosquitto* client) {
	mosquitto_subscribe_callback_set(client,(void (*)(struct mosquitto *, void *, int, int, const int *))(onSubscribe));
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Error   int
	Client  C.struct_mosquitto
	Message C.struct_mosquitto_message
	Option  C.enum_mosq_opt_t
)

type (
	ConnectCallback     func(userInfo uintptr, rc, flags int)
	DisconnectCallback  func(userInfo uintptr, rc int)
	SubscribeCallback   func(userInfo uintptr, messageId int, GrantedQOS []int)
	UnsubscribeCallback func(userInfo uintptr, messageId int)
	PublishCallback     func(userInfo uintptr, messageId int)
	MessageCallback     func(userInfo uintptr, message *Message)
	LogCallback         func(userInfo uintptr, level int, str string)
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MOSQ_DEFAULT_PORT = 8883
)

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
// GLOBALS

var (
	callbacks = make(map[*C.struct_mosquitto]struct {
		ConnectCallback
		DisconnectCallback
		SubscribeCallback
		UnsubscribeCallback
		PublishCallback
		MessageCallback
		LogCallback
	})
	mutex sync.RWMutex
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
	cs := (*C.char)(nil)
	if clientId != "" {
		cs = C.CString(clientId)
		defer C.free(unsafe.Pointer(cs))
	}
	if handle := C.mosquitto_new(cs, C.bool(clean), unsafe.Pointer(userInfo)); handle == nil {
		return nil, fmt.Errorf("mosquitto_new failed: %v", cs)
	} else {
		return (*Client)(handle), nil
	}
}

func (this *Client) Destroy() error {
	C.mosquitto_destroy((*C.struct_mosquitto)(this))

	// Remove all callbacks
	mutex.Lock()
	defer mutex.Unlock()
	delete(callbacks, (*C.struct_mosquitto)(this))

	return nil
}

func (this *Client) Reinitialise(clientId string, clean bool, userInfo uintptr) error {
	cs := (*C.char)(nil)
	if clientId != "" {
		cs = C.CString(clientId)
		defer C.free(unsafe.Pointer(cs))
	}
	if err := Error(C.mosquitto_reinitialise((*C.struct_mosquitto)(this), cs, C.bool(clean), unsafe.Pointer(userInfo))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// CONNECT & DISCONNECT

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

func (this *Client) Connect(host string, port int, keepalive int, async bool) error {
	cHost := (*C.char)(nil)
	if host != "" {
		cHost = C.CString(host)
		defer C.free(unsafe.Pointer(cHost))
	}
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

func (this *Client) ConnectBind(host, bindAddress string, port int, keepalive int, async bool) error {
	cHost, cBindAddress := (*C.char)(nil), (*C.char)(nil)
	if host != "" {
		cHost = C.CString(host)
		defer C.free(unsafe.Pointer(cHost))
	}
	if bindAddress != "" {
		cBindAddress = C.CString(bindAddress)
		defer C.free(unsafe.Pointer(cBindAddress))
	}
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

func (this *Client) Disconnect() error {
	if err := Error(C.mosquitto_disconnect((*C.struct_mosquitto)(this))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// LOOP

func (this *Client) LoopForever(timeout_ms int) error {
	if err := Error(C.mosquitto_loop_forever((*C.struct_mosquitto)(this), C.int(timeout_ms), C.int(1))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) LoopStart() error {
	if err := Error(C.mosquitto_loop_start((*C.struct_mosquitto)(this))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) LoopStop(force bool) error {
	if err := Error(C.mosquitto_loop_stop((*C.struct_mosquitto)(this), C.bool(force))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) Loop(timeout_ms int) error {
	if err := Error(C.mosquitto_loop((*C.struct_mosquitto)(this), C.int(timeout_ms), C.int(1))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// CLIENT OPTIONS

func (this *Client) SetOptionInt(key Option, value int) error {
	if err := Error(C.mosquitto_int_option((*C.struct_mosquitto)(this), C.enum_mosq_opt_t(key), C.int(value))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) SetOptionPtr(key Option, value uintptr) error {
	if err := Error(C.mosquitto_void_option((*C.struct_mosquitto)(this), C.enum_mosq_opt_t(key), unsafe.Pointer(value))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) SetOptionString(key Option, value string) error {
	cStr := (*C.char)(nil)
	if value != "" {
		cStr = C.CString(value)
		defer C.free(unsafe.Pointer(cStr))
	}
	if err := Error(C.mosquitto_string_option((*C.struct_mosquitto)(this), C.enum_mosq_opt_t(key), cStr)); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) SetReconnectDelay(delay, max uint, exponential bool) error {
	if err := Error(C.mosquitto_reconnect_delay_set((*C.struct_mosquitto)(this), C.uint(delay), C.uint(max), C.bool(exponential))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) SetMaxInflightMessages(max uint) error {
	if err := Error(C.mosquitto_max_inflight_messages_set((*C.struct_mosquitto)(this), C.uint(max))); err != MOSQ_ERR_SUCCESS {
		return err
	} else {
		return nil
	}
}

func (this *Client) SetUserInfo(userInfo uintptr) error {
	C.mosquitto_user_data_set((*C.struct_mosquitto)(this), unsafe.Pointer(userInfo))
	return nil
}

func (this *Client) userInfo() uintptr {
	ptr := C.mosquitto_userdata((*C.struct_mosquitto)(this))
	return uintptr(ptr)
}

////////////////////////////////////////////////////////////////////////////////
// CALLBACKS

func (this *Client) SetConnectCallback(cb ConnectCallback) error {
	C.set_connect_callback((*C.struct_mosquitto)(this))

	// Set callback
	mutex.Lock()
	defer mutex.Unlock()
	handle := (*C.struct_mosquitto)(this)
	c, _ := callbacks[handle]
	c.ConnectCallback = cb
	callbacks[handle] = c

	// Return success
	return nil
}

func (this *Client) SetDisconnectCallback(cb DisconnectCallback) error {
	C.set_disconnect_callback((*C.struct_mosquitto)(this))

	// Set callback
	mutex.Lock()
	defer mutex.Unlock()
	handle := (*C.struct_mosquitto)(this)
	c, _ := callbacks[handle]
	c.DisconnectCallback = cb
	callbacks[handle] = c

	return nil
}

func (this *Client) SetPublishCallback(cb PublishCallback) error {
	C.set_publish_callback((*C.struct_mosquitto)(this))

	// Set callback
	mutex.Lock()
	defer mutex.Unlock()
	handle := (*C.struct_mosquitto)(this)
	c, _ := callbacks[handle]
	c.PublishCallback = cb
	callbacks[handle] = c

	return nil
}

func (this *Client) SetSubscribeCallback(cb SubscribeCallback) error {
	C.set_subscribe_callback((*C.struct_mosquitto)(this))

	// Set callback
	mutex.Lock()
	defer mutex.Unlock()
	handle := (*C.struct_mosquitto)(this)
	c, _ := callbacks[handle]
	c.SubscribeCallback = cb
	callbacks[handle] = c

	return nil
}

func (this *Client) SetUnsubscribeCallback(cb UnsubscribeCallback) error {
	C.set_unsubscribe_callback((*C.struct_mosquitto)(this))

	// Set callback
	mutex.Lock()
	defer mutex.Unlock()
	handle := (*C.struct_mosquitto)(this)
	c, _ := callbacks[handle]
	c.UnsubscribeCallback = cb
	callbacks[handle] = c

	return nil
}

func (this *Client) SetMessageCallback(cb MessageCallback) error {
	C.set_message_callback((*C.struct_mosquitto)(this))

	// Set callback
	mutex.Lock()
	defer mutex.Unlock()
	handle := (*C.struct_mosquitto)(this)
	c, _ := callbacks[handle]
	c.MessageCallback = cb
	callbacks[handle] = c

	return nil
}

func (this *Client) SetLogCallback(cb LogCallback) error {
	C.set_log_callback((*C.struct_mosquitto)(this))

	// Set callback
	mutex.Lock()
	defer mutex.Unlock()
	handle := (*C.struct_mosquitto)(this)
	c, _ := callbacks[handle]
	c.LogCallback = cb
	callbacks[handle] = c

	return nil
}

//export onConnect
func onConnect(handle *C.struct_mosquitto, userInfo unsafe.Pointer, rc C.int, flags C.int) {
	mutex.RLock()
	defer mutex.RUnlock()

	if c, exists := callbacks[handle]; exists && c.ConnectCallback != nil {
		c.ConnectCallback(uintptr(userInfo), int(rc), int(flags))
	}
}

//export onDisconnect
func onDisconnect(handle *C.struct_mosquitto, userInfo unsafe.Pointer, rc C.int) {
	mutex.RLock()
	defer mutex.RUnlock()

	if c, exists := callbacks[handle]; exists && c.DisconnectCallback != nil {
		c.DisconnectCallback(uintptr(userInfo), int(rc))
	}
}

//export onPublish
func onPublish(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId C.int) {
	mutex.RLock()
	defer mutex.RUnlock()

	fmt.Println("onPublish called")
}

//export onSubscribe
func onSubscribe(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId C.int, qosCount C.int, grantedQos *C.int) {
	mutex.RLock()
	defer mutex.RUnlock()

	fmt.Println("onSubscribe called")
}

//export onUnsubscribe
func onUnsubscribe(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId C.int) {
	mutex.RLock()
	defer mutex.RUnlock()

	fmt.Println("onUnsubscribe called")
}

//export onMessage
func onMessage(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId *C.struct_mosquitto_message) {
	mutex.RLock()
	defer mutex.RUnlock()

	fmt.Println("onMessage called")
}

//export onLog
func onLog(handle *C.struct_mosquitto, userInfo unsafe.Pointer, level C.int, str *C.char) {
	mutex.RLock()
	defer mutex.RUnlock()

	fmt.Println("onLog called")
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
		return syscall.Errno(e).Error()
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
