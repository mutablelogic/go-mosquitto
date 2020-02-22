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
	Format = "%-8v %-40v %-30v\n"
	Header sync.Once
)

////////////////////////////////////////////////////////////////////////////////

func Main(app gopi.App, args []string) error {
	// Connect client
	client := app.UnitInstance("mosquitto").(mosquitto.Client)
	if err := client.Connect(mosquitto.MOSQ_FLAG_EVENT_MESSAGE); err != nil {
		return err
	}

	// Subscribe to topics
	for _, topic := range args {
		if _, err := client.Subscribe(topic, 0); err != nil {
			return err
		}
	}

	// Wait for CTRL+C
	fmt.Println("Press CTRL+C to exit")
	app.WaitForSignal(context.Background(), os.Interrupt)

	// Return success
	return nil
}

func EventHandler(_ context.Context, app gopi.App, evt_ gopi.Event) {
	evt := evt_.(mosquitto.Event)
	Header.Do(func() {
		fmt.Printf(Format, "TYPE", "TOPIC", "DATA")
		fmt.Printf(Format, strings.Repeat("-", 8), strings.Repeat("-", 40), strings.Repeat("-", 30))
	})
	message := strings.TrimPrefix(fmt.Sprint(evt.Type()), "MOSQ_FLAG_EVENT_")
	topic := TruncateString(evt.Topic(), 40)
	data := TruncateString(strconv.Quote(string(evt.Data())), 30)
	fmt.Printf(Format, message, topic, data)
}

func TruncateString(value string, l int) string {
	if len(value) > l {
		value = value[0:l-4] + "..."
	}
	return value
}
