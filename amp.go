package tpa2016

// https://www.adafruit.com/datasheets/TPA2016D2.pdf

import (
	"errors"

	"github.com/davecheney/i2c"
)

type AGCRatio byte

const (
	AGCOff AGCRatio = 0 // 1:1
	AGC2   AGCRatio = 1 // 2:1
	AGC4   AGCRatio = 2 // 4:1
	AGC8   AGCRatio = 3 // 8:1
)

const (
	setup          = 0x1
	setupREnabled  = 0x80
	setupLEnabled  = 0x40
	setupSWS       = 0x20
	setupRFault    = 0x10
	setupLFault    = 0x08
	setupThermal   = 0x04
	setupNoiseGate = 0x01

	regAtk      = 0x2
	regRel      = 0x3
	regHold     = 0x4
	regGain     = 0x5
	regAGCLimit = 0x6
	regAGC      = 0x7
	regAGCOff   = 0x00
	regAGC2     = 0x01
	regAGC4     = 0x02
	regAGC8     = 0x03

	i2cAddr = 0x58
)

type Amp struct {
	i2c *i2c.I2C
	buf [8]byte
}

func New(bus int) (*Amp, error) {
	iic, err := i2c.New(i2cAddr, bus)
	if err != nil {
		return nil, err
	}
	return &Amp{
		i2c: iic,
	}, nil
}

func (a *Amp) Close() error {
	return a.i2c.Close()
}

// Gain returns the fixed gain in dB
func (a *Amp) Gain() (int, error) {
	b, err := a.readByte(regGain)
	return int(b), err
}

// SetGain sets the fixed gain in dB
func (a *Amp) SetGain(g int) error {
	if g > 30 {
		g = 30
	} else if g < -28 {
		g = -28
	}
	return a.writeByte(regGain, byte(g)&0x3f)
}

// EnableChannel turns on/off the right and left channels
func (a *Amp) EnableChannel(right, left bool) error {
	sb, err := a.readByte(setup)
	if err != nil {
		return err
	}
	if right {
		sb |= setupREnabled
	} else {
		sb &^= setupREnabled
	}
	if left {
		sb |= setupLEnabled
	} else {
		sb &^= setupLEnabled
	}
	return a.writeByte(setup, sb)
}

func (a *Amp) Faults() (left, right, thermal bool, err error) {
	sb, err := a.readByte(setup)
	return sb&setupLFault != 0, sb&setupRFault != 0, sb&setupThermal != 0, err
}

func (a *Amp) SetAGCCompression(ratio AGCRatio) error {
	if ratio > 3 {
		return errors.New("tpa2016: agc ratio out of range")
	}
	b, err := a.readByte(regAGC)
	if err != nil {
		return err
	}
	b = (b &^ 3) | byte(ratio)
	return a.writeByte(regAGC, b)
}

// SetAGCMaxGain sets the maximum gain the AGC can achieve. Valid
// values are in the range [18, 30]
func (a *Amp) SetAGCMaxGain(g int) error {
	if g < 18 || g > 30 {
		return errors.New("tpa2016: agc max gain must be in the range [18, 30]")
	}
	b, err := a.readByte(regAGC)
	if err != nil {
		return err
	}
	b = (b &^ 0xf0) | (byte(g-18) << 4)
	return a.writeByte(regAGC, b)
}

func (a *Amp) readByte(register byte) (byte, error) {
	a.buf[0] = register
	if _, err := a.i2c.Write(a.buf[:1]); err != nil {
		return 0, err
	}
	_, err := a.i2c.Read(a.buf[:1])
	return a.buf[0], err
}

func (a *Amp) writeByte(register, value byte) error {
	a.buf[0] = register
	a.buf[1] = value
	_, err := a.i2c.Write(a.buf[:2])
	return err
}
