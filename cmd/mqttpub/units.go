package main

import (
	"fmt"
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	app "github.com/djthorpe/gopi/v2/app"

	// Units
	_ "github.com/djthorpe/gopi/v2/unit/bus"
	_ "github.com/djthorpe/gopi/v2/unit/logger"
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
	if app, err := app.NewCommandLineTool(Main, Events, "mosquitto"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		// Set flags
		app.Flags().FlagString("topic", "", "MQTT Topic")
		app.Flags().FlagInt("qos", 1, "MQTT Quality of service")

		// Run and exit
		os.Exit(app.Run())
	}
}
