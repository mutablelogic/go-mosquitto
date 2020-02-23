package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	mosquitto "github.com/djthorpe/mosquitto"
)

var (
	Format = "%-10v %-40v %-28v\n"
	Header sync.Once
)

////////////////////////////////////////////////////////////////////////////////

func Main(app gopi.App, args []string) error {
	// Check args
	if len(args) == 0 {
		return gopi.ErrHelp
	}

	// Connect client
	client := app.UnitInstance("mosquitto").(mosquitto.Client)
	if err := client.Connect(); err != nil {
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
		fmt.Printf(Format, strings.Repeat("-", 10), strings.Repeat("-", 40), strings.Repeat("-", 28))
	})
	message := strings.TrimPrefix(fmt.Sprint(evt.Type()), "MOSQ_FLAG_EVENT_")
	topic := TruncateString(evt.Topic(), 40)
	data := "<nil>"
	if len(evt.Data()) > 0 {
		data = TruncateString(strconv.Quote(string(evt.Data())), 28)
	}
	fmt.Printf(Format, message, topic, data)
}

func TruncateString(value string, l int) string {
	if len(value) > l {
		value = value[0:l-4] + "..."
	}
	return value
}
