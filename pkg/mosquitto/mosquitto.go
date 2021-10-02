package mosquitto

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	// Packages
	mosq "github.com/djthorpe/mosquitto/sys/mosquitto"
	multierror "github.com/hashicorp/go-multierror"
	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	//. "github.com/djthorpe/go-mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type EventFunc func(*Event)

type Config struct {
	clientId  string
	user      string
	password  string
	host      string
	port      uint
	keepalive time.Duration
	fn        EventFunc
}

type Client struct {
	client *mosq.Client
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	defaultConfig = Config{
		clientId:  "",
		port:      1883,
		keepalive: 60 * time.Second,
	}
	once sync.Once
)

////////////////////////////////////////////////////////////////////////////////
// CONFIGURATION OPTIONS

// Create a new empty configuration
func NewConfigWithBroker(host string) Config {
	return defaultConfig.WithHost(host)
}

func (c Config) WithClientId(v string) Config {
	c.clientId = v
	return c
}

func (c Config) WithCredentials(user, password string) Config {
	c.user = user
	c.password = password
	return c
}

func (c Config) WithHost(v string) Config {
	// Try host:port version first
	if host, port, err := net.SplitHostPort(v); err == nil {
		if port, err := strconv.ParseUint(port, 0, 16); err == nil {
			c.host = host
			c.port = uint(port)
			return c
		}
	}
	// fallback to interpreting as host only
	c.host = v

	// return config
	return c
}

func (c Config) WithKeepalive(d time.Duration) Config {
	c.keepalive = d
	return c
}

