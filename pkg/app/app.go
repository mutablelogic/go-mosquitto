package app

import (
	"context"
	"fmt"

	// Packages
	"github.com/djthorpe/go-mosquitto/pkg/mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type App struct {
	*mosquitto.Client
	qos int
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewApp(ctx context.Context, host string, qos int) (*App, error) {
	app := new(App)
	app.qos = qos

	// Connect to broker
	if client, err := mosquitto.New(ctx, host, func(evt *mosquitto.Event) {
		app.ProcessEvent(evt)
	}); err != nil {
		return nil, err
	} else {
		app.Client = client
	}

	// Return success
	return app, nil
}

// Subscribe to topics, wait until cancel then close app
func (app *App) Run(ctx context.Context, topics ...string) error {
	for _, topic := range topics {
		if _, err := app.Subscribe(topic, mosquitto.OptQoS(app.qos)); err != nil {
			return err
		}
	}

	<-ctx.Done()
	return app.Close()
}

// Publish data to topic
func (app *App) Publish(topic, data string) error {
	if _, err := app.Client.Publish(topic, []byte(data), mosquitto.OptQoS(app.qos)); err != nil {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (app *App) ProcessEvent(evt *mosquitto.Event) {
	fmt.Println(evt)
}
