package mosquitto

import (
	// Frameworks
	"github.com/djthorpe/gopi/v2"
)

type Client interface {
	// Connect to MQTT broker
	Connect() error

	gopi.Unit
}
