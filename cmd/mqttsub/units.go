package main

import (
	"fmt"
	"os"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	app "github.com/djthorpe/gopi/v2/app"

	// Units
	_ "github.com/djthorpe/gopi/v2/unit/bus"
	_ "github.com/djthorpe/gopi/v2/unit/logger"
	_ "github.com/djthorpe/gopi/v2/unit/mdns"
	_ "github.com/djthorpe/mosquitto/unit/mosquitto"
)

var (
	Events = []gopi.EventHandler{
		gopi.EventHandler{Name: "mosquitto.Event", Handler: EventHandler},
	}
)

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP

func main() {
	if app, err := app.NewCommandLineTool(Main, Events, "mosquitto", "discovery"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		app.Flags().FlagString("broker", "", "Host and port of the broker")
		app.Flags().FlagDuration("timeout", time.Second, "Timeout for broker discovery")

		// Run and exit
		os.Exit(app.Run())
	}
}
