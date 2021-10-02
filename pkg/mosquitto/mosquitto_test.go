package mosquitto_test

import (
	"context"
	"testing"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-mosquitto/pkg/mosquitto"
)

const (
	BrokerHost = "test.mosquitto.org"
)

func Test_Mosquitto_001(t *testing.T) {
	client, err := New(BrokerHost, func(evt *Event) {
		t.Log("Event", evt)
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Log(client)
	<-ctx.Done()

	if err := client.Close(); err != nil {
		t.Error(err)
	}
}

func Test_Mosquitto_002(t *testing.T) {
	cfg := NewConfigWithBroker(BrokerHost).WithCallback(func(evt *Event) {
		t.Log("Event", evt)
	}).WithTrace(func(message string) {
		t.Log(message)
	})

	client, err := NewWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if _, err := client.Subscribe("$SYS/#", 0); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	<-ctx.Done()
}
