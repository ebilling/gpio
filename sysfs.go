package gpio

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type direction uint

const (
	inDirection direction = iota
	outDirection
)

type Edge uint

const (
	EdgeNone Edge = iota
	EdgeRising
	EdgeFalling
	EdgeBoth
)

type LogicLevel uint

const (
	ActiveHigh LogicLevel = iota
	ActiveLow
)

type Value uint

const (
	Inactive Value = 0
	Active   Value = 1
)

func exportGPIO(p Pin) error {
	export, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0600)
	if err == nil {
		defer export.Close()
		_, err = export.Write([]byte(strconv.Itoa(int(p.Number))))
	}
	return fmt.Errorf("/sys/class/gpio/export, %w", err)
}

func unexportGPIO(p Pin) error {
	export, err := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, 0600)
	if err == nil {
		defer export.Close()
		export.Write([]byte(strconv.Itoa(int(p.Number))))
	}
	return fmt.Errorf("/sys/class/gpio/unexport, %w", err)
}

func setDirection(p Pin, d direction, initialValue uint) error {
	dir, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", p.Number), os.O_WRONLY, 0600)
	if err == nil {
		defer dir.Close()

		switch {
		case d == inDirection:
			dir.Write([]byte("in"))
		case d == outDirection && initialValue == 0:
			dir.Write([]byte("low"))
		case d == outDirection && initialValue == 1:
			dir.Write([]byte("high"))
		default:
			panic(fmt.Sprintf("setDirection called with invalid direction or initialValue, %d, %d", d, initialValue))
		}
	}
	return fmt.Errorf("/sys/class/gpio/gpio%d/direction, %w", p.Number, err)
}

func setEdgeTrigger(p Pin, e Edge) error {
	edge, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/edge", p.Number), os.O_WRONLY, 0600)
	if err == nil {
		defer edge.Close()

		switch e {
		case EdgeNone:
			edge.Write([]byte("none"))
		case EdgeRising:
			edge.Write([]byte("rising"))
		case EdgeFalling:
			edge.Write([]byte("falling"))
		case EdgeBoth:
			edge.Write([]byte("both"))
		default:
			panic(fmt.Sprintf("setEdgeTrigger called with invalid edge %d", e))
		}
	}

	return fmt.Errorf("/sys/class/gpio/gpio%d/edge, %w", p.Number, err)
}

func setLogicLevel(p Pin, l LogicLevel) error {
	level, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/active_low", p.Number), os.O_WRONLY, 0600)
	if err != nil {
		defer level.Close()

		switch l {
		case ActiveHigh:
			level.Write([]byte("0"))
		case ActiveLow:
			level.Write([]byte("1"))
		default:
			return errors.New("Invalid logic level setting.")
		}
	}
	return fmt.Errorf("/sys/class/gpio/gpio%d/active_low, %w", p.Number, err)
}

func openPin(p Pin, write bool) (Pin, error) {
	flags := os.O_RDONLY
	if write {
		flags = os.O_RDWR
	}
	f, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", p.Number), flags, 0600)
	if err != nil {
		return p, fmt.Errorf("failed to open gpio %d value file for reading, %w", p.Number, err)
	}
	p.f = f
	return p, nil
}

func readPin(p Pin) (val uint, err error) {
	file := p.f
	file.Seek(0, 0)
	buf := make([]byte, 1)
	_, err = file.Read(buf)
	if err != nil {
		return 0, err
	}
	c := buf[0]
	switch c {
	case '0':
		return 0, nil
	case '1':
		return 1, nil
	default:
		return 0, fmt.Errorf("read inconsistent value in pinfile, %c", c)
	}
}

func writePin(p Pin, v uint) error {
	var buf []byte
	switch v {
	case 0:
		buf = []byte{'0'}
	case 1:
		buf = []byte{'1'}
	default:
		return fmt.Errorf("invalid output value %d", v)
	}
	_, err := p.f.Write(buf)
	return err
}
