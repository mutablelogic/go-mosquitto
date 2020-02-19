package mosquitto_test

import (
	"testing"

	// Frameworks
	mosquitto "github.com/djthorpe/mosquitto/sys/mosquitto"
)

func Test_Mosquitto_000(t *testing.T) {
	major, minor, revision := mosquitto.Version()
	t.Log("Version=", major, minor, revision)
}

func Test_Mosquitto_001(t *testing.T) {
	if err := mosquitto.Init(); err != nil {
		t.Error(err)
	} else if err := mosquitto.Cleanup(); err != nil {
		t.Error(err)
	}
}

func Test_Mosquitto_002(t *testing.T) {
	if err := mosquitto.Init(); err != nil {
		t.Error(err)
	} else {
		defer mosquitto.Cleanup()
		if client, err := mosquitto.New("id", true, 0); err != nil {
			t.Error(err)
		} else {
			t.Log(client)
		}
	}
}
