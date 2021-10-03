package mosquitto_test

import (
	"sync"
	"testing"
	"time"

	// Namespace imports

	. "github.com/mutablelogic/go-mosquitto/sys/mosquitto"
)

const (
	TEST_SERVER         = "test.mosquitto.org"
	TEST_PORT_PLAINTEXT = MOSQ_DEFAULT_PORT
)

func Test_Mosquitto_000(t *testing.T) {
	major, minor, revision := Version()
	t.Log("Version=", major, minor, revision)
}

func Test_Mosquitto_001(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	if err := Cleanup(); err != nil {
		t.Fatal(err)
	}
}

func Test_Mosquitto_002(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	defer Cleanup()
	if client, err := NewEx("id", true); err != nil {
		t.Error(err)
	} else {
		t.Log(client)
	}
}

func Test_Mosquitto_003(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	defer Cleanup()
	if client, err := NewEx("id", true); err != nil {
		t.Error(err)
	} else if err := client.Connect(TEST_SERVER, TEST_PORT_PLAINTEXT, 5, false); err != nil {
		t.Error(err)
	} else if err := client.Disconnect(); err != nil {
		t.Error(err)
	} else {
		t.Log(client)
	}
}

func Test_Mosquitto_004(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	defer Cleanup()

	client, err := NewEx("id", true)
	if err != nil {
		t.Error(err)
	}
	client.SetConnectCallback(func(rc Error) {
		if rc != MOSQ_ERR_SUCCESS {
			t.Error("onConnect", rc)
		}
	})
	client.SetDisconnectCallback(func(rc Error) {
		if rc != MOSQ_ERR_SUCCESS {
			t.Error("onDisconnect", rc)
		}
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := client.LoopForever(100); err != nil {
			t.Error(err)
		} else {
			t.Log("Loop ended")
		}
	}()

	// Connect then disconnect
	err = client.Connect(TEST_SERVER, TEST_PORT_PLAINTEXT, 5, false)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 5)
	if err := client.Disconnect(); err != nil {
		t.Error(err)
	}

	// Wait for end of goroutine
	wg.Wait()
}

func Test_Mosquitto_005(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	defer Cleanup()

	if msg := NewMessage(100); msg == nil {
		t.Error("msg returned nil")
	} else {
		t.Log(msg)
	}
}

func Test_Mosquitto_006(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	defer Cleanup()

	client, err := NewEx("id", true)
	if err != nil {
		t.Error(err)
	}
	client.SetConnectCallback(func(rc Error) {
		if rc != MOSQ_ERR_SUCCESS {
			t.Error("onConnect", rc)
		}
	})
	client.SetDisconnectCallback(func(rc Error) {
		if rc != MOSQ_ERR_SUCCESS {
			t.Error("onDisconnect", rc)
		}
	})
	client.SetSubscribeCallback(func(messageId int, GrantedQOS []int) {
		t.Log("onSubscribe", messageId, GrantedQOS)
	})
	client.SetUnsubscribeCallback(func(messageId int) {
		t.Log("onUnsubscribe", messageId)
	})
	client.SetMessageCallback(func(message *Message) {
		t.Log("onMessage", message.Topic(), string(message.Data()))
	})

	if err := client.Connect(TEST_SERVER, TEST_PORT_PLAINTEXT, 60, false); err != nil {
		t.Error(err)
	}
	if err := client.LoopStart(); err != nil {
		t.Error(err)
	}
	if _, err := client.Subscribe("$SYS/#", 0); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second * 5)

	if _, err := client.Unsubscribe("#"); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)

	if err := client.Disconnect(); err != nil {
		t.Error(err)
	}
}

func Test_Mosquitto_007(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	defer Cleanup()

	client, err := NewEx("id", true)
	if err != nil {
		t.Error(err)
	}
	defer client.Destroy()

	client.SetConnectCallback(func(rc Error) {
		if rc != MOSQ_ERR_SUCCESS {
			t.Error("onConnect", rc)
		}
	})
	client.SetDisconnectCallback(func(rc Error) {
		if rc != MOSQ_ERR_SUCCESS {
			t.Error("onDisconnect", rc)
		}
	})
	client.SetSubscribeCallback(func(messageId int, GrantedQOS []int) {
		t.Log("onSubscribe", messageId, GrantedQOS)
	})
	client.SetUnsubscribeCallback(func(messageId int) {
		t.Log("onUnsubscribe", messageId)
	})
	client.SetMessageCallback(func(message *Message) {
		t.Log("onMessage", message.Topic(), string(message.Data()))
	})
	client.SetPublishCallback(func(messageId int) {
		t.Log("onPublish", messageId)
	})

	if err := client.Connect(TEST_SERVER, TEST_PORT_PLAINTEXT, 60, false); err != nil {
		t.Error(err)
	} else if err := client.LoopStart(); err != nil {
		t.Error(err)
	} else if id, err := client.Publish("mosquitto/test", []byte("hello, world"), 0, true); err != nil {
		t.Error(err)
	} else {
		t.Log("Publish", id)
	}
	time.Sleep(time.Second * 5)
}
