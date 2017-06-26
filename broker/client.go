package broker

import (
	"github.com/gomqtt/packet"
	"github.com/trafero/tstack/auth"
	"io"
	"log"
	"net"
	"sync"
	"unsafe"
)

type client struct {
	broker           *broker
	conn             net.Conn
	auth             auth.Auth
	processedConnect bool
	clientid         string
	username         string
	rights           string
	will             *packet.Message
	keepalive        uint16
	encoder          *packet.Encoder
	decoder          *packet.Decoder
	subscriptions    []packet.Subscription
	mutex            *sync.Mutex
	packetIDCounter  uint16

	// QOS 2 messages to be received (and passeed to broker)
	inboundInTransit map[uint16]packet.Message

	// QOS 2 messages to be sent
	outboundInTransit map[uint16]packet.Message
}

func NewClient(a auth.Auth, b *broker, c net.Conn) *client {
	return &client{
		auth:              a,
		broker:            b,
		conn:              c,
		processedConnect:  false,
		mutex:             &sync.Mutex{},
		inboundInTransit:  make(map[uint16]packet.Message),
		outboundInTransit: make(map[uint16]packet.Message),
		packetIDCounter:   0,
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
			log.Println("UnsubscribePacket not implemented")
			c.conn.Close()
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
		c.broker.receive(c.will)
	}

}

func (c *client) processPubrel(pkt *packet.PubrelPacket) {
	log.Println("Got Pubrel")
	msg := c.inboundInTransit[pkt.PacketID]

	if unsafe.Sizeof(msg) != 0 {
		// msg is not an empty stuct
		c.broker.receive(&msg)
		delete(c.inboundInTransit, pkt.PacketID)
		p := packet.NewPubcompPacket()
		p.PacketID = pkt.PacketID
		c.encoder.Write(p)
		c.encoder.Flush()
	}

}

func (c *client) processPing(pkt *packet.PingreqPacket) {
	log.Println("Got ping request")
	p := packet.NewPingrespPacket()
	c.encoder.Write(p)
	c.encoder.Flush()
}

func (c *client) processConnect(pkt *packet.ConnectPacket) {
	log.Printf("Got connect packet %v", pkt)
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
		log.Println("User could not be authenticated")
		c.conn.Close()
		return
	}
	if pkt.ClientID == "" {
		c.writeConnack(packet.ErrIdentifierRejected)
		log.Println("Blank client id")
		c.conn.Close()
		return
	}

	// TODO check Clinet ID is not already in use
	c.clientid = pkt.ClientID

	c.username = pkt.Username
	c.rights = c.auth.Rights(c.username)

	if pkt.Will != nil && !c.authorized(pkt.Will.Topic) {
		log.Println("Client not authorized to write this will")
	} else {
		c.will = pkt.Will // May be nil but that is ok
	}

	// TODO make use of keepalive
	c.keepalive = pkt.KeepAlive
	if pkt.CleanSession {
		// TODO clean session
	}

	c.broker.AddClient(c)
	c.writeConnack(packet.ConnectionAccepted)
}
func (c *client) writeConnack(code packet.ConnackCode) {
	connack := packet.NewConnackPacket()
	connack.SessionPresent = false
	connack.ReturnCode = code
	c.encoder.Write(connack)
	c.encoder.Flush()
}

func (c *client) processDisconnect(pkt *packet.DisconnectPacket) {
	//discard Will
	c.will = nil
	// Close connection if the client has not already done so
	c.conn.Close()
}

