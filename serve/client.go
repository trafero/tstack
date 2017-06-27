package serve

import (
	"github.com/gomqtt/packet"
	"github.com/trafero/tstack/auth"
	"io"
	"log"
	"net"
	"sync"
	"time"
	"unsafe"
)

type client struct {
	broker                *Broker
	conn                  net.Conn
	auth                  auth.Auth
	cleanSession          bool
	processedConnect      bool
	clientid              string
	username              string
	rights                string
	will                  *packet.Message
	keepalive             uint16
	encoder               *packet.Encoder
	decoder               *packet.Decoder
	subscriptions         map[string]packet.Subscription // Mapped by topic
	mutex                 *sync.Mutex
	connectionMutex       *sync.Mutex
	packetIDCounter       uint16
	inboundInTransit      map[uint16]packet.Message // QOS 2 messages to be received (and passeed to broker)
	outboundInTransit     map[uint16]packet.Message // QOS 2 messages to be sent
	internalClientCounter int                       // For internal client ids (MQTT-3.1.3-6)

}

func NewClient(a auth.Auth, b *Broker, c net.Conn) *client {
	return &client{
		auth:              a,
		broker:            b,
		conn:              c,
		processedConnect:  false,
		mutex:             &sync.Mutex{},
		connectionMutex:   &sync.Mutex{},
		inboundInTransit:  make(map[uint16]packet.Message),
		outboundInTransit: make(map[uint16]packet.Message),
		subscriptions:     make(map[string]packet.Subscription),
		packetIDCounter:   0,
		keepalive:         0, // in seconds
	}
}

func (c *client) HandleConnection() {
	var err error
	var pkt packet.Packet
	c.decoder = packet.NewDecoder(c.conn)
	c.encoder = packet.NewEncoder(c.conn)
	for {
		pkt, err = c.decoder.Read()
		if err != nil {
			if err == io.EOF {
				log.Println("Connection disconnected")
			}
			c.conn.Close()
			break
		}

		// Connection timeout
		c.setReadDeadline()

		switch pkt := pkt.(type) {
		default:
			log.Println("Unknown MQTT packet received")
			c.conn.Close()
		case *packet.ConnectPacket:
			c.processConnect(pkt)
		case *packet.PublishPacket:
			go c.processPublish(pkt)
		case *packet.SubscribePacket:
			go c.processSubscribe(pkt)
		case *packet.UnsubscribePacket:
			go c.processUnsubscribe(pkt)
		case *packet.PubackPacket:
			go c.processPuback(pkt)
		case *packet.PubcompPacket:
			go c.processComp(pkt)
		case *packet.PubrecPacket:
			go c.processPubrec(pkt)
		case *packet.PubrelPacket:
			go c.processPubrel(pkt)
		case *packet.PingreqPacket:
			c.processPing(pkt)
		case *packet.DisconnectPacket:
			c.processDisconnect(pkt)
		}
	}
	// Send out with last will. Last will set to nill if never set or
	// client send disconnect
	if c.will != nil {
		c.broker.deliver(c.will)
	}
	
	// Remove the client from the list
	if c.keepalive != 1 {
		c.broker.RemoveClient(c)
	}
}

/*
 * CONNECT – Client requests a connection to a Server (3.1)
 */
