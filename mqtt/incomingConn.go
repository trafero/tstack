package mqtt

import (
	"fmt"
	proto "github.com/huin/mqtt"
	"github.com/trafero/tstack/auth"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

// An IncomingConn represents a connection into a Server.
type incomingConn struct {
	svr      *Server
	conn     net.Conn
	jobs     chan job
	clientid string
	Done     chan struct{}
	username string
	auth     auth.Auth // struct for authentication functions
	rights   string    // User rights
}

var clients = make(map[string]*incomingConn)
var clientsMu sync.Mutex

const sendingQueueLength = 10000

type receipt chan struct{}

// Wait for the receipt to indicate that the job is done.
func (r receipt) wait() {
	// TODO: timeout
	<-r
}

type job struct {
	m proto.Message
	r receipt
}

// Start reading and writing on this connection.
func (c *incomingConn) start() {
	go c.reader()
	go c.writer()
}

// Add this	connection to the map, or find out that an existing connection
// already exists for the same client-id.
func (c *incomingConn) add() *incomingConn {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	existing, ok := clients[c.clientid]
	if !ok {
		// this client id already exists, return it
		return existing
	}

	clients[c.clientid] = c
	return nil
}

// Delete a connection; the conection must be closed by the caller first.
func (c *incomingConn) del() {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, c.clientid)
	return
}

// Replace any existing connection with this one. The one to be replaced,
// if any, must be closed first by the caller.
func (c *incomingConn) replace() {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// Check that any existing connection is already closed.
	existing, ok := clients[c.clientid]
	if ok {
		die := false
		select {
		case _, ok := <-existing.jobs:
			// what? we are expecting that this channel is closed!
			if ok {
				die = true
			}
		default:
			die = true
		}
		if die {
			panic("Attempting to replace a connection that is not closed")
		}

		delete(clients, c.clientid)
	}

	clients[c.clientid] = c
	return
}

// Queue a message; no notification of sending is done.
func (c *incomingConn) submit(m proto.Message) {
	j := job{m: m}
	select {
	case c.jobs <- j:
	default:
		log.Print(c, ": failed to submit message")
	}
	return
}

func (c *incomingConn) String() string {
	return fmt.Sprintf("{IncomingConn: %v}", c.clientid)
}

// Queue a message, returns a channel that will be readable
// when the message is sent.
func (c *incomingConn) submitSync(m proto.Message) receipt {
	j := job{m: m, r: make(receipt)}
	c.jobs <- j
	return j.r
}

func (c *incomingConn) reader() {
	// On exit, close the connection and arrange for the writer to exit
	// by closing the output channel.
	defer func() {
		c.conn.Close()
		c.svr.stats.clientDisconnect()
		close(c.jobs)
	}()

	for {
		// TODO: timeout (first message and/or keepalives)
		m, err := proto.DecodeOneMessage(c.conn, nil)
		if err != nil {
			if err == io.EOF {
				return
			}
			if strings.HasSuffix(err.Error(), "use of closed network connection") {
				return
			}
			log.Print("reader: ", err)
			return
		}
		c.svr.stats.messageRecv()

		if c.svr.Dump {
			log.Printf("dump  in: %T", m)
		}

		switch m := m.(type) {

		case *proto.Connect:
			c.connect(m)
		case *proto.Publish:
			c.publish(m)
		case *proto.PingReq:
			c.submit(&proto.PingResp{})
		case *proto.Subscribe:
			c.subscribe(m)
		case *proto.Unsubscribe:
			c.unsubscribe(m)
		case *proto.Disconnect:
			return

		default:
			log.Printf("reader: unknown msg type %T", m)
			return
		}
	}
}

func (c *incomingConn) connect(m *proto.Connect) {

	rc := proto.RetCodeAccepted

	if (m.ProtocolName != "MQIsdp" || m.ProtocolVersion != 3) &&
		(m.ProtocolName != "MQTT" || m.ProtocolVersion != 4) {
		log.Print("Reject connection from ", m.ProtocolName, " version ", m.ProtocolVersion)
		rc = proto.RetCodeUnacceptableProtocolVersion
	}

	// Set username from incomming connection. Used later for authorization
	c.username = m.Username

	// Authenticate user
	if !c.auth.Authenticate(m.Username, m.Password) {
		rc = proto.RetCodeBadUsernameOrPassword
		log.Printf("Connection refused for %v: %v", c.conn.RemoteAddr(), ConnectionErrors[rc])
	} else {
		// Check client id.
		if len(m.ClientId) < 1 || len(m.ClientId) > 23 {
			rc = proto.RetCodeIdentifierRejected
		}
		c.clientid = m.ClientId
	}

	// Get user acccess rights from authentication service
	c.rights = c.auth.Rights(m.Username)

	// Disconnect existing connections.
	if existing := c.add(); existing != nil {
		disconnect := &proto.Disconnect{}
		r := existing.submitSync(disconnect)
		r.wait()
		existing.del()
	}
	c.add()

	// TODO: Last will
	connack := &proto.ConnAck{
		ReturnCode: rc,
	}
	c.submit(connack)

	// close connection if it was a bad connect
	if rc != proto.RetCodeAccepted {
		log.Printf("Connection refused for %v: %v", c.conn.RemoteAddr(), ConnectionErrors[rc])
		return
	}

	// Log in mosquitto format.
	clean := 0
	if m.CleanSession {
		clean = 1
	}
	log.Printf("Client connected from %v as %v (c%v, k%v).", c.conn.RemoteAddr(), c.clientid, clean, m.KeepAliveTimer)

}

