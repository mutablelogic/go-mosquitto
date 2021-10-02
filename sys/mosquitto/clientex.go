package mosquitto

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libmosquitto
#include <stdlib.h>
#include <mosquitto.h>
*/
import "C"
import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientEx struct {
	*Client
	ConnectCallback
	DisconnectCallback
	SubscribeCallback
	UnsubscribeCallback
	PublishCallback
	MessageCallback
	LogCallback
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewEx returns a new client object, with callback support. If the clientId
// parameter is empty a random clientId will be generated. The clean flag
// instructs the broker to clean all messages and subscriptions on disconnect
// https://mosquitto.org/api/files/mosquitto-h.html#mosquitto_new
func NewEx(clientId string, clean bool) (*ClientEx, error) {
	c := new(ClientEx)
	userInfo := unsafe.Pointer(c)
	if client, err := New(clientId, clean, userInfo); err != nil {
		return nil, err
	} else {
		c.Client = client
	}

	// Return success
	return c, nil
}

// Destroy is called when you have finished using a client
func (c *ClientEx) Destroy() error {
	c.Client.SetUserInfo(nil)
	return c.Client.Destroy()
}

// Reinitalize a client object
func (c *ClientEx) Reinitialise(clientId string, clean bool) error {
	c.ConnectCallback = nil
	c.DisconnectCallback = nil
	c.SubscribeCallback = nil
	c.UnsubscribeCallback = nil
	c.PublishCallback = nil
	c.MessageCallback = nil
	c.LogCallback = nil
	return c.Client.Reinitialise(clientId, clean, unsafe.Pointer(c))
}
