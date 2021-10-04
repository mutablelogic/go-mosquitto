package main

import (

	// Namespace imports
	"fmt"
	"sort"
	"time"

	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-mosquitto"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type topics struct {
	req    map[int]string
	topics map[string]time.Time
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTopics() *topics {
	t := new(topics)
	t.req = make(map[int]string)
	t.topics = make(map[string]time.Time)
	return t
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *topics) String() string {
	str := "<topics"
	if len(t.topics) > 0 {
		str += fmt.Sprintf(" %q", t.Topics())
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return topic subscriptions
func (t *topics) Topics() []string {
	result := make([]string, 0, len(t.topics))
	for k := range t.topics {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// Return true if there is a subscription
func (t *topics) Has(topic string) bool {
	for v := range t.topics {
		if v == topic {
			return true
		}
	}
	return false
}

// Subscribe request sent to a topic
func (t *topics) Subscribe(topic string, req int) {
	t.req[req] = topic
}

// Unsubscribe request sent to a topic
func (t *topics) Unsubscribe(topic string, req int) {
	t.req[req] = topic
}

// Unsubscribe event
func (t *topics) Event(evt Flags, req int) error {
	if topic, exists := t.req[req]; exists {
		delete(t.req, req)
		if evt == MOSQ_FLAG_EVENT_SUBSCRIBE {
			return t.add(topic)
		} else if evt == MOSQ_FLAG_EVENT_UNSUBSCRIBE {
			return t.delete(topic)
		}
	}
	return ErrUnexpectedResponse.Withf("%v (req %v)", evt, req)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (t *topics) add(topic string) error {
	if _, exists := t.topics[topic]; !exists {
		t.topics[topic] = time.Now()
	}
	// Return success
	return nil
}

func (t *topics) delete(topic string) error {
	if _, exists := t.topics[topic]; exists {
		delete(t.topics, topic)
	}
	// Return success
	return nil
}
