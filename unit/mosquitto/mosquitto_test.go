package mosquitto_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	app "github.com/djthorpe/gopi/v2/app"
	mosq "github.com/djthorpe/mosquitto"

	// Units
	_ "github.com/djthorpe/gopi/v2/unit/bus"
	_ "github.com/djthorpe/mosquitto/unit/mosquitto"
)

const (
	TEST_SERVER = "test.mosquitto.org:1883"
)

func Test_Mosquitto_000(t *testing.T) {
	t.Log("Test_Mosquitto_000")
}

func Test_Mosquitto_001(t *testing.T) {
	args := []string{"-mqtt.broker", TEST_SERVER}
	if app, err := app.NewTestTool(t, Main_Test_Mosquitto_001, args, "mosquitto"); err != nil {
		t.Error(err)
	} else {
		app.Run()
	}
}

func Main_Test_Mosquitto_001(app gopi.App, t *testing.T) {
	mosquitto := app.UnitInstance("mosquitto").(mosq.Client)
	bus := app.Bus()
	bus.DefaultHandler(gopi.EVENT_NS_DEFAULT, func(_ context.Context, _ gopi.App, evt gopi.Event) {
		t.Log(evt)
	})
	if err := mosquitto.Connect(); err != nil {
		t.Error(err)
	} else if _, err := mosquitto.Subscribe("#"); err != nil {
		t.Error(err)
	} else {
		time.Sleep(1 * time.Second)
	}
}

func Test_Mosquitto_002(t *testing.T) {
	args := []string{"-mqtt.broker", TEST_SERVER}
	if app, err := app.NewTestTool(t, Main_Test_Mosquitto_002, args, "mosquitto"); err != nil {
		t.Error(err)
	} else {
		app.Run()
	}
}

func Main_Test_Mosquitto_002(app gopi.App, t *testing.T) {
	mosquitto := app.UnitInstance("mosquitto").(mosq.Client)
	bus := app.Bus()
	bus.DefaultHandler(gopi.EVENT_NS_DEFAULT, func(_ context.Context, _ gopi.App, evt gopi.Event) {
		t.Log(evt)
	})
	if err := mosquitto.Connect(); err != nil {
		t.Error(err)
	} else {
		time.Sleep(2 * time.Second)
		for i := 0; i < 10; i++ {
			payload := []byte(fmt.Sprintf("hello, world: %v", i))
			if _, err := mosquitto.Publish("test", payload); err != nil {
				t.Error(err)
			}
			time.Sleep(100 * time.Millisecond)
		}
		if _, err := mosquitto.Unsubscribe("test"); err != nil {
			t.Error(err)
		} else {
			time.Sleep(2 * time.Second)
		}
		if err := mosquitto.Disconnect(); err != nil {
			t.Error(err)
		}
	}
}
