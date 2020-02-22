package mosquitto

import (
	"fmt"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
	mosq "github.com/djthorpe/mosquitto/sys/mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Mosquitto struct {
	ClientId  string
	User      string
	Password  string
	Host      string
	Port      uint
	Keepalive time.Duration
}

type mosquitto struct {
	host      string
	port      uint
	client    *mosq.Client
	keepalive time.Duration
	connected bool

	base.Unit
	sync.Mutex
	sync.WaitGroup
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (Mosquitto) Name() string { return "mosquitto" }

func (config Mosquitto) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(mosquitto)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	}
	if err := this.Init(config); err != nil {
		return nil, err
	}
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION mosquitto.Client

func (this *mosquitto) Init(config Mosquitto) error {

	// Initialize
	if err := mosq.Init(); err != nil {
		return err
	} else if client, err := mosq.New(config.ClientId, true, uintptr(0)); err != nil {
		return fmt.Errorf("New: %w", err)
	} else {
		this.client = client
	}

	// Check host and port
	if config.Host == "" {
		return gopi.ErrBadParameter.WithPrefix("host")
	} else {
		this.host = config.Host
	}
	if config.Port == 0 {
		this.port = mosq.MOSQ_DEFAULT_PORT
	} else {
		this.port = config.Port
	}

	// Set keep alive
	if config.Keepalive == 0 {
		return gopi.ErrBadParameter.WithPrefix("keepalive")
	} else {
		this.keepalive = config.Keepalive
	}

	// Set credentials
	if config.User != "" {
		if err := this.client.SetCredentials(config.User, config.Password); err != nil {
			return err
		}
	}

	return nil
}

func (this *mosquitto) Close() error {
	// Disconnect client
	if err := this.Disconnect(); err != nil {
		return err
	}

	// Destroy client
	if err := this.client.Destroy(); err != nil {
		return err
	}

	// Cleanup
	if err := mosq.Cleanup(); err != nil {
		return err
	}

	// Return success
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// CONNECT AND DISCONNECT

func (this *mosquitto) Connect() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Perform connection
	if keepalive_secs := int(this.keepalive.Seconds()); keepalive_secs < 1 {
		return gopi.ErrBadParameter.WithPrefix("keepalive")
	} else if err := this.client.Connect(this.host, int(this.port), keepalive_secs, false); err != nil {
		return err
	} else {
		this.connected = true
	}

	// Set connect & disconnect callbacks
	this.client.SetConnectCallback(func(userInfo uintptr, rc, flags int) {
		fmt.Println("CONNECT userInfo=", userInfo, " rc=", rc, " flags=", flags)
	})
	this.client.SetDisconnectCallback(func(userInfo uintptr, rc int) {
		fmt.Println("DISCONNECT userInfo=", userInfo, " rc=", rc)
	})

	// Start the loop in the background
	go func() {
		this.WaitGroup.Add(1)
		this.Log.Debug("->loop_forever")
		this.client.LoopForever(100)
		this.Log.Debug("<-loop_forever")
		this.WaitGroup.Done()
	}()

	// Return success
	return nil
}

func (this *mosquitto) Disconnect() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Check for connection
	if this.connected == false {
		return nil
	}

	// Perform disconnect
	err := this.client.Disconnect()
	this.connected = false
	this.WaitGroup.Wait()

	// Return any error
	return err
}
