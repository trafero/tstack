package matrix

import (
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/chip"
	"log"
	"strconv"
	"sync"
	"time"
)

type Key int

func (k Key) String() string {
	switch k {
	case KStar:
		return "*"
	case KHash:
		return "#"
	case KA:
		return "A"
	case KB:
		return "B"
	case KC:
		return "C"
	case KD:
		return "D"
	default:
		return strconv.Itoa(int(k) - 1)
	}
}

const (
	KNone Key = iota
	K0
	K1
	K2
	K3
	K4
	K5
	K6
	K7
	K8
	K9
	KA
	KB
	KC
	KD
	KStar
	KHash

	debounce = 20 * time.Millisecond

	pollDelay = 10

	rows = 4
	cols = 4
)

var keyMap [][]Key

func init() {
	keyMap = make([][]Key, rows)
	for i := 0; i < rows; i++ {
		keyMap[i] = make([]Key, cols)
	}
	keyMap[0][3] = KStar
	keyMap[0][2] = K0
	keyMap[0][1] = KHash
	keyMap[0][0] = KD

	keyMap[1][3] = K7
	keyMap[1][2] = K8
	keyMap[1][1] = K9
	keyMap[1][0] = KC

	keyMap[2][3] = K4
	keyMap[2][2] = K5
	keyMap[2][1] = K6
	keyMap[2][0] = KB

	keyMap[3][3] = K1
	keyMap[3][2] = K2
	keyMap[3][1] = K3
	keyMap[3][0] = KA
}

// A Matrix4x4 struct represents access to the keypad.
type Matrix4x4 struct {
	rowPins, colPins []embd.DigitalPin
	initialized      bool
	mu               sync.RWMutex

	poll int

	keyPressed chan Key
	quit       chan bool
}

// New creates a new interface for matrix4x4.
func New(rowPins, colPins []int) (*Matrix4x4, error) {
	log.Println("New matrix")

	m := &Matrix4x4{
		rowPins: make([]embd.DigitalPin, rows),
		colPins: make([]embd.DigitalPin, cols),
		poll:    pollDelay,
	}
	var err error
	for i := 0; i < rows; i++ {
		m.rowPins[i], err = embd.NewDigitalPin(rowPins[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < cols; i++ {
		m.colPins[i], err = embd.NewDigitalPin(colPins[i])
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		panic(err)
	}

	return m, nil
}

// SetPollDelay sets the delay between run of key scan acquisition loop.
func (d *Matrix4x4) SetPollDelay(delay int) {
	d.poll = delay
}

func (d *Matrix4x4) setup() error {

	for i := 0; i < rows; i++ {
		if err := d.rowPins[i].SetDirection(embd.Out); err != nil {
			return err
		}
		if err := d.rowPins[i].Write(embd.Low); err != nil {
			return err
		}
		//log.Printf("Setting row pin %d direction", d.rowPins[i])
		if err := d.rowPins[i].SetDirection(embd.In); err != nil {
			return err
		}
	}

	for i := 0; i < cols; i++ {
		// log.Printf("Setting col pin %d", d.colPins[i])
		if err := d.colPins[i].SetDirection(embd.Out); err != nil {
			return err
		}
		if err := d.colPins[i].Write(embd.Low); err != nil {
			return err
		}
	}

	d.initialized = true

	return nil
}

func (d *Matrix4x4) findPressedKey() (Key, error) {
	// log.Println("Finding key pressed")

	if err := d.setup(); err != nil {
		return 0, err
	}

	for col := 0; col < cols; col++ {
		if err := d.colPins[col].Write(embd.High); err != nil {
			return KNone, err
		}
		for row := 0; row < rows; row++ {
			value, err := d.rowPins[row].Read()
			if err != nil {
				return KNone, err
			}
			if value == embd.High {
				return keyMap[row][col], nil
			}
		}
	}
	return KNone, nil
}

// Pressed key returns the current key pressed on the keypad.
func (d *Matrix4x4) PressedKey() (key Key, err error) {
	select {
	case key = <-d.keyPressed:
		return
	default:
		return d.findPressedKey()
	}
}

// Close.
func (d *Matrix4x4) Close() {
	if d.quit != nil {
		d.quit <- true
	}
}
