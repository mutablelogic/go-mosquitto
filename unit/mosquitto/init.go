package mosquitto

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
)

func init() {
	gopi.UnitRegister(gopi.UnitConfig{
		Name:     Mosquitto{}.Name(),
		Requires: []string{"bus"},
		Config: func(app gopi.App) error {
			app.Flags().FlagString("mqtt.user", "", "Username")
			app.Flags().FlagString("mqtt.password", "", "Password")
			app.Flags().FlagString("mqtt.broker", "", "Broker host:port")
			return nil
		},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Mosquitto{
				User:     app.Flags().GetString("mqtt.user", gopi.FLAG_NS_DEFAULT),
				Password: app.Flags().GetString("mqtt.password", gopi.FLAG_NS_DEFAULT),
				Broker:   app.Flags().GetString("mqtt.broker", gopi.FLAG_NS_DEFAULT),
				Bus:      app.Bus(),
			}, app.Log().Clone(Mosquitto{}.Name()))
		},
	})
}