func (c *incomingConn) publish(m *proto.Publish) {

	// TODO: Proper QoS support
	if m.Header.QosLevel != proto.QosAtMostOnce {
		log.Printf("No support for QoS %v yet", m.Header.QosLevel)
		return
	}
	if m.Header.QosLevel != proto.QosAtMostOnce && m.MessageId == 0 {
		// Invalid message ID. See MQTT-2.3.1-1.
		log.Printf("Invalid MessageId in PUBLISH.")
		return
	}

	// Validate topic against compiled rights regex
	if !validate(m.TopicName, c.rights) {
		log.Printf("User %s is unauthorized to write to topic %s", c.username, m.TopicName)
		return
	}

	if isWildcard(m.TopicName) {
		log.Print("Ignoring PUBLISH with wildcard topic ", m.TopicName)
	} else {
		// log.Printf("Message to topic %s", m.TopicName)
		c.svr.subs.submit(c, m)
	}

	// PUBACK Packet fro QoS 0 not expected on publish
	if m.Header.QosLevel != proto.QosAtMostOnce {
		c.submit(&proto.PubAck{MessageId: m.MessageId})
	}
}

func (c *incomingConn) subscribe(m *proto.Subscribe) {

	if m.Header.QosLevel != proto.QosAtLeastOnce {
		// protocol error, disconnect
		return
	}
	if m.MessageId == 0 {
		// Invalid message ID. See MQTT-2.3.1-1.
		log.Printf("Invalid MessageId in SUBSCRIBE.")
		return
	}

	suback := &proto.SubAck{
		MessageId: m.MessageId,
		TopicsQos: make([]proto.QosLevel, len(m.Topics)),
	}
	for i, tq := range m.Topics {
		if !validate(tq.Topic, c.rights) {
			log.Printf("User %s is unauthorized to subscribe to topic %s", c.username, tq.Topic)
		} else {
			// TODO: Handle varying QoS correctly
			c.svr.subs.add(tq.Topic, c)
			suback.TopicsQos[i] = proto.QosAtMostOnce
		}
	}
	c.submit(suback)

	// Process retained messages.
	for _, tq := range m.Topics {
		// Check each topic matches rights regex
		if !validate(tq.Topic, c.rights) {
			log.Printf("User %s is unauthorized to subscribe to topic %s", c.username, tq.Topic)
		} else {
			// Subscribe to topic
			c.svr.subs.sendRetain(tq.Topic, c)
		}
	}
}
func (c *incomingConn) unsubscribe(m *proto.Unsubscribe) {
	if m.Header.QosLevel != proto.QosAtMostOnce && m.MessageId == 0 {
		// Invalid message ID. See MQTT-2.3.1-1.
		log.Printf("Invalid MessageId in UNSUBSCRIBE.")
		return
	}
	for _, t := range m.Topics {
		c.svr.subs.unsub(t, c)
	}
	ack := &proto.UnsubAck{MessageId: m.MessageId}
	c.submit(ack)
}

func (c *incomingConn) writer() {

	// Close connection on exit in order to cause reader to exit.
	defer func() {
		c.conn.Close()
		c.del()
		c.svr.subs.unsubAll(c)
	}()

	for job := range c.jobs {
		if c.svr.Dump {
			log.Printf("dump out: %T", job.m)
		}

		// TODO: write timeout
		err := job.m.Encode(c.conn)
		if job.r != nil {
			// notifiy the sender that this message is sent
			close(job.r)
		}
		if err != nil {
			// This one is not interesting; it happens when clients
			// disappear before we send their acks.
			oe, isoe := err.(*net.OpError)
			if isoe && oe.Err.Error() == "use of closed network connection" {
				return
			}
			// In Go < 1.5, the error is not an OpError.
			if err.Error() == "use of closed network connection" {
				return
			}

			log.Print("writer: ", err)
			return
		}
		c.svr.stats.messageSend()

		if _, ok := job.m.(*proto.Disconnect); ok {
			log.Print("writer: sent disconnect message")
			return
		}
	}
}

// header is used to initialize a proto.Header when the zero value
// is not correct. The zero value of proto.Header is
// the equivalent of header(dupFalse, proto.QosAtMostOnce, retainFalse)
// and is correct for most messages.
func header(d dupFlag, q proto.QosLevel, r retainFlag) proto.Header {
	return proto.Header{
		DupFlag: bool(d), QosLevel: q, Retain: bool(r),
	}
}

type retainFlag bool
type dupFlag bool

const (
	retainFalse retainFlag = false
	retainTrue             = true
	dupFalse    dupFlag    = false
	dupTrue                = true
)

func isWildcard(topic string) bool {
	if strings.Contains(topic, "#") || strings.Contains(topic, "+") {
		return true
	}
	return false
}

func (w wild) valid() bool {
	for i, part := range w.wild {
		// catch things like finance#
		if isWildcard(part) && len(part) != 1 {
			return false
		}
		// # can only occur as the last part
		if part == "#" && i != len(w.wild)-1 {
			return false
		}
	}
	return true
}

func validate(topic string, rights string) bool {

	// log.Printf("Validating topic %s againsts rights %s", topic, rights)

	topiclevels := strings.Split(topic, "/")
	rightslevels := strings.Split(rights, "/")
	for i := 0; i < len(topiclevels); i++ {

		// Rights levels are not deep enough
		if len(rightslevels) < i {
			return false
		}
		// Wildcard here on in, so match everything
		if rightslevels[i] == "#" {
			return true
		}
		// Topics do not match, and not a topic level wildcard
		if rightslevels[i] != "+" && rightslevels[i] != topiclevels[i] {
			return false
		}
	}
	// Matched all of topic
	return true
}
