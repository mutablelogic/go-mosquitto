package mosquitto

import (
	"fmt"
	"reflect"
	"strconv"
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
	Message C.struct_mosquitto_message
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

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

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

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

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Message) String() string {
	str := "<message"
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
