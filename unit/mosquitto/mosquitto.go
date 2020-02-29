package mosquitto

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
	iface "github.com/djthorpe/mosquitto"
	mosq "github.com/djthorpe/mosquitto/sys/mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Mosquitto struct {
	ClientId string
	User     string
	Password string
	Bus      gopi.Bus
}

type mosquitto struct {
	host      string
	port      uint
	client    *mosq.Client
	connected bool
	bus       gopi.Bus

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
	// Bus
	if config.Bus == nil {
		return gopi.ErrBadParameter.WithPrefix("bus")
	} else {
		this.bus = config.Bus
	}

	// Initialize
	if err := mosq.Init(); err != nil {
		return err
	} else if client, err := mosq.New(config.ClientId, true, 0); err != nil {
		return fmt.Errorf("New: %w", err)
	} else {
		this.client = client
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

func (this *mosquitto) Connect(host string, port uint, opts ...iface.Opt) error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Process options
	flags := iface.MOSQ_FLAG_EVENT_ALL
	keepalive_secs := int(60)
	for _, opt := range opts {
		switch opt.Type {
		case iface.MOSQ_OPTION_FLAGS:
			flags = opt.Flags
		case iface.MOSQ_OPTION_KEEPALIVE:
			keepalive_secs = opt.Int
		default:
			return gopi.ErrBadParameter.WithPrefix(fmt.Sprint(opt.Type))
		}
	}

	// Set flags
	if flags&iface.MOSQ_FLAG_EVENT_CONNECT == iface.MOSQ_FLAG_EVENT_CONNECT {
		this.client.SetConnectCallback(func(userInfo uintptr, rc int) {
			this.bus.Emit(NewConnect(this, rc))
		})
	} else {
		this.client.SetConnectCallback(nil)
	}
	if flags&iface.MOSQ_FLAG_EVENT_DISCONNECT == iface.MOSQ_FLAG_EVENT_DISCONNECT {
		this.client.SetDisconnectCallback(func(userInfo uintptr, rc int) {
			this.bus.Emit(NewDisconnect(this, rc))
		})
	} else {
		this.client.SetDisconnectCallback(nil)
	}
	if flags&iface.MOSQ_FLAG_EVENT_SUBSCRIBE == iface.MOSQ_FLAG_EVENT_SUBSCRIBE {
		this.client.SetSubscribeCallback(func(userInfo uintptr, id int, qos []int) {
			this.bus.Emit(NewSubscribe(this, id))
		})
	} else {
		this.client.SetSubscribeCallback(nil)
	}
	if flags&iface.MOSQ_FLAG_EVENT_UNSUBSCRIBE == iface.MOSQ_FLAG_EVENT_UNSUBSCRIBE {
		this.client.SetUnsubscribeCallback(func(userInfo uintptr, id int) {
			this.bus.Emit(NewUnsubscribe(this, id))
		})
	} else {
		this.client.SetUnsubscribeCallback(nil)
	}
	if flags&iface.MOSQ_FLAG_EVENT_PUBLISH == iface.MOSQ_FLAG_EVENT_PUBLISH {
		this.client.SetPublishCallback(func(userInfo uintptr, id int) {
			this.bus.Emit(NewPublish(this, id))
		})
	} else {
		this.client.SetPublishCallback(nil)
	}
	if flags&iface.MOSQ_FLAG_EVENT_MESSAGE == iface.MOSQ_FLAG_EVENT_MESSAGE {
		this.client.SetMessageCallback(func(userInfo uintptr, message *mosq.Message) {
			// We make a copy of the data
			// as this is invalidated after the callback ends
			data := make([]byte, len(message.Data()))
			copy(data, message.Data())
			// Emit
			this.bus.Emit(NewMessage(this, message.Id(), message.Topic(), data))
		})
	} else {
		this.client.SetMessageCallback(nil)
	}
	if this.Log.IsDebug() || flags&iface.MOSQ_FLAG_EVENT_LOG == iface.MOSQ_FLAG_EVENT_LOG {
		this.client.SetLogCallback(func(userInfo uintptr, level mosq.Level, str string) {
			if level&mosq.MOSQ_LOG_DEBUG > 0 {
				this.Log.Debug(level, str)
			} else if level&mosq.MOSQ_LOG_ERR > 0 {
				this.Log.Error(fmt.Errorf("%v %v", level, str))
			} else {
				this.Log.Info(level, str)
			}
		})
	} else {
		this.client.SetLogCallback(nil)
	}

	// Perform connection, start loop
	if err := this.client.LoopStart(); err != nil {
		return err
	} else if err := this.client.Connect(host, int(port), keepalive_secs, false); err != nil {
		this.client.LoopStop(true)
		return err
	} else {
		this.host = host
		this.port = port
		this.connected = true
	}

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
	if err := this.client.Disconnect(); err != nil {
		return err
	} else if err := this.client.LoopStop(false); err != nil {
		return err
	} else {
		this.connected = false
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *mosquitto) Version() string {
	major, minor, revision := mosq.Version()
	return fmt.Sprintf("%d.%d.%d", major, minor, revision)
}

func (this *mosquitto) String() string {
	str := "<mosq.Client"
	str += " version=" + strconv.Quote(this.Version())
	str += " broker=" + fmt.Sprintf("%v:%v", this.host, this.port)
	str += " connected=" + fmt.Sprint(this.connected)
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// SUBSCRIBE, UNSUBSCRIBE AND PUBLISH

func (this *mosquitto) Subscribe(topics string, opts ...iface.Opt) (int, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Check for connection
	if this.connected == false {
		return 0, gopi.ErrOutOfOrder
	}

	// Process options
	qos := int(1)
	for _, opt := range opts {
		switch opt.Type {
		case iface.MOSQ_OPTION_QOS:
			qos = opt.Int
		default:
			return 0, gopi.ErrBadParameter.WithPrefix(fmt.Sprint(opt.Type))
		}

	}

	// Perform the subscribe
	if id, err := this.client.Subscribe(topics, qos); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (this *mosquitto) Unsubscribe(topics string) (int, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if this.connected == false {
		return 0, gopi.ErrOutOfOrder
	} else if id, err := this.client.Unsubscribe(topics); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (this *mosquitto) Publish(topic string, data []byte, opts ...iface.Opt) (int, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Check for connection
	if this.connected == false {
		return 0, gopi.ErrOutOfOrder
	}

	// Process options
	qos := int(1)
	retain := false
	for _, opt := range opts {
		switch opt.Type {
		case iface.MOSQ_OPTION_QOS:
			qos = opt.Int
		case iface.MOSQ_OPTION_RETAIN:
			retain = opt.Bool
		default:
			return 0, gopi.ErrBadParameter.WithPrefix(fmt.Sprint(opt.Type))
		}

	}

	// Perform the publish
	if id, err := this.client.Publish(topic, data, qos, retain); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLISH JSON & INFLUX FORMATS

func (this *mosquitto) PublishJSON(topic string, data interface{}, opts ...iface.Opt) (int, error) {
	if json, err := json.Marshal(data); err != nil {
		return 0, err
	} else {
		return this.Publish(topic, json, opts...)
	}
}

// Influx line protocol
// https://docs.influxdata.com/influxdb/v1.7/write_protocols/line_protocol_tutorial/
// Include one or more OptTag(name,value) for tags
// Include one OptTimestamp(time.Time) to set timestamp
func (this *mosquitto) PublishInflux(topic string, measurement string, fields map[string]interface{}, opts ...iface.Opt) (int, error) {
	// Check parameters
	if len(fields) == 0 {
		return 0, gopi.ErrBadParameter.WithPrefix("fields")
	}
	if measurement == "" {
		return 0, gopi.ErrBadParameter.WithPrefix("measurement")
	}

	// Process options
	str := strings.TrimSpace(measurement)
	ts := ""
	other := make([]iface.Opt, 0, len(opts))
	for _, opt := range opts {
		switch opt.Type {
		case iface.MOSQ_OPTION_TAG:
			str += "," + opt.String
		case iface.MOSQ_OPTION_TIMESTAMP:
			ts = " " + fmt.Sprint(opt.Timestamp.UnixNano())
		default:
			other = append(other, opt)
		}
	}

	// Process fields
	delim := " "
	for k, v := range fields {
		switch v.(type) {
		case float32, float64, bool, int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			str += delim + fmt.Sprintf("%v=%v", strings.TrimSpace(k), v)
		case string:
			str += delim + fmt.Sprintf("%v=%v", strings.TrimSpace(k), strconv.Quote(v.(string)))
		default:
			return 0, gopi.ErrBadParameter.WithPrefix(k)
		}
		delim = ","
	}

	return this.Publish(topic, []byte(str+ts), other...)
}

////////////////////////////////////////////////////////////////////////////////
// WAIT FOR

// Wait for a specific request-id and return the event
func (this *mosquitto) WaitFor(context.Context, int) (iface.Event, error) {
	return nil, gopi.ErrNotImplemented
}
