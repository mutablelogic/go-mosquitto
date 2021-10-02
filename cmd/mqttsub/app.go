package main

import (
	"context"
	"fmt"
	"time"

	// Packages

	"github.com/djthorpe/go-mosquitto/pkg/mosquitto"
	// Namespace imports
	//. "github.com/djthorpe/go-mosquitto"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type App struct {
	*mosquitto.Client
	topics []string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewApp(ctx context.Context, host string, topics []string) (*App, error) {
	app := new(App)
	app.topics = topics

	// Connect to broker
	if client, err := mosquitto.New(ctx, *flagHost, func(evt *mosquitto.Event) {
		app.ProcessEvent(evt)
		//app.ch <- evt
	}); err != nil {
		return nil, err
	} else {
		app.Client = client
	}

	// Return success
	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	// Subscribe to topics
	for _, topic := range app.topics {
		if _, err := app.Subscribe(topic, *flagQos); err != nil {
			return err
		}
	}

	// Process events until cancel
	for {
		select {
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return app.Close(ctx)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (app *App) ProcessEvent(evt *mosquitto.Event) {
	fmt.Println(evt)
}
