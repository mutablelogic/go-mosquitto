package mosquitto

import (
	"fmt"

	// Namespace imports
	. "github.com/djthorpe/go-mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Event struct {
	Type       Flags
	ReturnCode int
	Id         int
	Topic      string
	Data       []byte
}

////////////////////////////////////////////////////////////////////////////////
// NEW MESSAGES

func NewConnect(returnCode int) *Event {
	return &Event{
		Type:       MOSQ_FLAG_EVENT_CONNECT,
		ReturnCode: returnCode,
	}
}

func NewDisconnect(returnCode int) *Event {
	return &Event{
		Type:       MOSQ_FLAG_EVENT_DISCONNECT,
		ReturnCode: returnCode,
	}
}

func NewSubscribe(id int) *Event {
	return &Event{
		Type: MOSQ_FLAG_EVENT_SUBSCRIBE,
		Id:   id,
	}
}

func NewUnsubscribe(id int) *Event {
	return &Event{
		Type: MOSQ_FLAG_EVENT_UNSUBSCRIBE,
		Id:   id,
	}
}

func NewPublish(id int) *Event {
	return &Event{
		Type: MOSQ_FLAG_EVENT_PUBLISH,
		Id:   id,
	}
}

func NewMessage(id int, topic string, data []byte) *Event {
	return &Event{
		Type:  MOSQ_FLAG_EVENT_PUBLISH,
		Id:    id,
		Topic: topic,
		Data:  data[:],
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e *Event) String() string {
	str := "<event"
	if t := e.Type; t != 0 {
		str += fmt.Sprint(" type=", t)
	}
	return str + ">"
}