func (c Config) WithCallback(fn EventFunc) Config {
	c.fn = fn
	return c
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewWithConfig(cfg Config) (*Client, error) {
	c := new(Client)

	// Initialize
	var result error
	once.Do(func() {
		if err := mosq.Init(); err != nil {
			result = multierror.Append(result, err)
		}
	})

	// Create a new client
	if result != nil {
		return nil, result
	} else if client, err := mosq.New(cfg.clientId, true, 0); err != nil {
		return nil, err
	} else {
		c.client = client
	}

	// Set credentials
	if cfg.user != "" {
		if err := c.client.SetCredentials(cfg.user, cfg.password); err != nil {
			c.client.Destroy()
			return nil, err
		}
	}

	// Set callback
	if cfg.fn != nil {
		c.client.SetSubscribeCallback(func(userInfo uintptr, id int, qos []int) {
			cfg.fn(NewSubscribe(id))
		})
		c.client.SetUnsubscribeCallback(func(userInfo uintptr, id int) {
			cfg.fn(NewUnsubscribe(id))
		})
		c.client.SetPublishCallback(func(userInfo uintptr, id int) {
			cfg.fn(NewPublish(id))
		})
		c.client.SetMessageCallback(func(userInfo uintptr, message *mosq.Message) {
			// We make a copy of the data
			// as this is invalidated after the callback ends
			data := make([]byte, len(message.Data()))
			copy(data, message.Data())
			// Emit
			cfg.fn(NewMessage(message.Id(), message.Topic(), data))
		})
	}

	// Perform connection, start loop
	if err := c.client.LoopStart(); err != nil {
		c.client.Destroy()
		return nil, err
	} else if err := c.client.Connect(cfg.host, int(cfg.port), int(cfg.keepalive.Seconds()), false); err != nil {
		c.client.LoopStop(true)
		c.client.Destroy()
		return nil, err
	}

	// Return success
	return c, nil
}

func (c *Client) Close() error {
	var result error

	// Disconnect client
	if err := c.client.Disconnect(); err != nil {
		result = multierror.Append(result, err)
	}

	// Stop loop
	if err := c.client.LoopStop(false); err != nil {
		result = multierror.Append(result, err)
	}

	// Destroy client
	if err := c.client.Destroy(); err != nil {
		result = multierror.Append(result, err)
	}

	// Cleanup
	if err := mosq.Cleanup(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

/*
func (c *Client) Connect(host string, port uint, opts ...Opt) error {
	// Process options
	flags := MOSQ_FLAG_EVENT_ALL
	keepalive_secs := int(60)
	for _, opt := range opts {
		switch opt.Type {
		case MOSQ_OPTION_FLAGS:
			flags = opt.Flags
		case MOSQ_OPTION_KEEPALIVE:
			keepalive_secs = opt.Int
		default:
			return ErrBadParameter.With(opt.Type)
		}
	}

	// Set flags
	if flags&MOSQ_FLAG_EVENT_CONNECT == MOSQ_FLAG_EVENT_CONNECT {
		this.client.SetConnectCallback(func(userInfo uintptr, rc int) {
			this.bus.Emit(NewConnect(this, rc))
		})
	} else {
		this.client.SetConnectCallback(nil)
	}
	if flags&MOSQ_FLAG_EVENT_DISCONNECT == MOSQ_FLAG_EVENT_DISCONNECT {
		this.client.SetDisconnectCallback(func(userInfo uintptr, rc int) {
			this.bus.Emit(NewDisconnect(this, rc))
		})
	} else {
		this.client.SetDisconnectCallback(nil)
	}
	if flags&MOSQ_FLAG_EVENT_SUBSCRIBE == MOSQ_FLAG_EVENT_SUBSCRIBE {
		this.client.SetSubscribeCallback(func(userInfo uintptr, id int, qos []int) {
			this.bus.Emit(NewSubscribe(this, id))
		})
	} else {
		this.client.SetSubscribeCallback(nil)
	}
	if flags&MOSQ_FLAG_EVENT_UNSUBSCRIBE == MOSQ_FLAG_EVENT_UNSUBSCRIBE {
		this.client.SetUnsubscribeCallback(func(userInfo uintptr, id int) {
			this.bus.Emit(NewUnsubscribe(this, id))
		})
	} else {
		this.client.SetUnsubscribeCallback(nil)
	}
	if flags&MOSQ_FLAG_EVENT_PUBLISH == MOSQ_FLAG_EVENT_PUBLISH {
		this.client.SetPublishCallback(func(userInfo uintptr, id int) {
			this.bus.Emit(NewPublish(this, id))
		})
	} else {
		this.client.SetPublishCallback(nil)
	}
	if flags&MOSQ_FLAG_EVENT_MESSAGE == MOSQ_FLAG_EVENT_MESSAGE {
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
	if this.Log.IsDebug() || flags&MOSQ_FLAG_EVENT_LOG == MOSQ_FLAG_EVENT_LOG {
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

	// Return success
	return nil
}
*/

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c *Client) Version() string {
	major, minor, revision := mosq.Version()
	return fmt.Sprintf("%d.%d.%d", major, minor, revision)
}

func (c *Client) String() string {
	str := "<client"
	str += fmt.Sprintf(" version=%q", c.Version())
	//	str += fmt.Sprintf(" broker=%v:v", c.host, c.port)
	//	str += fmt.Sprintf(" connected=%v", c.connected)
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// SUBSCRIBE, UNSUBSCRIBE AND PUBLISH
/*
func (c *Client) Subscribe(topics string, opts ...Opt) (int, error) {
	// Check for connection
	if c.connected == false {
		return 0, ErrOutOfOrder.With("Subscribe")
	}

	// Process options
	qos := int(1)
	for _, opt := range opts {
		switch opt.Type {
		case MOSQ_OPTION_QOS:
			qos = opt.Int
		default:
			return 0, ErrBadParameter.With(opt.Type)
		}

	}

	// Perform the subscribe
	if id, err := c.client.Subscribe(topics, qos); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (c *Client) Unsubscribe(topics string) (int, error) {
	if c.connected == false {
		return 0, ErrOutOfOrder.With("Unsubscribe")
	} else if id, err := c.client.Unsubscribe(topics); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (c *Client) Publish(topic string, data []byte, opts ...Opt) (int, error) {
	// Check for connection
	if c.connected == false {
		return 0, ErrOutOfOrder.With("Publish")
	}

	// Process options
	qos := int(1)
	retain := false
	for _, opt := range opts {
		switch opt.Type {
		case MOSQ_OPTION_QOS:
			qos = opt.Int
		case MOSQ_OPTION_RETAIN:
			retain = opt.Bool
		default:
			return 0, ErrBadParameter.With(opt.Type)
		}
	}

	// Perform the publish
	if id, err := c.client.Publish(topic, data, qos, retain); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLISH JSON & INFLUX FORMATS

func (this *Client) PublishJSON(topic string, data interface{}, opts ...Opt) (int, error) {
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
func (this *Client) PublishInflux(topic string, measurement string, fields map[string]interface{}, opts ...Opt) (int, error) {
	// Check parameters
	if len(fields) == 0 {
		return 0, ErrBadParameter.With("PublishInflux")
	}
	if measurement == "" {
		return 0, ErrBadParameter.With("PublishInflux")
	}

	// Process options
	str := strings.TrimSpace(measurement)
	ts := ""
	other := make([]Opt, 0, len(opts))
	for _, opt := range opts {
		switch opt.Type {
		case MOSQ_OPTION_TAG:
			str += "," + opt.String
		case MOSQ_OPTION_TIMESTAMP:
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
			return 0, ErrBadParameter.With(k)
		}
		delim = ","
	}

	return this.Publish(topic, []byte(str+ts), other...)
}
*/
