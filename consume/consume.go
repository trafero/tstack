package consume

import (
	"github.com/trafero/tstack/client/mqtt"
)

type Consume interface {
	ControlMessageHandler(msg mqtt.Message)
}
