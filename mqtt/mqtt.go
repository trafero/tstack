package mqtt

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"time"
)

// A random number generator ready to make client-ids
var cliRand *rand.Rand

// ConnectionErrors is an array of errors corresponding to the
// Connect return codes specified in the specification.
var ConnectionErrors = [6]error{
	nil, // Connection Accepted (not an error)
	errors.New("Connection Refused: unacceptable protocol version"),
	errors.New("Connection Refused: identifier rejected"),
	errors.New("Connection Refused: server unavailable"),
	errors.New("Connection Refused: bad user name or password"),
	errors.New("Connection Refused: not authorized"),
}

func init() {
	var seed int64
	var sb [4]byte
	crand.Read(sb[:])
	seed = int64(time.Now().Nanosecond())<<32 |
		int64(sb[0])<<24 | int64(sb[1])<<16 |
		int64(sb[2])<<8 | int64(sb[3])
	cliRand = rand.New(rand.NewSource(seed))
}

// An intPayload implements proto.Payload, and is an int64 that
// formats itself and then prints itself into the payload.
type intPayload string

func newIntPayload(i int64) intPayload {
	return intPayload(fmt.Sprint(i))
}
func (ip intPayload) ReadPayload(r io.Reader) error {
	// not implemented
	return nil
}
func (ip intPayload) WritePayload(w io.Writer) error {
	_, err := w.Write([]byte(string(ip)))
	return err
}
func (ip intPayload) Size() int {
	return len(ip)
}