func (c *client) processConnect(pkt *packet.ConnectPacket) {
	if c.processedConnect {
		log.Println("Connect packet received for a second time on same connection")
		// No acknowledgement, just disconnect
		c.conn.Close()
	}
	c.processedConnect = true
	if pkt.Version != 4 {
		c.writeConnack(packet.ErrInvalidProtocolVersion)
		log.Println("Unsupported MQTT version")
		c.conn.Close()
		return
	}
	if c.auth.Authenticate(pkt.Username, pkt.Password) == false {
		c.writeConnack(packet.ErrNotAuthorized)
		log.Printf("User %s could not be authenticated", pkt.Username)
		c.conn.Close()
		return
	}
	// MQTT-3.1.3-8
	if pkt.ClientID == "" && pkt.CleanSession == false {
		c.writeConnack(packet.ErrIdentifierRejected)
		c.conn.Close()
		return
	}
	// MQTT-3.1.3-6
	if pkt.ClientID == "" {
		// Must be a clean session, which we know already from above
		pkt.ClientID = c.newInternalClientID()
	}

	// TODO check Clinet ID is not already in use
	c.clientid = pkt.ClientID

	c.cleanSession = pkt.CleanSession
	c.username = pkt.Username
	c.rights = c.auth.Rights(c.username)

	if pkt.Will != nil && !matches(c.rights, pkt.Will.Topic) {
		log.Println("Client not authorized to write this will")
	} else {
		c.will = pkt.Will // May be nil but that is ok
	}
	c.keepalive = pkt.KeepAlive
	c.setReadDeadline()
	c.broker.AddClient(c)
	c.writeConnack(packet.ConnectionAccepted)
}

/*
* CONNACK – Acknowledge connection request (3.2)
 */
func (c *client) writeConnack(code packet.ConnackCode) {
	connack := packet.NewConnackPacket()
	connack.SessionPresent = false
	connack.ReturnCode = code
	c.sendPacket(connack)

	// Now we are connected, check if there's any unfinished business
	// Unfinished packets
	for packetID, msg := range c.outboundInTransit {
		c.resend(packetID, &msg)
	}
}

/*
 * PUBLISH – Publish message (3.3)
 */
func (c *client) processPublish(pkt *packet.PublishPacket) {
	if !matches(c.rights, pkt.Message.Topic) {
		// TODO send code back?
		log.Printf("Not authorized to publish to topic %s", pkt.Message.Topic)
		// Give them a hint
		c.conn.Close()
	} else {
		// QOS 1

		switch pkt.Message.QOS {

		case packet.QOSAtMostOnce:
			// QOS 0
			c.broker.deliver(&pkt.Message)

		case packet.QOSAtLeastOnce:
			// QOS 1
			c.broker.deliver(&pkt.Message)
			p := packet.NewPubackPacket()
			p.PacketID = pkt.PacketID
			c.sendPacket(p)

		case packet.QOSExactlyOnce:
			// QOS 2
			c.mutex.Lock()
			c.inboundInTransit[pkt.PacketID] = pkt.Message
			c.mutex.Unlock()
			p := packet.NewPubrecPacket()
			p.PacketID = pkt.PacketID
			c.sendPacket(p)
			// Send it back to the main switch for a Pubrel

		default:
			// Unknown QOS
			log.Printf("Unknown QOS level")
			c.conn.Close()
		}
	}
}

/*
 *  PUBACK – Publish acknowledgement (3.4)
 */
func (c *client) processPuback(pkt *packet.PubackPacket) {
	delete(c.outboundInTransit, pkt.PacketID)
}

/*
 * PUBREC – Publish received (QoS 2 publish received, part 1) (3.5)
 */
func (c *client) processPubrec(pkt *packet.PubrecPacket) {
	// Only send resonse if we have the message
	if _, ok := c.outboundInTransit[pkt.PacketID]; !ok {
		log.Println("Pubrec for a message that I do not have")
		c.conn.Close()
		return
	}
	p := packet.NewPubrelPacket()
	p.PacketID = pkt.PacketID
	c.sendPacket(p)
}

/*
 * PUBREL – Publish release (QoS 2 publish received, part 2) (3.6)
 */
func (c *client) processPubrel(pkt *packet.PubrelPacket) {
	msg := c.inboundInTransit[pkt.PacketID]

	if unsafe.Sizeof(msg) != 0 {
		// msg is not an empty stuct
		c.broker.deliver(&msg)
		delete(c.inboundInTransit, pkt.PacketID)
		p := packet.NewPubcompPacket()
		p.PacketID = pkt.PacketID
		c.sendPacket(p)
	}

}

