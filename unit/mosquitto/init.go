package mosquitto

import (
	"time"

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
			app.Flags().FlagString("mqtt.host", "", "MQTT Broke hostname")
			app.Flags().FlagUint("mqtt.port", 0, "MQTT Broker port")
			app.Flags().FlagDuration("mqtt.keepalive", 60*time.Second, "MQTT Broker keepalive")
			return nil
		},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Mosquitto{
				User:      app.Flags().GetString("mqtt.user", gopi.FLAG_NS_DEFAULT),
				Password:  app.Flags().GetString("mqtt.password", gopi.FLAG_NS_DEFAULT),
				Host:      app.Flags().GetString("mqtt.host", gopi.FLAG_NS_DEFAULT),
				Port:      app.Flags().GetUint("mqtt.port", gopi.FLAG_NS_DEFAULT),
				Keepalive: app.Flags().GetDuration("mqtt.keepalive", gopi.FLAG_NS_DEFAULT),
				Bus:       app.Bus(),
			}, app.Log().Clone(Mosquitto{}.Name()))
		},
	})
}
