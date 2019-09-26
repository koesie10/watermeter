package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/warthog618/gpio"
)

var inputPin InputPin = gpio.GPIO21

var pin *gpio.Pin

var jsonOutput bool

var rootCmd = &cobra.Command{
	Use: "watermeter",
}

func init() {
	rootCmd.PersistentFlags().Var(&inputPin, "input-pin", "input pin")

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
}

func OpenPin() (*gpio.Pin, error) {
	if err := gpio.Open(); err != nil {
		return nil, fmt.Errorf("failed to open GPIO: %w", err)
	}

	pin = gpio.NewPin(uint8(inputPin))
	pin.SetMode(gpio.Input)
	pin.SetPull(gpio.PullDown)

	return pin, nil
}

func WaitForExit() error {
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		if pin != nil {
			pin.Unwatch()
		}
		gpio.Close()
		close(cleanupDone)
	}()
	<-cleanupDone

	return nil
}

type InputPin uint8

func (p *InputPin) String() string {
	pin := *p

	for name, v := range gpioPins{
		if pin == v {
			return name
		}
	}

	for name, v := range jp8Pins {
		if pin == v {
			return name
		}
	}

	return strconv.Itoa(int(pin))
}

func (p *InputPin) Set(str string) error {
	if len(str) < 1 {
		return fmt.Errorf("invalid input pin: empty")
	}

	if pin, ok := gpioPins[str]; ok {
		*p = pin
		return nil
	}

	if pin, ok := jp8Pins[str]; ok {
		*p = pin
		return nil
	}

	pin, err := strconv.ParseInt(str, 10, 8)
	if err != nil {
		return fmt.Errorf("invalid input pin: %w", err)
	}
	*p = InputPin(pin)

	return nil
}

func (p *InputPin) Type() string {
	return "string"
}
