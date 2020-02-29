package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	mosquitto "github.com/djthorpe/mosquitto"
)

var (
	Format = "%-10v %-40v %-40v\n"
	Header sync.Once
)

////////////////////////////////////////////////////////////////////////////////

const (
	MOSQUITTO_SERVICE = "_mosquitto._tcp"
)

var (
	reHostPort = regexp.MustCompile("^([^\\:]+)\\:(\\d+)$")
)

////////////////////////////////////////////////////////////////////////////////

func DiscoverBroker(app gopi.App) (string, uint, error) {
	discovery := app.UnitInstance("discovery").(gopi.RPCServiceDiscovery)
	timeout := app.Flags().GetDuration("timeout", gopi.FLAG_NS_DEFAULT)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if services, err := discovery.Lookup(ctx, MOSQUITTO_SERVICE); err != nil {
		return "", 0, err
	} else if len(services) == 0 {
		return "", 0, gopi.ErrNotFound.WithPrefix(MOSQUITTO_SERVICE)
	} else if len(services) > 1 {
		return "", 0, gopi.ErrDuplicateItem.WithPrefix(MOSQUITTO_SERVICE)
	} else {
		return services[0].Host, uint(services[0].Port), nil
	}
}

func GetHostPort(app gopi.App) (string, uint, error) {
	hostport := app.Flags().GetString("broker", gopi.FLAG_NS_DEFAULT)
	if hostport == "" {
		return DiscoverBroker(app)
	} else if reHostPort.MatchString(hostport) == false {
		hostport = fmt.Sprintf("%v:%v", hostport, 0)
	}
	if host, port, err := net.SplitHostPort(hostport); err != nil {
		return "", 0, gopi.ErrBadParameter.WithPrefix("-broker")
	} else if port_, err := strconv.ParseUint(port, 10, 32); err != nil {
		return "", 0, gopi.ErrBadParameter.WithPrefix("-broker")
	} else {
		return host, uint(port_), nil
	}
}

func Main(app gopi.App, args []string) error {
	client := app.UnitInstance("mosquitto").(mosquitto.Client)

	// Check args
	if len(args) == 0 {
		return gopi.ErrHelp
	}

	// If there is no -broker flag then use discovery
	if addr, port, err := GetHostPort(app); err != nil {
		return err
	} else if err := client.Connect(addr, port); err != nil {
		return err
	}

	// Wait for CTRL+C
	fmt.Println("Press CTRL+C to exit")
	app.WaitForSignal(context.Background(), os.Interrupt)

	// Return success
	return nil
}

func EventHandler(_ context.Context, app gopi.App, evt_ gopi.Event) {
	evt := evt_.(mosquitto.Event)
	client := app.UnitInstance("mosquitto").(mosquitto.Client)

	// Subscribe to topics
	if evt.Type() == mosquitto.MOSQ_FLAG_EVENT_CONNECT && evt.ReturnCode() == 0 {
		// Subscribe to topics
		for _, topic := range app.Flags().Args() {
			if _, err := client.Subscribe(topic); err != nil {
				app.Log().Error(err)
			}
		}
	}

	Header.Do(func() {
		fmt.Printf(Format, "TYPE", "TOPIC", "DATA")
		fmt.Printf(Format, strings.Repeat("-", 10), strings.Repeat("-", 40), strings.Repeat("-", 40))
	})
	message := strings.TrimPrefix(fmt.Sprint(evt.Type()), "MOSQ_FLAG_EVENT_")
	topic := TruncateString(evt.Topic(), 40)
	data := "<nil>"
	if len(evt.Data()) > 0 {
		data = TruncateString(strconv.Quote(string(evt.Data())), 40)
	}
	fmt.Printf(Format, message, topic, data)
}

func TruncateString(value string, l int) string {
	if len(value) > l {
		value = value[0:l-4] + "..."
	}
	return value
}
