package stdout

import (
	"fmt"
	"github.com/trafero/tstack/client/mqtt"
	"time"
)

type Stdout struct{}

func New() (c *Stdout, err error) {
	c = &Stdout{}
	return c, nil
}

func (c *Stdout) ControlMessageHandler(msg mqtt.Message) {
	// log.Printf("Received topic: %s message: %s\n", msg.Topic, msg.Payload)
	fmt.Printf("\"%s\",\"%s\",\"%s\"\n", time.Now().Format("02/01/2006 15:04:05"), msg.Topic, string(msg.Payload))
}