func (c *client) processPublish(pkt *packet.PublishPacket) {
	log.Printf("Got publish packet %v", pkt)
	if !c.authorized(pkt.Message.Topic) {
		// TODO send code back?
		log.Printf("Not authorized to publish to topic %s", pkt.Message.Topic)
		// Give them a hint
		c.conn.Close()
	} else {
		// QOS 1

		switch pkt.Message.QOS {

		case packet.QOSAtMostOnce:
			// QOS 0
			c.broker.receive(&pkt.Message)

		case packet.QOSAtLeastOnce:
			// QOS 1
			c.broker.receive(&pkt.Message)
			p := packet.NewPubackPacket()
			p.PacketID = pkt.PacketID
			c.encoder.Write(p)
			c.encoder.Flush()

		case packet.QOSExactlyOnce:
			// QOS 2
			c.mutex.Lock()
			c.inboundInTransit[pkt.PacketID] = pkt.Message
			c.mutex.Unlock()
			p := packet.NewPubrecPacket()
			p.PacketID = pkt.PacketID
			c.encoder.Write(p)
			c.encoder.Flush()
			// Send it back to the main switch for a Pubrel

		default:
			// Unknown QOS
			log.Printf("Unknown QOS level")
			c.conn.Close()

		}
	}
}

func (c *client) processSubscribe(pkt *packet.SubscribePacket) {

	log.Printf("Got subscribe packet %v", pkt)

	suback := packet.NewSubackPacket()
	suback.PacketID = pkt.PacketID

	for _, s := range pkt.Subscriptions {
		if !c.authorized(s.Topic) {
			log.Printf("Not authorized to subscribe to topic %s", s.Topic)
			suback.ReturnCodes = append(suback.ReturnCodes, 0x80) // sec 3.9.3 of spec

		} else {
			// Look for existing subscription and replace if we find it,
			found := false
			for i, existingSubs := range c.subscriptions {
				if s.Topic == existingSubs.Topic {
					found = true
					c.mutex.Lock()
					c.subscriptions[i] = s
					c.mutex.Unlock()
					break // can only be one match at most
				}
			}
			// Otherwise add the new subscription
			if found == false {
				c.mutex.Lock()
				c.subscriptions = append(c.subscriptions, s)
				c.mutex.Unlock()
			}
			suback.ReturnCodes = append(suback.ReturnCodes, s.QOS)
		}
	}
	c.encoder.Write(suback)
	c.encoder.Flush()
}

func (c *client) authorized(topic string) bool {
	return matches(c.rights, topic)
}

func (c *client) Send(msg *packet.Message, qos byte) {
	log.Println("Sending message")
	p := packet.NewPublishPacket()

	// Re-pack the messgae to use the reciever's QoS
	// and set retain to false
	m := &packet.Message{
		Topic:   msg.Topic,
		Payload: msg.Payload,
		QOS:     qos,
		Retain:  false,
	}

	p.Message = *m
	// TODO set to true if this is a retry
	p.Dup = false
	// Sec. 2.3.1
	if qos > 0 {
		p.PacketID = c.newPacketID()
	}
	c.encoder.Write(p)
	c.encoder.Flush()
	c.outboundInTransit[p.PacketID] = p.Message
}

// TODO build in resend if do not get a puback

// For QOS 1 (sec. 3.4)
func (c *client) processPuback(pkt *packet.PubackPacket) {
	log.Printf("Got Publish Ack for packet id %d", pkt.PacketID)
	delete(c.outboundInTransit, pkt.PacketID)
}

// For QOS 2
func (c *client) processPubrec(pkt *packet.PubrecPacket) {
	log.Printf("Got Publish Received for packet id %d", pkt.PacketID)

	// Only send resonse if we have the message
	/*
		if _,ok := c.outboundInTransit[pkt.PacketID]; ! ok {
			log.Println("Pubrec for a message that I do not have")
			c.conn.Close()
			return
		}
	*/
	p := packet.NewPubrelPacket()
	p.PacketID = pkt.PacketID
	c.encoder.Write(p)
	c.encoder.Flush()
}

// For QOS 2
func (c *client) processComp(pkt *packet.PubcompPacket) {
	log.Printf("Got Pubcomp for packet id %d", pkt.PacketID)
	delete(c.outboundInTransit, pkt.PacketID)
}

func (c *client) newPacketID() uint16 {
	c.packetIDCounter++
	return c.packetIDCounter
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
