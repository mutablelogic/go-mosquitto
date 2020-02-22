package mosquitto_test

import (
	"testing"
	"time"

	// Frameworks
	mosquitto "github.com/djthorpe/mosquitto/sys/mosquitto"
)

const (
	TEST_SERVER         = "test.mosquitto.org"
	TEST_PORT_PLAINTEXT = 1883
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

func Test_Mosquitto_003(t *testing.T) {
	if err := mosquitto.Init(); err != nil {
		t.Error(err)
	} else {
		defer mosquitto.Cleanup()
		if client, err := mosquitto.New("id", true, 0); err != nil {
			t.Error(err)
		} else if err := client.Connect(TEST_SERVER, TEST_PORT_PLAINTEXT, 1, false); err != nil {
			t.Error(err)
		} else if err := client.Disconnect(); err != nil {
			t.Error(err)
		} else {
			t.Log(client)
		}
	}
}

func Test_Mosquitto_004(t *testing.T) {
	if err := mosquitto.Init(); err != nil {
		t.Error(err)
	} else {
		defer mosquitto.Cleanup()
		if client, err := mosquitto.New("id", true, 0); err != nil {
			t.Error(err)
		} else if err := client.SetConnectCallback(func(userInfo uintptr, rc, flags int) {
			t.Log("onConnect", userInfo, rc, flags)
		}); err != nil {
			t.Error(err)
		} else if err := client.SetDisconnectCallback(func(userInfo uintptr, rc int) {
			t.Log("onDisconnect", userInfo, rc)
		}); err != nil {
			t.Error(err)
		} else {
			go func() {
				if err := client.LoopForever(100); err != nil {
					t.Error(err)
				} else {
					t.Log("Loop ended")
				}
			}()
			if err := client.Connect(TEST_SERVER, TEST_PORT_PLAINTEXT, 1, false); err != nil {
				t.Error(err)
			} else {
				time.Sleep(time.Second * 5)
				if err := client.Disconnect(); err != nil {
					t.Error(err)
				} else {
					t.Log(client)
				}
			}
		}
	}
}
