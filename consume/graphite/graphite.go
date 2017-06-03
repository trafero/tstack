package graphite

import (
	"github.com/marpaia/graphite-golang"
	"github.com/trafero/tstack/client/mqtt"
	"github.com/trafero/tstack/tstackutil"
)

type Graphite struct {
	graphitehost string
	graphiteport int
	graphite     *graphite.Graphite
}

const (
	clientid = "consumer"
)

func New(
	graphitehost string,
	graphiteport int,
) (c *Graphite, err error) {

	c = &Graphite{
		graphitehost: graphitehost,
		graphiteport: graphiteport,
	}

	tstackutil.WaitForTcp(graphitehost, graphiteport)
	c.graphite, err = graphite.NewGraphite(graphitehost, graphiteport)

	return c, err

}

func (c *Graphite) ControlMessageHandler(msg mqtt.Message) {
	// log.Printf("Received topic: %s message: %s", msg.Topic, msg.Payload)
	c.graphite.SimpleSend(msg.Topic, msg.Payload)
}
