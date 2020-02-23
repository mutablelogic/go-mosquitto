package main

import (
	"context"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	mosquitto "github.com/djthorpe/mosquitto"
)

var (
	wg sync.WaitGroup
)

////////////////////////////////////////////////////////////////////////////////

func Main(app gopi.App, args []string) error {
	client := app.UnitInstance("mosquitto").(mosquitto.Client)

	// Connect to client
	if topic := app.Flags().GetString("topic", gopi.FLAG_NS_DEFAULT); topic == "" {
		return gopi.ErrBadParameter.WithPrefix("topic")
	} else if len(args) == 0 {
		return gopi.ErrHelp
	} else if err := client.Connect(); err != nil {
		return err
	} else {
		// Wait for connect
		wg.Add(1)

		// Wait for all publish acknowledgements
		wg.Wait()
	}

	// Return success
	return nil
}

func EventHandler(_ context.Context, app gopi.App, evt_ gopi.Event) {
	evt := evt_.(mosquitto.Event)
	client := app.UnitInstance("mosquitto").(mosquitto.Client)
	topic := app.Flags().GetString("topic", gopi.FLAG_NS_DEFAULT)
	qos := app.Flags().GetInt("qos", gopi.FLAG_NS_DEFAULT)

	if evt.Type() == mosquitto.MOSQ_FLAG_EVENT_CONNECT && evt.ReturnCode() == 0 {
		for _, arg := range app.Flags().Args() {
			if id, err := client.Publish(topic, []byte(arg), mosquitto.OptQOS(qos)); err != nil {
				app.Log().Error(err)
			} else {
				app.Log().Info("PUBLISH:", id)
				wg.Add(1)
			}
		}
		wg.Done()
	}

	if evt.Type() == mosquitto.MOSQ_FLAG_EVENT_PUBLISH {
		app.Log().Info("PUBACK:", evt.Id())
		wg.Done()
	}
}
