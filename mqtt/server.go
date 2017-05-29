package mqtt

import (
	"github.com/trafero/tstack/auth"
	"log"
	"math/rand"
	"net"
	"time"
)

// A Server holds all the state associated with an MQTT server.
type Server struct {
	l             net.Listener
	subs          *subscriptions
	stats         *stats
	Done          chan struct{}
	StatsInterval time.Duration // Defaults to 10 seconds. Must be set using sync/atomic.StoreInt64().
	Dump          bool          // When true, dump the messages in and out.
	rand          *rand.Rand
	auth          auth.Auth // Authentication mechanism
}

// NewServer creates a new MQTT server, which accepts connections from
// the given listener. When the server is stopped (for instance by
// another goroutine closing the net.Listener), channel Done will become
// readable.
func NewServer(auth auth.Auth, l net.Listener, s *subscriptions) *Server {
	svr := &Server{
		l:             l,
		stats:         &stats{},
		Done:          make(chan struct{}),
		StatsInterval: time.Second * 10,
		subs:          s,
		auth:          auth,
	}

	// start the stats reporting goroutine
	go func() {
		for {
			svr.stats.publish(svr.subs, svr.StatsInterval)
			select {
			case <-svr.Done:
				return
			default:
				// keep going
			}
			time.Sleep(svr.StatsInterval)
		}
	}()

	return svr
}

// Start makes the Server start accepting and handling connections.
func (s *Server) Start() {
	go func() {
		for {
			conn, err := s.l.Accept()
			if err != nil {
				log.Print("Accept: ", err)
				break
			}

			cli := s.newIncomingConn(conn)
			s.stats.clientConnect()
			cli.start()
		}
		close(s.Done)
	}()
}

// newIncomingConn creates a new incomingConn associated with this
// server. The connection becomes the property of the incomingConn
// and should not be touched again by the caller until the Done
// channel becomes readable.
func (s *Server) newIncomingConn(conn net.Conn) *incomingConn {
	return &incomingConn{
		svr:  s,
		conn: conn,
		auth: s.auth,
		jobs: make(chan job, sendingQueueLength),
		Done: make(chan struct{}),
	}
}
