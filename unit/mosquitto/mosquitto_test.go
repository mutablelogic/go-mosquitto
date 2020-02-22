package mosquitto_test

import (
	"fmt"
	"testing"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	app "github.com/djthorpe/gopi/v2/app"
	mosquitto "github.com/djthorpe/mosquitto"

	// Units
	_ "github.com/djthorpe/mosquitto/unit/mosquitto"
)

const (
	TEST_SERVER         = "test.mosquitto.org"
	TEST_PORT_PLAINTEXT = 1883
)

func Test_Mosquitto_000(t *testing.T) {
	t.Log("Test_Mosquitto_000")
}

func Test_Mosquitto_001(t *testing.T) {
	args := []string{"-mqtt.host", TEST_SERVER, "-mqtt.port", fmt.Sprint(TEST_PORT_PLAINTEXT)}
	if app, err := app.NewTestTool(t, Main_Test_Mosquitto_001, args, "mosquitto"); err != nil {
		t.Error(err)
	} else {
		app.Run()
	}
}

func Main_Test_Mosquitto_001(app gopi.App, t *testing.T) {
	mosquitto := app.UnitInstance("mosquitto").(mosquitto.Client)

	if err := mosquitto.Connect(); err != nil {
		t.Error(err)
	} else {
		t.Log(mosquitto)
	}

}
