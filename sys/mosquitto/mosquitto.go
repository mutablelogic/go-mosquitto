package mosquitto

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libmosquitto
#include <stdio.h>
#include <stdlib.h>
#include <mosquitto.h>

extern void onConnect(struct mosquitto*, void*, int);
extern void onDisconnect(struct mosquitto*, void*, int);
extern void onPublish(struct mosquitto*, void*, int);
extern void onSubscribe(struct mosquitto*, void*, int,int,int*);
extern void onUnsubscribe(struct mosquitto*, void*, int);
extern void onMessage(struct mosquitto*, void*, struct mosquitto_message*);
extern void onLog(struct mosquitto*,void*,int,char*);

static void set_connect_callback(struct mosquitto*	client) {
	mosquitto_connect_callback_set(client,onConnect);
}
static void set_disconnect_callback(struct mosquitto* client) {
	mosquitto_disconnect_callback_set(client,onDisconnect);
}
static void set_publish_callback(struct mosquitto* client) {
	mosquitto_publish_callback_set(client,onPublish);
}
static void set_subscribe_callback(struct mosquitto* client) {
	mosquitto_subscribe_callback_set(client,(void (*)(struct mosquitto *, void *, int, int, const int *))(onSubscribe));
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
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Error   int
	Level   int
	Client  C.struct_mosquitto
	Message C.struct_mosquitto_message
	Option  C.enum_mosq_opt_t
)

type (
	ConnectCallback     func(uintptr, int)
	DisconnectCallback  func(uintptr, int)
	SubscribeCallback   func(uintptr, int, []int)
	UnsubscribeCallback func(uintptr, int)
	PublishCallback     func(uintptr, int)
	MessageCallback     func(uintptr, *Message)
	LogCallback         func(uintptr, Level, string)
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MOSQ_DEFAULT_PORT = 1883
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

// Init initiaizes the library
func Init() error {
	if callbacks == nil {
		callbacks = make(map[*C.struct_mosquitto]struct {
			ConnectCallback
			DisconnectCallback
			SubscribeCallback
			UnsubscribeCallback
			PublishCallback
			MessageCallback
			LogCallback
		})
	}

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

	// Release callback resources
	mutex.Lock()
	defer mutex.Unlock()
	for k := range callbacks {
		delete(callbacks, k)
	}
	callbacks = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// NEW & DESTROY

// New is called to create a new empty client object
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

// Destroy is called when you have finished using a client
func (this *Client) Destroy() error {
	C.mosquitto_destroy((*C.struct_mosquitto)(this))

	// Remove all callbacks
	mutex.Lock()
	defer mutex.Unlock()
	delete(callbacks, (*C.struct_mosquitto)(this))

	return nil
}

// Reinitalize a client object
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

/*
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
*/

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

/*
func (this *Client) SetUserInfo(userInfo uintptr) error {
	C.mosquitto_user_data_set((*C.struct_mosquitto)(this), unsafe.Pointer(userInfo))
	return nil
}

func (this *Client) userInfo() uintptr {
	ptr := C.mosquitto_userdata((*C.struct_mosquitto)(this))
	return uintptr(ptr)
}
*/

////////////////////////////////////////////////////////////////////////////////
// MESSAGES

func NewMessage(id int) *Message {
	this := new(Message)
	this.mid = C.int(id)
	return this
}

func (this *Message) Free() {
	handle := (*C.struct_mosquitto_message)(this)
	C.mosquitto_message_free(&handle)
}

func (this *Message) Copy() *Message {
	other := new(Message)
	if err := Error(C.mosquitto_message_copy((*C.struct_mosquitto_message)(other), (*C.struct_mosquitto_message)(this))); err != MOSQ_ERR_SUCCESS {
		return nil
	} else {
		return other
	}
}

func (this *Message) FreeContents() {
	C.mosquitto_message_free_contents((*C.struct_mosquitto_message)(this))
}

func (this *Message) Id() int {
	return int(this.mid)
}

func (this *Message) Topic() string {
	if this.topic == nil {
		return ""
	} else {
		return C.GoString(this.topic)
	}
}

func (this *Message) Len() uint {
	return uint(this.payloadlen)
}

func (this *Message) Qos() int {
	return int(this.qos)
}

func (this *Message) Retain() bool {
	return bool(this.retain)
}

func (this *Message) Data() []byte {
	var data []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	header.Data = uintptr(this.payload)
	header.Len = int(this.payloadlen)
	header.Cap = int(this.payloadlen)
	return data
}

func (this *Message) String() string {
	str := "<mosq.Message"
	if id := this.Id(); id != 0 {
		str += " id=" + fmt.Sprint(this.Id())
	}
	if topic := this.Topic(); topic != "" {
		str += " topic=" + strconv.Quote(topic)
	}
	if this.Len() > 0 {
		str += " data=" + fmt.Sprint(this.Data())
	}
	if this.Retain() {
		str += " retain=" + fmt.Sprint(true)
	}
	return str + ">"
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

	// Return success
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

	// Return success
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

	// Return success
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

	// Return success
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

	// Return success
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

	// Return success
	return nil
}

//export onConnect
func onConnect(handle *C.struct_mosquitto, userInfo unsafe.Pointer, rc C.int) {
	mutex.RLock()
	defer mutex.RUnlock()

	if c, exists := callbacks[handle]; exists && c.ConnectCallback != nil {
		c.ConnectCallback(uintptr(userInfo), int(rc))
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

	if c, exists := callbacks[handle]; exists && c.PublishCallback != nil {
		c.PublishCallback(uintptr(userInfo), int(messageId))
	}
}

//export onSubscribe
func onSubscribe(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId C.int, qosCount C.int, grantedQos *C.int) {
	mutex.RLock()
	defer mutex.RUnlock()

	if c, exists := callbacks[handle]; exists && c.SubscribeCallback != nil {
		var data []C.int
		header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
		header.Data = uintptr(unsafe.Pointer(grantedQos))
		header.Len = int(qosCount)
		header.Cap = int(qosCount)

		qos := make([]int, len(data))
		for i, value := range data {
			qos[i] = int(value)
		}
		c.SubscribeCallback(uintptr(userInfo), int(messageId), qos)
	}
}

//export onUnsubscribe
func onUnsubscribe(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId C.int) {
	mutex.RLock()
	defer mutex.RUnlock()

	if c, exists := callbacks[handle]; exists && c.UnsubscribeCallback != nil {
		c.UnsubscribeCallback(uintptr(userInfo), int(messageId))
	}
}

//export onMessage
func onMessage(handle *C.struct_mosquitto, userInfo unsafe.Pointer, message *C.struct_mosquitto_message) {
	mutex.RLock()
	defer mutex.RUnlock()

	if c, exists := callbacks[handle]; exists && c.MessageCallback != nil {
		c.MessageCallback(uintptr(userInfo), (*Message)(message))
	}
}

//export onLog
func onLog(handle *C.struct_mosquitto, userInfo unsafe.Pointer, level C.int, str *C.char) {
	mutex.RLock()
	defer mutex.RUnlock()

	if c, exists := callbacks[handle]; exists && c.LogCallback != nil {
		if str != nil {
			c.LogCallback(uintptr(userInfo), Level(level), C.GoString(str))
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

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
