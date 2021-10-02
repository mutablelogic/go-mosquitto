package mosquitto_test

import (
	"testing"

	// Namespace imports
	. "github.com/djthorpe/go-mosquitto/pkg/mosquitto"
)

const (
	BrokerHost = "test.mosquitto.org"
)

func Test_Mosquitto_001(t *testing.T) {
	cfg := NewConfigWithBroker(BrokerHost)
	client, err := NewWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(client)
	if err := client.Close(); err != nil {
		t.Error(err)
	}
}

func Test_Mosquitto_002(t *testing.T) {
	cfg := NewConfigWithBroker(BrokerHost).WithCallback(func(evt *Event) {
		t.Log(evt)
	})
	client, err := NewWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(client)
	if err := client.Close(); err != nil {
		t.Error(err)
	}
}
