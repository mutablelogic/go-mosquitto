package mosquitto

import (
	"fmt"

	// Namespace imports
	. "github.com/mutablelogic/go-mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Event struct {
	Type  Flags
	Err   error
	Id    int
	Topic string
	Data  []byte
}

////////////////////////////////////////////////////////////////////////////////
// NEW MESSAGES

func NewConnect(err error) *Event {
	return &Event{
		Type: MOSQ_FLAG_EVENT_CONNECT,
		Err:  err,
	}
}

func NewDisconnect(err error) *Event {
	return &Event{
		Type: MOSQ_FLAG_EVENT_DISCONNECT,
		Err:  err,
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
		Type:  MOSQ_FLAG_EVENT_MESSAGE,
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
		str += fmt.Sprint(" ", t)
	}
	if err := e.Err; err != nil {
		str += fmt.Sprintf(" err=%q", err)
	}
	if id := e.Id; id != 0 {
		str += fmt.Sprint(" id=", id)
	}
	if topic := e.Topic; topic != "" {
		str += fmt.Sprintf(" topic=%q", topic)
	}
	if data := e.Data; len(data) > 0 {
		str += fmt.Sprintf(" data=%q", string(data))
	}
	return str + ">"
}
