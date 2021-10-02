package mosquitto

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	// Packages
	mosq "github.com/djthorpe/go-mosquitto/sys/mosquitto"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	sync.WaitGroup
	client     *mosq.ClientEx
	ch         chan *Event
	disconnect bool
}

type EventFunc func(*Event)
type TraceFunc func(string)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	once = new(sync.Once)
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New client connection to a broker, which will callback on events
func New(ctx context.Context, host string, callback EventFunc) (*Client, error) {
	return NewWithConfig(ctx, defaultConfig.WithHost(host).WithCallback(callback))
}

// New client connection with additional configuration
func NewWithConfig(ctx context.Context, cfg Config) (*Client, error) {
	c := new(Client)

	// Initialize once
	var result error
	once.Do(func() {
		if err := mosq.Init(); err != nil {
			result = multierror.Append(result, err)
		}
		runtime.SetFinalizer(&once, func() {
			mosq.Cleanup()
		})
	})

	// Create a new client
	if result != nil {
		return nil, result
	} else if client, err := mosq.NewEx(cfg.clientId, true); err != nil {
		return nil, err
	} else {
		c.client = client
		c.ch = make(chan *Event)
	}

	// Set credentials
	if cfg.user != "" {
		if err := c.client.SetCredentials(cfg.user, cfg.password); err != nil {
			c.client.Destroy()
			return nil, err
		}
	}

	// Set TLS
	if cfg.capath != "" {
		if err := c.client.SetTLS(cfg.capath, cfg.certpath, cfg.keypath); err != nil {
			c.client.Destroy()
			return nil, err
		}
		if err := c.client.SetTLSInsecure(!cfg.certverify); err != nil {
			c.client.Destroy()
			return nil, err
		}
		if cfg.port == 0 {
			cfg.port = mosq.MOSQ_DEFAULT_SECURE_PORT
		}
	} else {
		if cfg.port == 0 {
			cfg.port = mosq.MOSQ_DEFAULT_PORT
		}
	}

	// Always set connect and disconnect callbacks
	c.client.SetConnectCallback(func(err mosq.Error) {
		err_ := toError(err)
		select {
		case c.ch <- NewConnect(err_):
			break
		default:
			break
		}
		if cfg.fn != nil {
			cfg.fn(NewConnect(err_))
		}
	})
	c.client.SetDisconnectCallback(func(err mosq.Error) {
		err_ := toError(err)
		select {
		case c.ch <- NewDisconnect(err_):
			break
		default:
			break
		}
		if cfg.fn != nil {
			cfg.fn(NewDisconnect(err_))
		}
	})

	// Set event callbacks
	if cfg.fn != nil {
		c.client.SetSubscribeCallback(func(id int, qos []int) {
			cfg.fn(NewSubscribe(id))
		})
		c.client.SetUnsubscribeCallback(func(id int) {
			cfg.fn(NewUnsubscribe(id))
		})
		c.client.SetPublishCallback(func(id int) {
			cfg.fn(NewPublish(id))
		})
		c.client.SetMessageCallback(func(message *mosq.Message) {
			// We make a copy of the data
			// as this is invalidated after the callback ends
			data := make([]byte, len(message.Data()))
			copy(data, message.Data())
			// Emit
			cfg.fn(NewMessage(message.Id(), message.Topic(), data))
		})
	}

	// Set trace callback
	if cfg.trace != nil {
		c.client.SetLogCallback(func(level mosq.Level, message string) {
			cfg.trace(message)
		})
	}

	// Perform connection, start loop
	if err := c.client.Connect(cfg.host, int(cfg.port), int(cfg.keepalive.Seconds()), false); err != nil {
		c.client.LoopStop(true)
		c.client.Destroy()
		return nil, err
	}

	// Run the loop in the background
	c.WaitGroup.Add(1)
	go func(delta time.Duration) {
		defer c.WaitGroup.Done()
		for !c.disconnect {
			if err := c.client.Loop(int(delta.Milliseconds())); err != nil {
				break
			}
		}
	}(time.Second)

	// Wait for connection, cancel, or some other unknown issue
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case evt := <-c.ch:
		if evt.Type == MOSQ_FLAG_EVENT_CONNECT && evt.Err == nil {
			return c, nil
		} else if evt.Err != nil {
			return nil, evt.Err
		} else {
			return nil, ErrOutOfOrder.With(evt.Type)
		}
	}
}

func (c *Client) Close() error {
	var result error

	c.disconnect = true
	if err := c.client.Disconnect(); err != nil {
		result = multierror.Append(result, err)
	}

	// Wait for loop to be completed
	c.WaitGroup.Wait()

	// Destroy client
	if err := c.client.Destroy(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c *Client) Version() string {
	major, minor, revision := mosq.Version()
	return fmt.Sprintf("%d.%d.%d", major, minor, revision)
}

func (c *Client) String() string {
	str := "<client"
	str += fmt.Sprintf(" version=%q", c.Version())
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) Subscribe(topics string, opts ...ClientOpt) (int, error) {
	// Apply options
	v := defaultOpts
	for _, opt := range opts {
		opt(&v)
	}
	// Perform the subscribe
	if id, err := c.client.Subscribe(topics, v.qos); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (c *Client) Unsubscribe(topics string) (int, error) {
	if id, err := c.client.Unsubscribe(topics); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (c *Client) Publish(topic string, data []byte, opts ...ClientOpt) (int, error) {
	// Apply options
	v := defaultOpts
	for _, opt := range opts {
		opt(&v)
	}
	// Send message
	if id, err := c.client.Publish(topic, data, v.qos, v.retain); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLISH JSON & INFLUX FORMATS

func (c *Client) PublishJSON(topic string, data interface{}, opts ...ClientOpt) (int, error) {
	if json, err := json.Marshal(data); err != nil {
		return 0, err
	} else {
		return c.Publish(topic, json, opts...)
	}
}

/*
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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toError(err mosq.Error) error {
	if err == mosq.MOSQ_ERR_SUCCESS {
		return nil
	} else {
		return err
	}
}