/*
 * PUBCOMP – Publish complete (QoS 2 publish received, part 3) (3.7)
 */
func (c *client) processComp(pkt *packet.PubcompPacket) {
	delete(c.outboundInTransit, pkt.PacketID)
}

/*
 * SUBSCRIBE - Subscribe to topics (3.8)
 */
func (c *client) processSubscribe(pkt *packet.SubscribePacket) {
	suback := packet.NewSubackPacket()
	suback.PacketID = pkt.PacketID

	for _, s := range pkt.Subscriptions {
		if !matches(c.rights, s.Topic) {
			log.Printf("Not authorized to subscribe to topic %s", s.Topic)
			suback.ReturnCodes = append(suback.ReturnCodes, 0x80) // sec 3.9.3 of spec
		} else {
			c.mutex.Lock()
			c.subscriptions[s.Topic] = s
			c.mutex.Unlock()
			suback.ReturnCodes = append(suback.ReturnCodes, s.QOS)
			// Send any retained messages for this subscription
			c.sendRetained(s.Topic, s.QOS)
		}
	}
	c.sendPacket(suback) // SUBACK 3.9
}

/*
 * UNSUBSCRIBE – Unsubscribe from topics (3.10)
 */
func (c *client) processUnsubscribe(pkt *packet.UnsubscribePacket) {
	c.mutex.Lock()
	for _, t := range pkt.Topics {
		delete(c.subscriptions, t)
	}
	c.mutex.Unlock()
	p := packet.NewUnsubackPacket()
	p.PacketID = pkt.PacketID
	c.sendPacket(p) // UNSUBACK 3.11
}

/*
 * PINGREQ – PING request (3.12)
 */
func (c *client) processPing(pkt *packet.PingreqPacket) {
	p := packet.NewPingrespPacket()
	c.sendPacket(p) // PINGRESP 3.13
}

/*
 * DISCONNECT – Disconnect notification(3.14)
 */
func (c *client) processDisconnect(pkt *packet.DisconnectPacket) {
	//discard Will
	c.will = nil
	// Close connection if the client has not already done so
	c.conn.Close()
}

func (c *client) send(msg *packet.Message, qos byte, retain bool) {
	p := packet.NewPublishPacket()

	// Re-pack the messgae to use the reciever's QoS
	// and set retain to false
	m := &packet.Message{
		Topic:   msg.Topic,
		Payload: msg.Payload,
		QOS:     qos,
		Retain:  retain,
	}
	p.Message = *m
	// TODO set to true if this is a retry
	p.Dup = false
	// Sec. 2.3.1
	if qos > 0 {
		p.PacketID = c.newPacketID()
		c.outboundInTransit[p.PacketID] = p.Message
	}
	c.sendPacket(p)
}

func (c *client) resend(packetID uint16, msg *packet.Message) {
	log.Printf("Re-sending message %d", packetID)
	p := packet.NewPublishPacket()
	p.Message = *msg
	p.Dup = true
	p.PacketID = packetID
	c.sendPacket(p)
}

func (c *client) sendRetained(topic string, qos uint8) {

	// Retained messages [MQTT-3.3.1-6]
	for t, msg := range c.broker.retained {
		if matches(topic, t) {
			// Retain flag set to 1 [MQTT-3.3.1-8]
			c.send(msg, qos, true)
		}
	}
}

func (c *client) setReadDeadline() {
	if c.keepalive > 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.keepalive) * time.Second)); err != nil {
			log.Printf("Error setting read deadline on network connection", err)
		}
	}
}

func (c *client) sendPacket(p packet.Packet) {
	c.connectionMutex.Lock()
	c.encoder.Write(p)
	c.encoder.Flush()
	c.connectionMutex.Unlock()
}

func (c *client) newInternalClientID() string {
	c.internalClientCounter++
	return "internalClient" + string(c.internalClientCounter)

}

func (c *client) newPacketID() uint16 {
	c.packetIDCounter++
	return c.packetIDCounter
}
