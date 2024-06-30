package test

import (
	"testing"

	"github.com/ebilling/gpio"
	"github.com/stretchr/testify/assert"
)

func TestGPIO(t *testing.T) {
	pin, err := gpio.NewOutput(22, false)
	assert.NoError(t, err)
	assert.Equal(t, pin.Number, uint(22))
	pin.Close()
}
