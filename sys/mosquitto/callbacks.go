package mosquitto

import (
	"reflect"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libmosquitto
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
	mosquitto_connect_callback_set(client, onConnect);
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
	ConnectCallback     func(Error)         // Connect(return_code int)
	DisconnectCallback  func(Error)         // Disconnect(return_code int)
	SubscribeCallback   func(int, []int)    // Subscribe(message_id int, granted_qos []int)
	UnsubscribeCallback func(int)           // Unsubscribe(message_id int)
	PublishCallback     func(int)           // Publish(message_id int)
	MessageCallback     func(*Message)      // Message(message *Message)
	LogCallback         func(Level, string) // Log(level Level, message string)
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *ClientEx) SetConnectCallback(cb ConnectCallback) {
	C.set_connect_callback((*C.struct_mosquitto)(c.Client))
	c.ConnectCallback = cb
}

func (c *ClientEx) SetDisconnectCallback(cb DisconnectCallback) {
	C.set_disconnect_callback((*C.struct_mosquitto)(c.Client))
	c.DisconnectCallback = cb
}

func (c *ClientEx) SetPublishCallback(cb PublishCallback) {
	C.set_publish_callback((*C.struct_mosquitto)(c.Client))
	c.PublishCallback = cb
}

func (c *ClientEx) SetSubscribeCallback(cb SubscribeCallback) {
	C.set_subscribe_callback((*C.struct_mosquitto)(c.Client))
	c.SubscribeCallback = cb
}

func (c *ClientEx) SetUnsubscribeCallback(cb UnsubscribeCallback) {
	C.set_unsubscribe_callback((*C.struct_mosquitto)(c.Client))
	c.UnsubscribeCallback = cb
}

func (c *ClientEx) SetMessageCallback(cb MessageCallback) {
	C.set_message_callback((*C.struct_mosquitto)(c.Client))
	c.MessageCallback = cb
}

func (c *ClientEx) SetLogCallback(cb LogCallback) {
	C.set_log_callback((*C.struct_mosquitto)(c.Client))
	c.LogCallback = cb
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

//export onConnect
func onConnect(handle *C.struct_mosquitto, userInfo unsafe.Pointer, rc C.int) {
	client := (*ClientEx)(userInfo)
	if client.ConnectCallback != nil {
		client.ConnectCallback(Error(rc))
	}
}

//export onDisconnect
func onDisconnect(handle *C.struct_mosquitto, userInfo unsafe.Pointer, rc C.int) {
	client := (*ClientEx)(userInfo)
	if client.DisconnectCallback != nil {
		client.DisconnectCallback(Error(rc))
	}
}

//export onPublish
func onPublish(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId C.int) {
	client := (*ClientEx)(userInfo)
	if client.PublishCallback != nil {
		client.PublishCallback(int(messageId))
	}
}

//export onSubscribe
func onSubscribe(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId C.int, qosCount C.int, grantedQos *C.int) {
	client := (*ClientEx)(userInfo)
	if client.SubscribeCallback != nil {
		var data []C.int
		header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
		header.Data = uintptr(unsafe.Pointer(grantedQos))
		header.Len = int(qosCount)
		header.Cap = int(qosCount)

		qos := make([]int, len(data))
		for i, value := range data {
			qos[i] = int(value)
		}
		client.SubscribeCallback(int(messageId), qos)
	}
}

//export onUnsubscribe
func onUnsubscribe(handle *C.struct_mosquitto, userInfo unsafe.Pointer, messageId C.int) {
	client := (*ClientEx)(userInfo)
	if client.UnsubscribeCallback != nil {
		client.UnsubscribeCallback(int(messageId))
	}
}

//export onMessage
func onMessage(handle *C.struct_mosquitto, userInfo unsafe.Pointer, message *C.struct_mosquitto_message) {
	client := (*ClientEx)(userInfo)
	if client.MessageCallback != nil {
		client.MessageCallback((*Message)(message))
	}
}

//export onLog
func onLog(handle *C.struct_mosquitto, userInfo unsafe.Pointer, level C.int, str *C.char) {
	client := (*ClientEx)(userInfo)
	if client.LogCallback != nil {
		client.LogCallback(Level(level), C.GoString(str))
	}
}
