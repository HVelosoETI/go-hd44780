package hd44780

// Hitachi HD44780U support library

import (
	"github.com/stianeikeland/go-rpio"
	"strings"
	"sync"
	"time"
)

type GPIO4bit struct {
	sync.Mutex

	RSPin int // GOIO 7   --> Raspberry Physical Pin 26
	EPin int // GPIO 8    --> Raspberry Physical Pin 24
	D4Pin int // GPIO 25  --> Raspberry Physical Pin 22
	D5Pin int // GPIO 24  --> Raspberry Physical Pin 18
	D6Pin int // GPIO 23  --> Raspberry Physical Pin 16
	D7Pin int // GPIO 18  --> Raspberry Physical Pin 12

	// max lines
	Lines int
	// Memory address for each line
	LinesAddr []byte
	// LCD width (number of character in line)
	Width int

	lcdRS rpio.Pin
	lcdE  rpio.Pin
	lcdD4 rpio.Pin
	lcdD5 rpio.Pin
	lcdD6 rpio.Pin
	lcdD7 rpio.Pin

	lastLines []string
	active    bool
}

// NewGPIO4bit create new GPIO4bit structure with some defaults
func NewGPIO4bit() (h *GPIO4bit) {
	h = &GPIO4bit{
		RSPin:     RSPin,
		EPin:      EPin,
		D4Pin:     D4Pin,
		D5Pin:     D5Pin,
		D6Pin:     D6Pin,
		D7Pin:     D7Pin,
		Lines:     Lines,
		LinesAddr: []byte{lcdLine1, lcdLine2, lcdLine3, lcdLine4},
		Width:     Width,
	}
	return
}

// Open / initialize LCD interface
func (h *GPIO4bit) Open() (err error) {
	h.Lock()
	defer h.Unlock()

	if h.active {
		return
	}

	if err := rpio.Open(); err != nil {
		return err
	}

	h.lcdRS = initPin(h.RSPin)
	h.lcdE = initPin(h.EPin)
	h.lcdD4 = initPin(h.D4Pin)
	h.lcdD5 = initPin(h.D5Pin)
	h.lcdD6 = initPin(h.D6Pin)
	h.lcdD7 = initPin(h.D7Pin)
	h.lastLines = make([]string, h.Lines, h.Lines)
	h.reset()
	h.active = true

	return
}

// Active return true when interface is working ok
func (h *GPIO4bit) Active() bool {
	return h.active
}

// Reset interface
func (h *GPIO4bit) Reset() {
	h.Lock()
	defer h.Unlock()
	h.reset()
}

func (h *GPIO4bit) reset() {
	// initialize
	h.write4Bits(0x3, lcdCmd)
	time.Sleep(5 * time.Millisecond)
	h.write4Bits(0x3, lcdCmd)
	time.Sleep(240 * time.Microsecond)
	h.write4Bits(0x3, lcdCmd)
	time.Sleep(240 * time.Microsecond)

	h.write4Bits(0x2, lcdCmd)
	time.Sleep(240 * time.Microsecond)

	h.writeByte(0x28, lcdCmd) // Data length, number of lines, font size
	h.writeByte(0x0C, lcdCmd) // Display On,Cursor Off, Blink Off
	h.writeByte(0x06, lcdCmd) // Cursor move direction

	h.writeByte(0x01, lcdCmd) // Clear display
	time.Sleep(5 * time.Millisecond)
}

// Clear display
func (h *GPIO4bit) Clear() {
	h.Lock()
	defer h.Unlock()

	if !h.active {
		return
	}

	h.writeByte(lcdLine1, lcdCmd)
	for i := 0; i < Width; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine2, lcdCmd)
	for i := 0; i < Width; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine3, lcdCmd)
	for i := 0; i < Width; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine4, lcdCmd)
	for i := 0; i < Width; i++ {
		h.writeByte(' ', lcdChr)
	}
}

// Close interface, clear display.
func (h *GPIO4bit) Close() {
	h.Lock()
	defer h.Unlock()

	if !h.active {
		return
	}

	h.writeByte(lcdLine1, lcdCmd)
	for i := 0; i < Width; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine2, lcdCmd)
	for i := 0; i < Width; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine3, lcdCmd)
	for i := 0; i < Width; i++ {
		h.writeByte(' ', lcdChr)
	}
	h.writeByte(lcdLine4, lcdCmd)
	for i := 0; i < Width; i++ {
		h.writeByte(' ', lcdChr)
	}

	h.writeByte(0x01, lcdCmd) // 000001 Clear display
	time.Sleep(5 * time.Millisecond)

	h.writeByte(0x0C, lcdCmd) // 001000 Display Off

	h.lcdRS.Low()
	h.lcdE.Low()
	h.lcdD4.Low()
	h.lcdD5.Low()
	h.lcdD6.Low()
	h.lcdD7.Low()
	rpio.Close()

	h.active = false
}

// writeByte send byte to lcd
func (h *GPIO4bit) writeByte(bits byte, characterMode bool) {
	if characterMode {
		h.lcdRS.High()
	} else {
		h.lcdRS.Low()
	}

	// High bits
	h.push4bits(bits >> 4)

	// Low bits
	h.push4bits(bits)

	time.Sleep(eDelay)
}

// write4Bits send (lower) 4bits  to lcd
func (h *GPIO4bit) write4Bits(bits byte, characterMode bool) {
	if characterMode {
		h.lcdRS.High()
	} else {
		h.lcdRS.Low()
	}

	h.push4bits(bits)

	time.Sleep(eDelay)
}

// push4bits push 4 bites on data lines
func (h *GPIO4bit) push4bits(bits byte) {
	if bits&0x01 == 0x01 {
		h.lcdD4.High()
	} else {
		h.lcdD4.Low()
	}
	if bits&0x02 == 0x02 {
		h.lcdD5.High()
	} else {
		h.lcdD5.Low()
	}
	if bits&0x04 == 0x04 {
		h.lcdD6.High()
	} else {
		h.lcdD6.Low()
	}
	if bits&0x08 == 0x08 {
		h.lcdD7.High()
	} else {
		h.lcdD7.Low()
	}
	// Toggle 'Enable' pin
	time.Sleep(ePulse)
	h.lcdE.High()
	time.Sleep(ePulse)
	h.lcdE.Low()
	time.Sleep(ePulse)
}

// DisplayLines sends one or more lines separated by \n to lcd
func (h *GPIO4bit) DisplayLines(msg string) {
	for line, text := range strings.Split(msg, "\n") {
		h.Display(line, text)
	}
}

// Display only one line
func (h *GPIO4bit) Display(line int, text string) {
	h.Lock()
	defer h.Unlock()

	if !h.active {
		return
	}

	if line >= h.Lines {
		return
	}

	if len(text) < Width {
		text = text + strings.Repeat(" ", h.Width-len(text))
	} else {
		text = text[:h.Width]
	}

	// skip not changed lines
	if h.lastLines[line] == text {
		return
	}

	h.lastLines[line] = text

	h.writeByte(h.LinesAddr[line], lcdCmd)

	for c := 0; c < h.Width; c++ {
		h.writeByte(byte(text[c]), lcdChr)
	}
}

func (h *GPIO4bit) SetChar(pos byte, def []byte) {
	if len(def) != 8 {
		panic("invalid def - req 8 bytes")
	}
	h.writeByte(0x40+pos*8, lcdCmd)
	for _, d := range def {
		h.writeByte(d, lcdChr)
	}
}

func (h *GPIO4bit) ToggleBacklight() {
}
